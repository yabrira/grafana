package connectors

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/grafana/grafana/pkg/apimachinery/errutil"
	"github.com/grafana/grafana/pkg/login/social"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/services/ssosettings"
	ssoModels "github.com/grafana/grafana/pkg/services/ssosettings/models"
	"github.com/grafana/grafana/pkg/services/ssosettings/ssosettingstests"
	"github.com/grafana/grafana/pkg/services/user"
	"github.com/grafana/grafana/pkg/setting"
)

func testAppleKeyPEM(t *testing.T) ([]byte, *ecdsa.PrivateKey) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), key
}

func testAppleProvider(t *testing.T, info *social.OAuthInfo) *SocialApple {
	t.Helper()
	if info == nil {
		info = &social.OAuthInfo{}
	}
	cfg := setting.NewCfg()
	cfg.AppURL = "https://grafana.example.com/"
	return NewAppleProvider(info, cfg, nil, ssosettingstests.NewFakeService(), featuremgmt.WithFeatures(), nil)
}

func unsignedJWT(t *testing.T, claims map[string]any) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	require.NoError(t, err)
	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".sig"
}

func TestGenerateAppleClientSecret(t *testing.T) {
	pemBytes, key := testAppleKeyPEM(t)

	t.Run("signs an ES256 JWT with Apple claims", func(t *testing.T) {
		secret, err := generateAppleClientSecret("TEAMID", "KEYID", "com.example.grafana", string(pemBytes), "")
		require.NoError(t, err)

		tok, err := jwt.ParseSigned(secret, []jose.SignatureAlgorithm{jose.ES256})
		require.NoError(t, err)
		require.Equal(t, "KEYID", tok.Headers[0].KeyID)

		var claims jwt.Claims
		require.NoError(t, tok.Claims(key.Public(), &claims))
		require.Equal(t, "TEAMID", claims.Issuer)
		require.Equal(t, "com.example.grafana", claims.Subject)
		require.Equal(t, jwt.Audience{appleAudience}, claims.Audience)
		require.WithinDuration(t, time.Now().Add(appleClientSecretTTL), claims.Expiry.Time(), time.Minute)
	})

	t.Run("reads the key from private_key_path", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "AuthKey.p8")
		require.NoError(t, os.WriteFile(path, pemBytes, 0o600))

		secret, err := generateAppleClientSecret("TEAMID", "KEYID", "com.example.grafana", "", path)
		require.NoError(t, err)
		require.NotEmpty(t, secret)
	})

	t.Run("fails when the key is missing", func(t *testing.T) {
		_, err := generateAppleClientSecret("TEAMID", "KEYID", "com.example.grafana", "", "")
		require.Error(t, err)
	})

	t.Run("fails when the PEM is invalid", func(t *testing.T) {
		_, err := generateAppleClientSecret("TEAMID", "KEYID", "com.example.grafana", "not-a-key", "")
		require.Error(t, err)
	})
}

func TestSocialApple_AuthCodeURL(t *testing.T) {
	s := testAppleProvider(t, &social.OAuthInfo{
		ClientId: "com.example.grafana",
		Scopes:   []string{"name", "email"},
	})

	authURL := s.AuthCodeURL("state-value")
	require.Contains(t, authURL, "response_mode=form_post")
	require.Contains(t, authURL, "client_id=com.example.grafana")
	require.Contains(t, authURL, appleAuthURL)
}

func TestSocialApple_ExchangeSendsClientSecretJWT(t *testing.T) {
	pemBytes, key := testAppleKeyPEM(t)

	var gotSecret string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		gotSecret = r.FormValue("client_secret")
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"access_token":"tok","token_type":"Bearer","expires_in":3600}`)
	}))
	t.Cleanup(srv.Close)

	s := testAppleProvider(t, &social.OAuthInfo{
		ClientId: "com.example.grafana",
		TokenUrl: srv.URL,
		Extra: map[string]string{
			teamIDKey:     "TEAMID",
			keyIDKey:      "KEYID",
			privateKeyKey: string(pemBytes),
		},
	})

	token, err := s.Exchange(context.Background(), "auth-code")
	require.NoError(t, err)
	require.Equal(t, "tok", token.AccessToken)
	require.NotEmpty(t, gotSecret)

	parsed, err := jwt.ParseSigned(gotSecret, []jose.SignatureAlgorithm{jose.ES256})
	require.NoError(t, err)
	var claims jwt.Claims
	require.NoError(t, parsed.Claims(key.Public(), &claims))
	require.Equal(t, "TEAMID", claims.Issuer)
}

func TestSocialApple_UserInfo(t *testing.T) {
	s := testAppleProvider(t, &social.OAuthInfo{SkipOrgRoleSync: true})

	t.Run("reads identity from the ID token", func(t *testing.T) {
		idToken := unsignedJWT(t, map[string]any{
			"sub":            "apple-sub",
			"email":          "user@privaterelay.appleid.com",
			"email_verified": "true",
		})
		token := (&oauth2.Token{}).WithExtra(map[string]any{"id_token": idToken})

		info, err := s.UserInfo(context.Background(), http.DefaultClient, token)
		require.NoError(t, err)
		require.Equal(t, "apple-sub", info.Id)
		require.Equal(t, "user@privaterelay.appleid.com", info.Email)
		require.Equal(t, "user@privaterelay.appleid.com", info.Login)
	})

	t.Run("uses the first-login form user for name and missing email", func(t *testing.T) {
		idToken := unsignedJWT(t, map[string]any{"sub": "apple-sub"})
		token := (&oauth2.Token{}).WithExtra(map[string]any{"id_token": idToken})
		ctx := ContextWithAppleFormUser(context.Background(), `{"name":{"firstName":"Jane","lastName":"Doe"},"email":"jane@example.com"}`)

		info, err := s.UserInfo(ctx, http.DefaultClient, token)
		require.NoError(t, err)
		require.Equal(t, "Jane Doe", info.Name)
		require.Equal(t, "jane@example.com", info.Email)
	})

	t.Run("rejects an unverified email", func(t *testing.T) {
		idToken := unsignedJWT(t, map[string]any{
			"sub":            "apple-sub",
			"email":          "user@example.com",
			"email_verified": false,
		})
		token := (&oauth2.Token{}).WithExtra(map[string]any{"id_token": idToken})

		_, err := s.UserInfo(context.Background(), http.DefaultClient, token)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not verified")
	})

	t.Run("requires an email", func(t *testing.T) {
		idToken := unsignedJWT(t, map[string]any{"sub": "apple-sub"})
		token := (&oauth2.Token{}).WithExtra(map[string]any{"id_token": idToken})

		_, err := s.UserInfo(context.Background(), http.DefaultClient, token)
		require.ErrorIs(t, err, ErrEmailNotFound)
	})

	t.Run("requires an ID token", func(t *testing.T) {
		_, err := s.UserInfo(context.Background(), http.DefaultClient, &oauth2.Token{})
		require.ErrorIs(t, err, ErrIDTokenNotFound)
	})
}

func TestSocialApple_Validate(t *testing.T) {
	pemBytes, _ := testAppleKeyPEM(t)
	s := testAppleProvider(t, &social.OAuthInfo{})

	testCases := []struct {
		name     string
		settings ssoModels.SSOSettings
		wantErr  error
	}{
		{
			name: "SSOSettings is valid",
			settings: ssoModels.SSOSettings{
				Settings: map[string]any{
					"client_id":   "com.example.grafana",
					"team_id":     "TEAMID",
					"key_id":      "KEYID",
					"private_key": string(pemBytes),
				},
			},
		},
		{
			name: "fails if client id is empty",
			settings: ssoModels.SSOSettings{
				Settings: map[string]any{
					"client_id":   "",
					"team_id":     "TEAMID",
					"key_id":      "KEYID",
					"private_key": string(pemBytes),
				},
			},
			wantErr: ssosettings.ErrBaseInvalidOAuthConfig,
		},
		{
			name: "fails if team id is empty",
			settings: ssoModels.SSOSettings{
				Settings: map[string]any{
					"client_id":   "com.example.grafana",
					"team_id":     "",
					"key_id":      "KEYID",
					"private_key": string(pemBytes),
				},
			},
			wantErr: ssosettings.ErrBaseInvalidOAuthConfig,
		},
		{
			name: "fails if key id is empty",
			settings: ssoModels.SSOSettings{
				Settings: map[string]any{
					"client_id":   "com.example.grafana",
					"team_id":     "TEAMID",
					"key_id":      "",
					"private_key": string(pemBytes),
				},
			},
			wantErr: ssosettings.ErrBaseInvalidOAuthConfig,
		},
		{
			name: "fails if private key and path are empty",
			settings: ssoModels.SSOSettings{
				Settings: map[string]any{
					"client_id": "com.example.grafana",
					"team_id":   "TEAMID",
					"key_id":    "KEYID",
				},
			},
			wantErr: ssosettings.ErrBaseInvalidOAuthConfig,
		},
		{
			name: "succeeds if private_key_path is set",
			settings: ssoModels.SSOSettings{
				Settings: map[string]any{
					"client_id":        "com.example.grafana",
					"team_id":          "TEAMID",
					"key_id":           "KEYID",
					"private_key_path": "/tmp/AuthKey.p8",
				},
			},
		},
		{
			name: "fails if auth url is not empty",
			settings: ssoModels.SSOSettings{
				Settings: map[string]any{
					"client_id":   "com.example.grafana",
					"team_id":     "TEAMID",
					"key_id":      "KEYID",
					"private_key": string(pemBytes),
					"auth_url":    "https://example.com/auth",
				},
			},
			wantErr: ssosettings.ErrBaseInvalidOAuthConfig,
		},
		{
			name: "fails if settings map contains an invalid field",
			settings: ssoModels.SSOSettings{
				Settings: map[string]any{
					"client_id":     "com.example.grafana",
					"invalid_field": []int{1, 2, 3},
				},
			},
			wantErr: ssosettings.ErrInvalidSettings,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := s.Validate(context.Background(), tc.settings, ssoModels.SSOSettings{}, &user.SignedInUser{IsGrafanaAdmin: true})
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			if err != nil {
				var e errutil.Error
				require.True(t, errors.As(err, &e))
				require.NoError(t, e, "expected no error, got %v", e.PublicMessage)
			}
		})
	}
}

func TestSocialApple_Reload(t *testing.T) {
	pemBytes, _ := testAppleKeyPEM(t)
	s := testAppleProvider(t, &social.OAuthInfo{ClientId: "old-id"})

	err := s.Reload(context.Background(), ssoModels.SSOSettings{
		Settings: map[string]any{
			"client_id":   "com.example.grafana",
			"team_id":     "TEAMID",
			"key_id":      "KEYID",
			"private_key": string(pemBytes),
			"enabled":     true,
		},
	})
	require.NoError(t, err)

	info := s.GetOAuthInfo()
	require.Equal(t, "com.example.grafana", info.ClientId)
	require.Equal(t, "TEAMID", info.Extra[teamIDKey])
	require.Equal(t, appleAuthURL, info.AuthUrl)
	require.Equal(t, appleTokenURL, info.TokenUrl)
	require.Equal(t, "inparams", info.AuthStyle)
}

func TestParseAppleUserForm(t *testing.T) {
	require.Equal(t, "Jane Doe", parseAppleUserForm(`{"name":{"firstName":"Jane","lastName":"Doe"}}`).displayName())
	require.Equal(t, "", parseAppleUserForm("").displayName())
	require.Equal(t, "", parseAppleUserForm("not-json").displayName())
}
