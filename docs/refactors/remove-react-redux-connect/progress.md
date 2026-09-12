# Progress: remove-react-redux-connect

> Coordination file for the plan in `plan.md`. This file is the only channel between agents.
> Update it via commits — never rely on chat or memory. Claim before you work.

## Protocol

1. **Claim:** set the unit's Status to `claimed`, fill Agent/Branch, commit immediately (message: `claim <unit-id>`). If the commit conflicts because another agent claimed first, pick a different unit.
2. **Work:** stay inside the unit's Scope. If anything matches no pattern in the plan/spec: set Status `flagged`, write the reason in Notes, commit, stop.
3. **Deliver:** verification green → open PR, set Status `in-review` with the PR link, commit.
4. **Done:** set on merge (by the merging human or a follow-up agent).

Statuses: `unclaimed` → `claimed` → `in-review` → `done`, plus `flagged` (needs a human decision) and `blocked` (dependency not done).

## Units

| Unit | Title | Status | Agent / Branch | PR | Notes |
|---|---|---|---|---|---|
| S1-1 | Connect-import ratchet | in-review | cursor-agent / refactor/remove-react-redux-connect/S1-1 | https://github.com/yabrira/grafana/pull/4 | |
| S1-2 | Characterization — admin | unclaimed | | | |
| S1-3 | Characterization — org/invites/support-bundles/auth-config | unclaimed | | | |
| S1-4 | Characterization — dashboard/explore/variables gaps | unclaimed | | | |
| S2-1 | Pilot — ErrorContainer, UsersActionBar, InviteeRow | unclaimed | | | deps: S1-1, S1-3 |
| S2-2 | Lever — functional connect → hooks | unclaimed | | | deps: S2-1 |
| S3-1 | auth-config remainder | unclaimed | | | deps: S2-2, S1-3 |
| S3-2 | users + org + invites remainder | unclaimed | | | deps: S2-2, S1-3 |
| S3-3 | admin | unclaimed | | | deps: S2-2, S1-2 |
| S3-4 | profile + serviceaccounts + support-bundles | unclaimed | | | deps: S2-2, S1-3 |
| S3-5 | explore — functional containers | unclaimed | | | deps: S2-2, S1-4 |
| S3-6 | dashboard — functional only | unclaimed | | | deps: S2-2, S1-4 |
| S3-7 | variables — functional only | unclaimed | | | deps: S2-2, S1-4 |
| S3-8 | Class components — Explore, DashboardPanel, QueryVariableEditor | unclaimed | | | deps: S3-5, S3-6, S3-7, S1-4 |
| S4-1 | ESLint ban + docs | unclaimed | | | deps: all S3 done; ratchet must be 0 |

## Flags needing a decision

_(none yet)_
