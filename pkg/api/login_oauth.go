package api

import (
	"net/http"
	"net/url"
	"path"

	"github.com/grafana/grafana/pkg/apimachinery/errutil"
	"github.com/grafana/grafana/pkg/infra/metrics"
	"github.com/grafana/grafana/pkg/login/social"
	"github.com/grafana/grafana/pkg/middleware/cookies"
	"github.com/grafana/grafana/pkg/services/authn"
	contextmodel "github.com/grafana/grafana/pkg/services/contexthandler/model"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/web"
)

const (
	OauthStateCookieName = "oauth_state"
	OauthPKCECookieName  = "oauth_code_verifier"
)

func (hs *HTTPServer) OAuthLogin(reqCtx *contextmodel.ReqContext) {
	name := web.Params(reqCtx.Req)[":name"]
	cookieOpts := hs.oauthCookieOptions(name)

	if errorParam := oauthCallbackParam(reqCtx, "error"); errorParam != "" {
		errorDesc := oauthCallbackParam(reqCtx, "error_description")
		hs.log.Error("failed to login ", "error", errorParam, "errorDesc", errorDesc)
		hs.redirectWithError(reqCtx, errutil.Unauthorized("oauth.login", errutil.WithPublicMessage(hs.Cfg.OAuthLoginErrorMessage)).Errorf("Login provider denied login request"))
		return
	}

	code := oauthCallbackParam(reqCtx, "code")
	redirectTo := oauthCallbackParam(reqCtx, "redirectTo")

	req := &authn.Request{HTTPRequest: reqCtx.Req}
	if code == "" {
		redirect, err := hs.authnService.RedirectURL(reqCtx.Req.Context(), authn.ClientWithPrefix(name), req)
		if err != nil {
			reqCtx.Redirect(hs.redirectURLWithErrorCookie(reqCtx, err))
			return
		}

		cookies.WriteCookie(reqCtx.Resp, OauthStateCookieName, redirect.Extra[authn.KeyOAuthState], hs.Cfg.OAuthCookieMaxAge, cookieOpts)

		//nolint:staticcheck // not yet migrated to OpenFeature
		if hs.Features.IsEnabledGlobally(featuremgmt.FlagUseSessionStorageForRedirection) {
			cookies.WriteCookie(reqCtx.Resp, "redirectTo", url.QueryEscape(redirectTo), hs.Cfg.OAuthCookieMaxAge, cookieOpts)
		}
		if pkce := redirect.Extra[authn.KeyOAuthPKCE]; pkce != "" {
			cookies.WriteCookie(reqCtx.Resp, OauthPKCECookieName, pkce, hs.Cfg.OAuthCookieMaxAge, cookieOpts)
		}

		reqCtx.Redirect(redirect.URL)
		return
	}

	identity, err := hs.authnService.Login(reqCtx.Req.Context(), authn.ClientWithPrefix(name), req)
	// NOTE: always delete these cookies, even if login failed
	cookies.DeleteCookie(reqCtx.Resp, OauthStateCookieName, cookieOpts)
	cookies.DeleteCookie(reqCtx.Resp, OauthPKCECookieName, cookieOpts)

	if err != nil {
		reqCtx.Redirect(hs.redirectURLWithErrorCookie(reqCtx, err))
		return
	}

	metrics.MApiLoginOAuth.Inc()
	authn.HandleLoginRedirect(reqCtx.Req, reqCtx.Resp, hs.Cfg, identity, hs.ValidateRedirectTo, hs.Features)
}

func oauthCallbackParam(reqCtx *contextmodel.ReqContext, name string) string {
	if v := reqCtx.Query(name); v != "" {
		return v
	}
	if reqCtx.Req.Method == http.MethodPost {
		return reqCtx.Req.FormValue(name)
	}
	return ""
}

func (hs *HTTPServer) oauthCookieOptions(provider string) func() cookies.CookieOptions {
	return func() cookies.CookieOptions {
		opts := hs.CookieOptionsFromCfg()
		if provider == social.AppleProviderName {
			// Apple's authorization response is a cross-site POST. Lax cookies
			// would be dropped, so oauth state/PKCE cookies must be SameSite=None; Secure.
			opts.SameSiteMode = http.SameSiteNoneMode
			opts.SameSiteDisabled = false
			opts.Secure = true
		}
		return opts
	}
}

func (hs *HTTPServer) registerAppleLoginCSRFExemption() {
	if hs.Csrf == nil {
		return
	}
	hs.Csrf.AddSafeEndpoint("/login/apple")
	if hs.Cfg != nil && hs.Cfg.AppSubURL != "" {
		hs.Csrf.AddSafeEndpoint(path.Join(hs.Cfg.AppSubURL, "login/apple"))
	}
}
