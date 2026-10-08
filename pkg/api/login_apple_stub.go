package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

	contextmodel "github.com/grafana/grafana/pkg/services/contexthandler/model"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/services/user"
)

const appleStubEmail = "apple-stub@example.com"

// AppleStubLogin signs in the demo user when authAppleStub is on.
// The toggle off path is a not found response, with no call to Apple.
func (hs *HTTPServer) AppleStubLogin(c *contextmodel.ReqContext) {
	if hs.Features == nil || !hs.Features.IsEnabled(c.Req.Context(), featuremgmt.FlagAuthAppleStub) {
		c.Resp.WriteHeader(http.StatusNotFound)
		return
	}

	usr, err := hs.appleStubUser(c)
	if err != nil {
		hs.redirectWithError(c, err)
		return
	}

	if err := hs.loginUserWithUser(usr, c); err != nil {
		hs.redirectWithError(c, err)
		return
	}

	c.Redirect(hs.Cfg.AppSubURL + "/")
}

func (hs *HTTPServer) appleStubUser(c *contextmodel.ReqContext) (*user.User, error) {
	ctx := c.Req.Context()
	usr, err := hs.userService.GetByEmail(ctx, &user.GetUserByEmailQuery{Email: appleStubEmail})
	if err == nil {
		return usr, nil
	}
	if !errors.Is(err, user.ErrUserNotFound) {
		return nil, err
	}

	// Session login does not use this password. A fresh random value keeps a
	// shared password out of the source tree. Hex is only [0-9a-f], so the
	// suffix supplies the classes [auth.basic] password_policy requires.
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}

	return hs.userService.Create(ctx, &user.CreateUserCommand{
		Email:         appleStubEmail,
		Login:         appleStubEmail,
		Name:          "Apple Stub",
		Password:      user.Password(hex.EncodeToString(buf) + "Aa1!"),
		EmailVerified: true,
	})
}
