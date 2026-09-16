# RABO-1847 — Sign in with Apple (frontend stub)

ADO work item RABO-1847 remains the source of truth for status, scope, and acceptance. This file
only records implementation detail that does not belong in the ticket.

## Scope

Frontend stub only. There is no Apple OAuth connector, no client ID, no client secret, and no
backend route. Nothing in this change talks to Apple.

## What was built

`public/app/core/components/Login/AppleSignInButton.tsx` renders a "Sign in with Apple" affordance
on the login page, styled like the existing social sign-in buttons.

Two decisions are worth calling out:

- It renders a `Button`, not a `LinkButton`. Every other entry in `loginServices()` is a link to
  `login/<provider>`, which only works because the backend has a matching OAuth connector. Apple
  has none, so a link would 404 and surface to the user as a generic login failure. A `Button` has
  no href and therefore no route to navigate to.
- Clicking it reveals an inline `Alert severity="info"` saying the method is not configured. This
  is the "safe no-op with feedback" behaviour the architect asked for — the user learns why nothing
  happened instead of being bounced to an error page.

`getServiceStyles` and `getButtonStyleFor` in `LoginServiceButtons.tsx` are now exported so the
stub matches the other buttons instead of duplicating their styling. `getButtonStyleFor` takes a
`bgColor` string rather than a whole `LoginService`, since the stub is not a configured service.

## Behaviour change to review

`LoginServiceButtons` previously rendered `null` when no OAuth provider was configured. Because the
stub must be visible on a default instance (where no OAuth is configured), that guard is gone and
the component always renders the divider plus the Apple button. On an instance with no social
logins, the login page now shows an "or" divider and the Apple stub where it previously showed
nothing. This is intentional for the stub but is the main thing to sign off on.

## Removed

A previously merged `apple` entry in `loginServices()` was deleted. It was always enabled, linked
to the non-existent `login/apple` route, and displayed the raw provider id `appleid` to the user.

## Not done

- No Apple icon exists in the Grafana icon set, so the stub reuses the generic `signin` icon. Real
  integration should add the Apple mark.
- No feature toggle. If the stub should be hideable per instance before it ships, that needs a
  toggle in `pkg/services/featuremgmt/` and a separate backend PR.

## Verification

```bash
yarn jest --watchAll=false public/app/core/components/Login
yarn eslint public/app/core/components/Login --ext .ts,.tsx
yarn typecheck
make i18n-extract   # produces no diff beyond the three added login.apple.* keys
```
