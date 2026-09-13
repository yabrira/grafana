# ACDY-18 agent plan

This is the Demo 3 Engineer-story plan that produced [PR #14](https://github.com/yabrira/grafana/pull/14) on `demo/sama-apple-auth`. Yanis walks agents in Cursor from this plan, not only from the GitHub tabs.

## Inputs

- **Jira:** [ACDY-18](https://fe-anysphere-demo.atlassian.net/browse/ACDY-18) — incomplete Apple Sign-In on this Grafana fork (ticket text may still mention an earlier spend-limit story; the work is Apple auth).
- **Figma:** [Grafana Login — Sign in with Apple](https://www.figma.com/design/r0eXTXohXl3OqiL1euwWlH?node-id=1-8) (`r0eXTXohXl3OqiL1euwWlH`, node `1:8`).
- **Design handoff:** [ACDY-18-design-handoff.md](./ACDY-18-design-handoff.md) — Designer agent write-up of that wireframe (layout, Apple button, dark theme).
- **Base:** `demo/sama` already plants an always-visible Apple button with no backend OAuth connector.

## Goal

Ship a small, reviewable **frontend Apple Sign-In stub** on the Grafana login page so the walkthrough can show plan → implement → Bugbot / Security → Demo CI.

## Scope

- Refine the planted Apple button in `public/app/core/components/Login/LoginServiceButtons.tsx` (label **Apple**, Apple-black styling, still `enabled: true`, still `/login/apple`).
- Add `.cursor/BUGBOT.md` so Bugbot and Security know what to flag vs ignore.
- Add `.github/workflows/demo-ci.yml` (fails only if `DEMO_CI_FAIL` appears in `LoginServiceButtons.tsx`).
- Keep the change login-adjacent. Do not expand into a real IdP.

## Non-goals

- Production Apple IdP, Services ID, or `.p8` key handling.
- Identity-token / nonce / state verification.
- Billing, quota, or soft-spend UI.

## Demo plants (do not ship)

Intentional findings next to the Apple button so Bugbot and Security leave **visible inline comments**:

- XSS: `dangerouslySetInnerHTML` of the unsanitized `apple_error` query param.
- Secret: hardcoded `APPLE_CLIENT_SECRET` private-key stub in source (also on `data-apple-client-secret`).

Marked `// DEMO: intentional finding for Bugbot/Security walkthrough — do not ship`. Leave them in place for the walkthrough.

## Gates

- **Bugbot** — XSS and secret plants must produce review comments.
- **Security Agent** — same two issues on the Apple login path.
- **Demo CI** — green unless `DEMO_CI_FAIL` is added to `LoginServiceButtons.tsx` (CI triage narrative).

## Cloud agent

This plan was executed on cloud agent run [bc-57518ef2-5dc9-595d-83ab-8a6f187d0cd6](https://cursor.com/agents/bc-57518ef2-5dc9-595d-83ab-8a6f187d0cd6).
