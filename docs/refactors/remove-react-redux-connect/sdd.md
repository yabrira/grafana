# SDD: remove-react-redux-connect

> How. Target architecture, mapping, invariants, sequencing.

## Target architecture

All React components that read or write Redux state use the typed hooks exported from `app/types/store`:

- `useSelector` — typed with `RootState`
- `useDispatch` — typed with `AppDispatch`

ESLint already forbids importing `useSelector` / `useDispatch` directly from `react-redux` (must use `app/types/store`). After this migration, importing **`connect`** from `react-redux` is also forbidden.

`react-redux` remains for `<Provider>` and the hook implementations. No change to store shape, middleware, or RTK Query APIs.

## Old → new mapping

| Old | New | Notes |
|---|---|---|
| `import { connect, ConnectedProps } from 'react-redux'` | `import { useDispatch, useSelector } from 'app/types/store'` | Never import hooks from `react-redux` directly |
| `mapStateToProps(state)` → props | `useSelector((state) => …)` inside the component | Prefer existing selectors when present |
| `mapStateToProps(state, ownProps)` | `useSelector` + closure over props / `useParams` etc. | OwnProps stay as normal function props |
| `mapDispatchToProps` object of action creators | `const dispatch = useDispatch();` then `dispatch(actionCreator(…))` or wrap in callbacks | Bound action-creator props become local functions |
| `mapDispatchToProps` function form | Same — call `dispatch` explicitly | Rare; flag if unusual |
| `connect(undefined, mapDispatch)` | `useDispatch` only | e.g. ExploreRunQueryButton |
| `connect(mapState)` (no dispatch) | `useSelector` only | |
| `type Props = ConnectedProps<typeof connector> & OwnProps` | Explicit props interface for parent-passed props only; store values are local | Tests that injected store props must be updated |
| `export const FooUnconnected` + `connector(FooUnconnected)` | Single function component using hooks; default/named exports preserved for routers | Update tests that imported `*Unconnected` |
| Class + `connect` | Convert to function component + hooks in the same unit | Hooks cannot be used in class bodies |

## Invariants (must hold throughout the migration)

- **No intentional behavior change** — same selectors, same actions, same render conditions
- **Export surface for routers** — default exports / page wrappers keep working (`UsersListPage` pattern: page shell + content)
- **Typed hooks only via `app/types/store`** — existing ESLint rule stays enforced
- **Ratchet never increases** — connect-import count is monotonically non-increasing
- **Tests pin current behavior** — characterization tests added for previously untested files must not “fix” bugs; migration PRs update assertions only when wiring shape changes (Unconnected → hooks)

## Sequencing

Strangler per feature directory:

1. **Concrete** — ratchet so feature work cannot add new `connect` usage
2. **Characterization** — add smoke/render tests for connect files that lack adjacent tests (before editing those files)
3. **Pilot** — 2–3 simple functional components; fold surprises into `spec.md`
4. **Lever** — small transform helper / codemod for the common functional pattern; class components stay hand-done
5. **Fan-out** — one PR per directory group; parallel agents OK after pilot+lever
6. **Flip** — ESLint ban `connect`; docs update; confirm count = 0

Explore / `DashboardPanel` / `QueryVariableEditor` (class components) run late, after characterization, in dedicated units.

## Failure modes this architecture eliminates

- **Pattern drift** — ESLint ban makes new `connect` impossible
- **Wrong import path for hooks** — existing restriction to `app/types/store` remains
- **Silent HOC wrapper bugs** — no more `connector(Component)` / `withTheme2(connector(…))` stacking surprises for agents
- **Test-only Unconnected forks** — over time one component definition; tests use Provider or mocked hooks consistently
