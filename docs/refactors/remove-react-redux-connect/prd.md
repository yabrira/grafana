# PRD: remove-react-redux-connect

> Why and what. Keep to ~1 page. The out-of-scope list is the most load-bearing section.

## Problem

Grafana still has **38 files** importing `connect` from `react-redux` (HOC-era Redux bindings). Newer code uses typed `useSelector` / `useDispatch` from `app/types/store`. The two patterns coexist, which slows agents and humans: every Redux touch requires knowing which style a file uses, and `contribute/style-guides/redux.md` still documents `connect` as the typed example. Removing `connect` is a low-behavior-risk, high-mechanical-yield cleanup that aligns the frontend with the hooks-first direction already used in browse-dashboards, alerting, and elsewhere.

## Done metric

**Zero** files under `public/app/` that import `connect` from `react-redux`.

Ratchet command (may only decrease):

```bash
rg -l "import\s*\{[^}]*\bconnect\b[^}]*\}\s*from\s*['\"]react-redux['\"]" public/app --glob '*.{ts,tsx}' | wc -l
```

Starting count: **38** (2026-08-04).

## In scope

- Replace `connect(mapStateToProps, mapDispatchToProps)` (and variants) with `useSelector` / `useDispatch` from `app/types/store`
- Convert class components that only use `connect` for store access into function components as part of the same unit
- Update adjacent Jest tests that import `*Unconnected` / `ConnectedProps`
- Add a count ratchet, then an ESLint ban on importing `connect` from `react-redux`
- Update `contribute/style-guides/redux.md` to show the hooks pattern only

## Out of scope

> Executors: anything on this list appearing in a diff is grounds to reject the PR.

- Migrating `getBackendSrv()` call sites or thunks to `@grafana/api-clients` (separate refactor; intake answer was ambiguous — see plan Open questions)
- Rewriting Redux slices / reducers / thunks themselves
- Migrating legacy dashboard → Scenes, or the variables state model
- Removing `react-redux` as a dependency (Provider and hooks stay)
- Backend / Go changes
- Plugin public APIs outside the listed `connect` call sites (only `public/app/**` files in the inventory)
- Behavioral product changes, copy, or UX redesigns while converting bindings

## Constraints

- Feature work continues — no freeze; ratchet must land early so new `connect` usage cannot grow
- Medium risk tolerance — prefer small PRs per feature directory; no feature-flag required (prop-wiring only)
- Verification: scoped `yarn jest --no-watch <paths>`, plus `yarn typecheck` and `yarn lint` on touched areas
- Timeline / budget: **unknown** (intake) — plan sized for incremental stacked PRs over days–weeks, not a big-bang
- Fragile / politically untouchable: **unknown** (intake) — treat Explore, `DashboardPanel`, and variables editors as extra-care (characterization + smaller PRs)
