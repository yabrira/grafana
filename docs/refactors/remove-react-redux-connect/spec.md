# Migration spec: remove-react-redux-connect

> The mechanical contract executors follow. Updated after the pilot — pilot surprises live here.

## Transformation patterns

### Pattern 1: Functional component — mapState + mapDispatch object

**Before:**

```tsx
import { connect, type ConnectedProps } from 'react-redux';
import { type StoreState } from 'app/types/store';
import { resetError, resetWarning } from './state/reducers';

function mapStateToProps(state: StoreState) {
  return {
    error: state.authConfig.updateError,
    warning: state.authConfig.warning,
  };
}

const mapDispatchToProps = {
  resetError,
  resetWarning,
};

const connector = connect(mapStateToProps, mapDispatchToProps);
type Props = ConnectedProps<typeof connector>;

export const ErrorContainerUnconnected = ({ error, warning, resetError, resetWarning }: Props) => {
  return (
    <div>
      {error && <button onClick={() => resetError()} />}
    </div>
  );
};

export default connector(ErrorContainerUnconnected);
```

**After:**

```tsx
import { useDispatch, useSelector } from 'app/types/store';
import { resetError, resetWarning } from './state/reducers';

export default function ErrorContainer() {
  const dispatch = useDispatch();
  const error = useSelector((state) => state.authConfig.updateError);
  const warning = useSelector((state) => state.authConfig.warning);

  return (
    <div>
      {error && <button onClick={() => dispatch(resetError())} />}
    </div>
  );
}
```

**Applies when:** Function component; `mapDispatchToProps` is an object of action creators; no `OwnProps`; default export is `connector(...)`.

### Pattern 2: Functional component — with OwnProps

**Before:** Store props mixed with parent props via `ConnectedProps & OwnProps`.

**After:**

- Keep parent-passed props as the component’s props interface
- Move store reads to `useSelector` / store writes to `useDispatch`
- Drop `ConnectedProps` and `connector`

**Applies when:** `OwnProps` (or equivalent) exists alongside connect.

### Pattern 3: Dispatch-only connect

**Before:** `connect(undefined, mapDispatchToProps)` or `connect(null, mapDispatchToProps)`.

**After:** `useDispatch` only; no `useSelector` unless needed.

**Applies when:** No `mapStateToProps` (e.g. `ExploreRunQueryButton.tsx`).

### Pattern 4: Class component + connect

**Before:** `class Foo extends PureComponent<Props>` wrapped with `connector`.

**After:** Rewrite as a function component using Patterns 1–3. Preserve memoization with `memo` only if profiling or existing `PureComponent` behavior is load-bearing and tests cover it; default is plain function unless the file already depended on bail-outs.

**Applies when:** File contains `extends PureComponent` / `extends Component` and imports `connect` from `react-redux`.

### Pattern 5: HOC stacking

**Before:** `export default withTheme2(connector(Explore));` (or similar).

**After:** Convert inner component to hooks first; keep outer HOCs (`withTheme2`, etc.) unless they can be replaced with existing hooks (`useStyles2` / `useTheme2`) **in a follow-up** — do not expand scope to theme migration inside this unit. Minimum: `withTheme2(ExploreWithHooks)`.

**Applies when:** `connector` is composed with another HOC.

## Edge cases

### Edge case 1: Tests import `*Unconnected` and inject store props

**Detect:** `*.test.tsx` imports `FooUnconnected` or builds `Props` from `ConnectedProps`.

**Handle:** In the same unit as the component migration:

1. Prefer rendering the default export inside a Redux `Provider` with a real/partial store (see existing feature test helpers), **or**
2. Mock `app/types/store`’s `useSelector` / `useDispatch`

Do not leave a permanent `*Unconnected` export solely for tests unless the pilot finds a blocker — if so, flag and document here.

### Edge case 2: `mapStateToProps(state, ownProps)`

**Detect:** Second argument to `mapStateToProps`.

**Handle:** Use component props inside `useSelector` carefully — either select broadly and derive, or select with a stable key from props (e.g. `state.panels[props.stateKey]`). Avoid returning new object identities every render without need; match prior memoization behavior if tests fail on render counts.

### Edge case 3: Named export `FooContent = connector(...)` + separate default page shell

**Detect:** e.g. `UsersListPage.tsx` — `UsersListPageContent = connector(...)` and `export default function UsersListPage()`.

**Handle:** Convert the content component to hooks; keep the page shell unchanged.

### Edge case 4: False-positive `connect(`

**Detect:** `centrifuge.connect()`, string literals containing `connect(`, etc.

**Handle:** Not in scope. Ratchet counts **only** `import { … connect … } from 'react-redux'`.

### Edge case 5: `connect` with mergeProps or options (4th argument)

**Detect:** `connect(mapState, mapDispatch, mergeProps, options)`.

**Handle:** **Flag and stop** — not covered by Patterns 1–3. Record in `progress.md`.

## Escape hatch

If a case matches none of the patterns or edge cases above: **do not improvise.** Set the unit to `flagged` in `progress.md` with the file, line, and why it doesn't match, then stop. A flagged unit is a successful outcome; a force-fitted diff is not.

## Verification

After each unit (adjust paths to the unit’s Scope):

```bash
yarn jest --no-watch --coverage=false <touched-test-files-or-directories>
yarn typecheck
yarn lint -- <touched-files>
```

Ratchet must not increase:

```bash
rg -l "import\s*\{[^}]*\bconnect\b[^}]*\}\s*from\s*['\"]react-redux['\"]" public/app --glob '*.{ts,tsx}' | wc -l
```

No Playwright requirement for this migration unless a unit touches Explore/DashboardPanel and the executor lacks Jest coverage — then note manual smoke of that surface in the PR.
