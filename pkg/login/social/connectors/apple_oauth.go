package connectors

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"golang.org/x/oauth2"

	"github.com/grafana/grafana/pkg/apimachinery/identity"
	"github.com/grafana/grafana/pkg/infra/remotecache"
	"github.com/grafana/grafana/pkg/login/social"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/services/ssosettings"
	ssoModels "github.com/grafana/grafana/pkg/services/ssosettings/models"
	"github.com/grafana/grafana/pkg/services/ssosettings/validation"
	"github.com/grafana/grafana/pkg/setting"
)

const (
	appleAuthURL  = "https://appleid.apple.com/auth/authorize"
	appleTokenURL = "https://appleid.apple.com/auth/token"
	appleJWKSURL  = "https://appleid.apple.com/auth/keys"
	appleAudience = "https://appleid.apple.com"

	teamIDKey         = "team_id"
	keyIDKey          = "key_id"
	privateKeyKey     = "private_key"
	privateKeyPathKey = "private_key_path"

	// Apple allows client-secret JWTs up to 6 months. A short TTL is enough
	// because Grafana mints a new one on every token exchange and refresh.
	appleClientSecretTTL = 5 * time.Minute
)

var ExtraAppleSettingKeys = map[string]ExtraKeyInfo{
	teamIDKey:         {Type: String},
	keyIDKey:          {Type: String},
	privateKeyKey:     {Type: String},
	privateKeyPathKey: {Type: String},
}

var _ social.SocialConnector = (*SocialApple)(nil)
var _ ssosettings.Reloadable = (*SocialApple)(nil)

type appleFormUserCtxKey struct{}

// ContextWithAppleFormUser stores Apple's first-login `user` form field so UserInfo can read name/email.
func ContextWithAppleFormUser(ctx context.Context, raw string) context.Context {
	if raw == "" {
		return ctx
	}
	return context.WithValue(ctx, appleFormUserCtxKey{}, raw)
}

func appleFormUserFromContext(ctx context.Context) string {
	v, _ := ctx.Value(appleFormUserCtxKey{}).(string)
	return v
}

type SocialApple struct {
	*SocialBase
}

type appleIDTokenClaims struct {
	Sub           string          `json:"sub"`
	Email         string          `json:"email"`
	EmailVerified json.RawMessage `json:"email_verified"`
	Name          string          `json:"name"`
	rawJSON       []byte
}

type appleUserForm struct {
	Name struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"name"`
	Email string `json:"email"`
}

type appleTokenSource struct {
	ctx            context.Context
	conf           *oauth2.Config
	token          *oauth2.Token
	teamID         string
	keyID          string
	clientID       string
	privateKey     string
	privateKeyPath string
}

func applyAppleEndpointDefaults(info *social.OAuthInfo) {
	if info.AuthStyle == "" {
		info.AuthStyle = "inparams"
	}
	if info.AuthUrl == "" {
		info.AuthUrl = appleAuthURL
	}
	if info.TokenUrl == "" {
		info.TokenUrl = appleTokenURL
	}
	if info.JwkSetURL == "" {
		info.JwkSetURL = appleJWKSURL
	}
}

func NewAppleProvider(info *social.OAuthInfo, cfg *setting.Cfg, orgRoleMapper *OrgRoleMapper, ssoSettings ssosettings.Service, features featuremgmt.FeatureToggles, cache remotecache.CacheStorage) *SocialApple {
	applyAppleEndpointDefaults(info)

	provider := &SocialApple{
		SocialBase: newSocialBaseWithCache(social.AppleProviderName, orgRoleMapper, info, features, cfg, cache),
	}

	ssoSettings.RegisterReloadable(social.AppleProviderName, provider)

	return provider
}

func (s *SocialApple) Validate(ctx context.Context, newSettings ssoModels.SSOSettings, oldSettings ssoModels.SSOSettings, requester identity.Requester) error {
	info, err := CreateOAuthInfoFromKeyValues(newSettings.Settings)
	if err != nil {
		return ssosettings.ErrInvalidSettings.Errorf("SSO settings map cannot be converted to OAuthInfo: %v", err)
	}
	oldInfo, err := CreateOAuthInfoFromKeyValues(oldSettings.Settings)
	if err != nil {
		oldInfo = &social.OAuthInfo{}
	}

	err = validateInfo(info, oldInfo, requester)
	if err != nil {
		return err
	}

	return validation.Validate(info, requester,
		validation.MustBeEmptyValidator(info.AuthUrl, "Auth URL"),
		validation.MustBeEmptyValidator(info.TokenUrl, "Token URL"),
		validation.MustBeEmptyValidator(info.ApiUrl, "API URL"),
		validation.RequiredValidator(info.Extra[teamIDKey], "Team Id"),
		validation.RequiredValidator(info.Extra[keyIDKey], "Key Id"),
		applePrivateKeyValidator,
		validation.ValidateIDTokenValidator)
}

func applePrivateKeyValidator(info *social.OAuthInfo, requester identity.Requester) error {
	if strings.TrimSpace(info.Extra[privateKeyKey]) == "" && strings.TrimSpace(info.Extra[privateKeyPathKey]) == "" {
		return ssosettings.ErrInvalidOAuthConfig("Private Key or Private Key Path is required.")
	}
	return nil
}

func (s *SocialApple) Reload(ctx context.Context, settings ssoModels.SSOSettings) error {
	newInfo, err := CreateOAuthInfoFromKeyValuesWithLogging(s.log, social.AppleProviderName, settings.Settings)
	if err != nil {
		return ssosettings.ErrInvalidSettings.Errorf("SSO settings map cannot be converted to OAuthInfo: %v", err)
	}

	applyAppleEndpointDefaults(newInfo)

	s.reloadMutex.Lock()
	defer s.reloadMutex.Unlock()

	s.updateInfo(ctx, social.AppleProviderName, newInfo)
	return nil
}

func (s *SocialApple) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	s.reloadMutex.RLock()
	defer s.reloadMutex.RUnlock()

	// Apple requires form_post when requesting name or email; Grafana's callback is GET-only without this.
	opts = append(opts, oauth2.SetAuthURLParam("response_mode", "form_post"))
	return s.getAuthCodeURL(state, opts...)
}

func (s *SocialApple) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	s.reloadMutex.RLock()
	defer s.reloadMutex.RUnlock()

	conf, err := s.oauthConfigWithClientSecret()
	if err != nil {
		return nil, err
	}
	return conf.Exchange(ctx, code, opts...)
}

func (s *SocialApple) TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	s.reloadMutex.RLock()
	defer s.reloadMutex.RUnlock()

	conf := *s.Config
	return &appleTokenSource{
		ctx:            ctx,
		conf:           &conf,
		token:          t,
		teamID:         s.info.Extra[teamIDKey],
		keyID:          s.info.Extra[keyIDKey],
		clientID:       s.info.ClientId,
		privateKey:     s.info.Extra[privateKeyKey],
		privateKeyPath: s.info.Extra[privateKeyPathKey],
	}
}

func (ts *appleTokenSource) Token() (*oauth2.Token, error) {
	secret, err := generateAppleClientSecret(ts.teamID, ts.keyID, ts.clientID, ts.privateKey, ts.privateKeyPath)
	if err != nil {
		return nil, err
	}
	conf := *ts.conf
	conf.ClientSecret = secret
	conf.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	return conf.TokenSource(ts.ctx, ts.token).Token()
}

func (s *SocialApple) oauthConfigWithClientSecret() (*oauth2.Config, error) {
	secret, err := generateAppleClientSecret(
		s.info.Extra[teamIDKey],
		s.info.Extra[keyIDKey],
		s.info.ClientId,
		s.info.Extra[privateKeyKey],
		s.info.Extra[privateKeyPathKey],
	)
	if err != nil {
		return nil, err
	}

	conf := *s.Config
	conf.ClientSecret = secret
	conf.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	return &conf, nil
}

func (s *SocialApple) UserInfo(ctx context.Context, client *http.Client, token *oauth2.Token) (*social.BasicUserInfo, error) {
	s.reloadMutex.RLock()
	defer s.reloadMutex.RUnlock()

	data, err := s.extractFromToken(ctx, client, token)
	if err != nil {
		return nil, err
	}

	formUser := parseAppleUserForm(appleFormUserFromContext(ctx))
	if data.Name == "" {
		data.Name = formUser.displayName()
	}
	if data.Email == "" {
		data.Email = formUser.Email
	}

	if data.Sub == "" {
		return nil, fmt.Errorf("error getting user info: id is empty")
	}

	if data.Email == "" {
		return nil, ErrEmailNotFound
	}

	if verified, present := appleEmailVerified(data.EmailVerified); present && !verified {
		return nil, fmt.Errorf("user email is not verified")
	}

	userInfo := &social.BasicUserInfo{
		Id:    data.Sub,
		Name:  data.Name,
		Email: data.Email,
		Login: data.Email,
	}

	if s.info.AllowAssignGrafanaAdmin && s.info.SkipOrgRoleSync {
		s.log.Debug("AllowAssignGrafanaAdmin and skipOrgRoleSync are both set, Grafana Admin role will not be synced, consider setting one or the other")
	}

	if !s.info.SkipOrgRoleSync {
		directlyMappedRole, grafanaAdmin, err := s.extractRoleAndAdminOptional(data.rawJSON, userInfo.Groups)
		if err != nil {
			s.log.Warn("Failed to extract role", "err", err)
		}

		if s.info.AllowAssignGrafanaAdmin {
			userInfo.IsGrafanaAdmin = &grafanaAdmin
		}

		userInfo.OrgRoles = s.orgRoleMapper.MapOrgRoles(s.orgMappingCfg, userInfo.Groups, directlyMappedRole)
		if s.info.RoleAttributeStrict && len(userInfo.OrgRoles) == 0 {
			return nil, errRoleAttributeStrictViolation.Errorf("could not evaluate any valid roles using IdP provided data")
		}
	}

	s.log.Debug("Resolved user info", "data", fmt.Sprintf("%+v", userInfo))
	return userInfo, nil
}

func (s *SocialApple) extractFromToken(ctx context.Context, client *http.Client, token *oauth2.Token) (*appleIDTokenClaims, error) {
	idToken := token.Extra("id_token")
	if idToken == nil {
		return nil, ErrIDTokenNotFound
	}

	idTokenString, ok := idToken.(string)
	if !ok {
		return nil, fmt.Errorf("id_token is not a string")
	}

	var rawJSON []byte
	var err error

	if s.info.ValidateIDToken && s.info.JwkSetURL != "" {
		rawJSON, err = s.validateIDTokenSignature(ctx, client, idTokenString, s.info.JwkSetURL)
		if err != nil {
			return nil, err
		}
	} else {
		rawJSON, err = s.retrieveRawJWTPayload(idTokenString)
		if err != nil {
			return nil, fmt.Errorf("error retrieving id_token: %w", err)
		}
	}

	var data appleIDTokenClaims
	if err := json.Unmarshal(rawJSON, &data); err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}
	data.rawJSON = rawJSON
	return &data, nil
}

func generateAppleClientSecret(teamID, keyID, clientID, privateKey, privateKeyPath string) (string, error) {
	if teamID == "" || keyID == "" || clientID == "" {
		return "", fmt.Errorf("apple team_id, key_id, and client_id are required to generate a client secret")
	}

	pemBytes, err := loadApplePrivateKeyPEM(privateKey, privateKeyPath)
	if err != nil {
		return "", err
	}

	key, err := parseApplePrivateKey(pemBytes)
	if err != nil {
		return "", err
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.ES256, Key: key},
		(&jose.SignerOptions{}).WithHeader("kid", keyID).WithType("JWT"),
	)
	if err != nil {
		return "", fmt.Errorf("error creating apple client secret signer: %w", err)
	}

	now := time.Now()
	claims := jwt.Claims{
		Issuer:   teamID,
		Subject:  clientID,
		Audience: jwt.Audience{appleAudience},
		IssuedAt: jwt.NewNumericDate(now),
		Expiry:   jwt.NewNumericDate(now.Add(appleClientSecretTTL)),
	}

	serialized, err := jwt.Signed(signer).Claims(claims).Serialize()
	if err != nil {
		return "", fmt.Errorf("error signing apple client secret: %w", err)
	}
	return serialized, nil
}

func loadApplePrivateKeyPEM(privateKey, privateKeyPath string) ([]byte, error) {
	if strings.TrimSpace(privateKey) != "" {
		return []byte(privateKey), nil
	}
	if strings.TrimSpace(privateKeyPath) == "" {
		return nil, fmt.Errorf("apple private_key or private_key_path is required")
	}
	return os.ReadFile(privateKeyPath)
}

func parseApplePrivateKey(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode apple private key PEM")
	}

	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		ecKey, ok := key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("apple private key must be ECDSA")
		}
		return ecKey, nil
	}

	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, fmt.Errorf("failed to parse apple private key")
}

func parseAppleUserForm(raw string) appleUserForm {
	if raw == "" {
		return appleUserForm{}
	}
	var user appleUserForm
	if err := json.Unmarshal([]byte(raw), &user); err != nil {
		return appleUserForm{}
	}
	return user
}

func (u appleUserForm) displayName() string {
	return strings.TrimSpace(strings.TrimSpace(u.Name.FirstName) + " " + strings.TrimSpace(u.Name.LastName))
}

func appleEmailVerified(raw json.RawMessage) (verified bool, present bool) {
	if len(raw) == 0 {
		return false, false
	}
	var asBool bool
	if err := json.Unmarshal(raw, &asBool); err == nil {
		return asBool, true
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return strings.EqualFold(asString, "true"), true
	}
	return false, false
}
