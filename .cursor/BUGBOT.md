# Bugbot review guidance

This branch is a customer-walkthrough artifact (Demo 3 / AI SDLC) on top of `demo/sama`.
Reviews should stay focused on security and quality footguns — not on turning the
soft-spend fixture into a production billing product.

## Soft spend banner is awareness-only

`SoftSpendLimitBanner` is a **static login-page fixture** for [ACDY-18](https://fe-anysphere-demo.atlassian.net/browse/ACDY-18).

- Numbers (for example ~82% of a monthly soft limit) are hardcoded demo data.
- The banner must **not** block login, payments, or any auth path.
- It must **not** call billing, quota, or payment APIs.
- It must **not** carry secrets, tokens, or live customer balances.
- Do not flag the fixture amounts themselves as "hardcoded credentials."

## Security checklist

- **Secrets**: reject API keys, tokens, passwords, `.env` files, or vault material
  checked into this PR. The `DEMO_CI_FAIL` string is a CI tripwire, not a secret.
- **XSS**: banner copy must stay static React text (`t` / `Trans` / `Alert` children).
  Reject `dangerouslySetInnerHTML`, `innerHTML`, or unsanitized query/hash input
  rendered into the banner.
- **Auth footguns**: do not change OAuth / SAML / password login behavior except
  the existing Apple sign-in demo plant on `demo/sama` (leave that plant alone).
  Do not treat a client-side soft-limit banner as authorization.
- **eval / injection**: reject `eval`, `new Function`, `document.write`, and
  shelling out from the banner or from `.github/workflows/demo-ci.yml`.

## Quality checklist

- Keep the change login-adjacent and reviewable.
- Login must still submit when the banner is visible.
- `.github/workflows/demo-ci.yml` fails only when `DEMO_CI_FAIL` appears in
  `LoginServiceButtons.tsx`; otherwise it sanity-checks that Login files exist.

## Do not flag

- The existing Apple sign-in demo plant in `LoginServiceButtons.tsx`.
- Hardcoded SAR fixture amounts in `SoftSpendLimitBanner`.
- The `DEMO_CI_FAIL` grep tripwire in demo CI (intentional for CI triage demos).
