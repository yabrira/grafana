# ACDY-18 refined story

This is the Demo 3 PM review of [ACDY-18](https://fe-anysphere-demo.atlassian.net/browse/ACDY-18) (`Add Sign in with Apple on Grafana login`). The Jira ticket is the source of truth for intent, but it is incomplete: acceptance criteria are vague, the IdP is still TBD, security notes are missing, and error states are unspecified. This document tightens the story so design and engineering can ship a reviewable **frontend Sign in with Apple stub** without treating the ticket as a production Apple Sign In project.

Related implementation: [PR #14](https://github.com/yabrira/grafana/pull/14) on `demo/sama-apple-auth` (base `demo/sama`).

## Problem

Admins and users who expect Apple as a login option cannot see a Sign in with Apple path on this Grafana fork. The current ticket asks for “add Apple login” without saying whether that means a real Apple IdP, a UI-only stub, or a full OAuth connector.

For this demo increment, the product ask is the **login-page affordance only**: a visible, correctly labeled Apple button that matches design and sits next to the existing password and social login paths. A production Apple Services ID and token verification are **not** in this story.

## User story

As a Grafana user on the login page, I want a **Sign in with Apple** button that looks like the other social login actions so I can start an Apple sign-in from the same place I already use email/password or Google/GitHub/Okta.

As a reviewer of ACDY-18, I want the stub to be explicit and incomplete on purpose so we can walk the AI SDLC path (plan → implement → Bugbot / security review → CI) without pretending Apple is a configured IdP.

## Design

Use the existing login wireframe:

- [Grafana Login — Sign in with Apple](https://www.figma.com/design/r0eXTXohXl3OqiL1euwWlH?node-id=1-8) (`r0eXTXohXl3OqiL1euwWlH`, node `1:8`)

The frame shows the standard Grafana login card (email or username, password, **Log in**), an **or** divider, then a full-width black **Sign in with Apple** button. The canvas caption calls this an incomplete OAuth stub. Design is the source for label, placement, and Apple-black styling. It is not a spec for a working Apple identity provider.

## In scope

This story covers a **frontend-only stub** on the Grafana login page:

- A **Sign in with Apple** action that is always visible on `/login` for this demo, even when no Apple OAuth connector is configured.
- Display name **Apple**, so the existing “Sign in with {provider}” pattern renders **Sign in with Apple**.
- Apple-black button styling consistent with the Figma frame and with other social login buttons (full width, icon + label).
- Navigation target `/login/apple` (or `login/apple` as used by the existing social-button href pattern). The route may still be incomplete.
- Password login and already-enabled social providers (Google, GitHub, Okta, and others gated by `config.oauth`) stay unchanged.
- Keep the change login-adjacent. Do not add billing, quota, or soft-spend UI.

## Non-goals

Do **not** do the following in this story or this demo increment:

- Register or require a production Apple Services ID, Team ID, Key ID, or `.p8` key.
- Implement Apple token verification (identity token `iss` / `aud` / `exp` / `nonce` checks).
- Exchange an authorization code, mint a client-secret JWT, or add a backend OAuth connector for Apple.
- Treat the IdP as TBD-and-still-in-scope. For this increment the IdP is **explicitly deferred**.
- Hide-my-email account provisioning, account linking, or org-mapping for Apple relay addresses.
- Feature-flag the stub behind a GA toggle for this walkthrough (the plant is meant to be visible).
- Remove or “fix” DEMO walkthrough plants if they are present on the branch. Those findings are intentional for Bugbot and security review.

## Acceptance criteria

These criteria replace the vague “Apple login works” language on ACDY-18.

1. **Visible stub.** Opening `/login` shows a **Sign in with Apple** control without any Apple Services ID, OAuth app, or backend Apple connector being configured.
2. **Label.** The control’s accessible name is **Sign in with Apple** (provider display name **Apple**).
3. **Placement.** The button appears below the **or** divider, after the email/password **Log in** path, matching the [Figma login card](https://www.figma.com/design/r0eXTXohXl3OqiL1euwWlH?node-id=1-8).
4. **Styling.** The button uses Apple-black styling (black background, light label/icon contrast) and the same full-width social-button treatment as other login providers.
5. **Stub href.** Activating the button navigates to the Apple stub path `login/apple`. A completed Apple callback is **not** required to pass this story.
6. **No regression.** Email/password login still renders when the login form is enabled. Existing social providers still appear only when their `config.oauth` entries are enabled.
7. **No production IdP work.** Shipping this increment does not add Apple token verification, a Services ID, or server-side client-secret generation.
8. **Demo plants stay.** If the branch includes DEMO-marked XSS or client-secret plants next to the Apple button, they remain in place for the walkthrough. This story does not ask anyone to sanitize or delete them.

## Error states to cover later

ACDY-18 had no error-state criteria. The following are **out of scope** for the frontend stub and should become their own stories once a real Apple IdP exists:

- **User cancels** the Apple identity sheet or returns `access_denied` / `user_cancelled_authorize`.
- **Apple IdP errors** (invalid_request, invalid_client, invalid_grant, server_error).
- **Incomplete route.** `/login/apple` is not registered or returns 404 because there is no backend connector yet.
- **Token failures.** Authorization-code exchange fails; identity token is missing, expired, unsigned, or has the wrong audience.
- **Nonce / state mismatch.** Callback state or nonce does not match the value started on `/login`.
- **Account identity.** Hide-my-email relay addresses, missing email, or an Apple account that does not match an existing Grafana user.
- **Account linking.** Apple identity is already tied to a different Grafana user.
- **Network / timeout** talking to Apple or the Grafana callback.
- **Safe error display.** Any `apple_error` (or similar) query param must be shown as escaped text through the existing login alert — never as raw HTML. Defining the copy and the sanitized render path is a follow-up, not this stub.

Until those stories exist, a click on **Sign in with Apple** may land on an incomplete `/login/apple` path. That is acceptable for this increment.

## Security notes and follow-ups

The original ticket omitted security. Capture these as follow-ups; do not block the frontend stub on them.

Must not ship in a real Apple Sign In (file separately from this demo):

- **Secrets in the client.** Never put an Apple client secret, `.p8` private key, Team ID, or Services ID in frontend source, attributes, or repo files. Generate the client-secret JWT on the server and store material in env/secret storage.
- **Reflected XSS.** Never render login or callback query/hash input with `dangerouslySetInnerHTML` / `innerHTML`. Use static React text or the existing login `Alert`.
- **Verify tokens server-side.** Validate Apple identity tokens (`iss`, `aud`, `exp`, nonce) before creating a session. Do not trust an unsigned or client-supplied identity.
- **CSRF / replay.** Persist and check OAuth `state` (and Apple `nonce`) on the callback. Do not accept a callback that skips those checks.
- **Redirect URIs.** Register exact callback URLs with Apple. Do not take redirect targets from the client.
- **Do not weaken** password login or SAML while iterating on Apple.

Walkthrough note: DEMO-marked findings next to the Apple button (unsanitized `apple_error`, hardcoded `APPLE_CLIENT_SECRET`) are planted so Bugbot and the security agent leave inline comments. Flag them in review. Do not treat them as the production design.

## Resolved from the incomplete ticket

| Original gap | PM call |
| --- | --- |
| Vague “Apple login works” ACs | Testable UI stub criteria above. Success is a visible, labeled, Apple-black button to `login/apple`. |
| IdP TBD | IdP is **out of scope**. Do not wait on a Services ID for this increment. |
| Missing security notes | Security follow-ups listed; production verification is a later story. |
| No error-state criteria | Error matrix listed as later work. Stub may hit an incomplete route. |
| Possible leftover spend-limit wording on ACDY-18 | This story is Apple auth only. No quota or soft-spend UI. |

## Related

- Jira: [ACDY-18](https://fe-anysphere-demo.atlassian.net/browse/ACDY-18)
- Figma: [Sign in with Apple](https://www.figma.com/design/r0eXTXohXl3OqiL1euwWlH?node-id=1-8)
- PR: [yabrira/grafana#14](https://github.com/yabrira/grafana/pull/14)
- Engineer plan: [`docs/ACDY-18-agent-plan.md`](./ACDY-18-agent-plan.md)
- PM agent run: [Demo3 PM — refine ACDY-18](https://cursor.com/agents/bc-cda2b208-9423-58e2-ac85-485cb3ef4f04)
