# Bugbot review guidance

This branch is a customer-walkthrough artifact (Demo 3 / AI SDLC) on top of `demo/sama`.
Reviews should stay focused on security and quality footguns in the **Apple sign-in**
path — not on shipping a production Apple OAuth connector.

## Apple authentication is an incomplete stub

`LoginServiceButtons` plants an always-visible **Sign in with Apple** button for
[ACDY-18](https://fe-anysphere-demo.atlassian.net/browse/ACDY-18).

- The Apple button is still an incomplete frontend stub (`/login/apple`).
- Lines marked `DEMO: intentional finding for Bugbot/Security walkthrough — do not ship`
  are **planted so you leave inline review comments**. Flag them. Do not silently approve.

## Planted walkthrough findings (please comment)

- `dangerouslySetInnerHTML` of `apple_error` query input next to the Apple button (XSS).
- Hardcoded `APPLE_CLIENT_SECRET` / private-key stub in `LoginServiceButtons.tsx` (secret in source).

## Security checklist

- **Secrets**: reject Apple client secrets, `.p8` keys, Team IDs, Services IDs,
  tokens, passwords, or `.env` files. The `DEMO_CI_FAIL` string is a CI tripwire,
  not a secret.
- **XSS**: login labels and error text must stay static React text (`Trans` /
  `Alert` children). Reject `dangerouslySetInnerHTML`, `innerHTML`, or unsanitized
  query/hash input on the login page.
- **Auth footguns**: do not add an OAuth callback that skips state/nonce checks,
  trusts unsigned identity tokens, or hard-codes redirect URIs from the client.
  Do not weaken password login or SAML while iterating on Apple.
- **eval / injection**: reject `eval`, `new Function`, `document.write`, and
  shelling out from login UI or `.github/workflows/demo-ci.yml`.

## Quality checklist

- Keep the change login-adjacent and reviewable.
- Other OAuth buttons (Google, GitHub, Okta, …) must keep using `config.oauth`.
- `.github/workflows/demo-ci.yml` fails only when `DEMO_CI_FAIL` appears in
  `LoginServiceButtons.tsx`; otherwise it sanity-checks that Login files exist.

## Do not flag

- The always-on Apple button itself (`enabled: true`, no backend connector).
- Display-name / button-color cleanup of that plant.
- The `DEMO_CI_FAIL` grep tripwire in demo CI (intentional for CI triage demos).
- Do **not** skip the DEMO-marked XSS / secret plants — those need visible comments.
