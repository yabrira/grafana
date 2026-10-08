package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/auth/authtest"
	contextmodel "github.com/grafana/grafana/pkg/services/contexthandler/model"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/services/user"
	"github.com/grafana/grafana/pkg/services/user/usertest"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/web"
)

func TestAppleStubLoginToggleOff(t *testing.T) {
	lookedUp := false
	hs := appleStubServer(t, featuremgmt.WithFeatures(), &usertest.FakeUserService{
		GetByEmailFn: func(context.Context, *user.GetUserByEmailQuery) (*user.User, error) {
			lookedUp = true
			return nil, user.ErrUserNotFound
		},
	})

	rec := appleStubRequest(t, hs)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.False(t, lookedUp)
}

func TestAppleStubLoginToggleOn(t *testing.T) {
	var gotEmail string
	hs := appleStubServer(t, featuremgmt.WithFeatures(featuremgmt.FlagAuthAppleStub), &usertest.FakeUserService{
		GetByEmailFn: func(_ context.Context, q *user.GetUserByEmailQuery) (*user.User, error) {
			gotEmail = q.Email
			return &user.User{ID: 7, Email: appleStubEmail, Login: appleStubEmail}, nil
		},
	})

	rec := appleStubRequest(t, hs)

	require.Equal(t, appleStubEmail, gotEmail)
	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "/", rec.Header().Get("Location"))
	require.Contains(t, rec.Header().Get("Set-Cookie"), "grafana_session=")
}

func TestAppleStubLoginCreatesPolicyCompliantPassword(t *testing.T) {
	var created user.Password
	hs := appleStubServer(t, featuremgmt.WithFeatures(featuremgmt.FlagAuthAppleStub), &usertest.FakeUserService{
		GetByEmailFn: func(context.Context, *user.GetUserByEmailQuery) (*user.User, error) {
			return nil, user.ErrUserNotFound
		},
		CreateFn: func(_ context.Context, cmd *user.CreateUserCommand) (*user.User, error) {
			created = cmd.Password
			return &user.User{ID: 7, Email: appleStubEmail, Login: appleStubEmail}, nil
		},
	})

	rec := appleStubRequest(t, hs)

	cfg := setting.NewCfg()
	cfg.BasicAuthStrongPasswordPolicy = true
	require.NoError(t, created.Validate(cfg))
	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "/", rec.Header().Get("Location"))
}

func appleStubServer(t *testing.T, features featuremgmt.FeatureToggles, users user.Service) *HTTPServer {
	t.Helper()
	cfg := setting.NewCfg()
	cfg.LoginCookieName = "grafana_session"
	return &HTTPServer{
		log:              log.New("test"),
		Cfg:              cfg,
		Features:         features,
		userService:      users,
		AuthTokenService: authtest.NewFakeUserAuthTokenService(),
	}
}

func appleStubRequest(t *testing.T, hs *HTTPServer) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/login/apple-stub", nil)
	rec := httptest.NewRecorder()
	hs.AppleStubLogin(&contextmodel.ReqContext{
		Context: &web.Context{
			Req:  req,
			Resp: web.NewResponseWriter(req.Method, rec),
		},
		SignedInUser: &user.SignedInUser{},
		Logger:       log.New("test"),
	})
	return rec
}
