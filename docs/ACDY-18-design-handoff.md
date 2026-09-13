# ACDY-18 design handoff

This handoff describes the Grafana login **Sign in with Apple** wireframe for [ACDY-18](https://fe-anysphere-demo.atlassian.net/browse/ACDY-18). Use it with the Figma frame when you refine the planted Apple button on the login page. The story is an incomplete OAuth stub, not a production Apple IdP.

## Source

The Designer agent run started from the ACDY-18 ticket and the existing demo wireframe (not a new file).

- **Jira:** [ACDY-18](https://fe-anysphere-demo.atlassian.net/browse/ACDY-18) — incomplete Apple Sign-In on this Grafana fork.
- **Figma:** [Grafana Login — Sign in with Apple](https://www.figma.com/design/r0eXTXohXl3OqiL1euwWlH?node-id=1-8) (file `r0eXTXohXl3OqiL1euwWlH`, node `1:8`).
- **Engineering trail:** [PR #14](https://github.com/yabrira/grafana/pull/14) on `demo/sama-apple-auth`, and the [ACDY-18 agent plan](./ACDY-18-agent-plan.md).
- **Designer agent run:** [bc-1251eae9-585c-54de-9831-0f99c855dcbb](https://cursor.com/agents/bc-1251eae9-585c-54de-9831-0f99c855dcbb).

The Figma caption on the card reads **Demo wireframe · ACDY-18 · incomplete OAuth stub**. That caption is walkthrough context. Do not ship it as product UI.

## Layout

The frame is a **1280 × 800** desktop login. A single **400 × 518** card sits in the horizontal center of a near-black page (card origin `x=440`, `y=141`). Form controls share a **328px** content width with **36px** side inset.

Stack the card from top to bottom:

1. **Grafana** wordmark — Grafana orange, centered.
2. **Welcome to Grafana** — large white heading.
3. **Sign in to continue** — smaller secondary line.
4. **Email or username** — 15px label over a **40px** full-width field.
5. **Password** — same field geometry, **16px** gap under the email block.
6. **Log in** — primary CTA, **328 × 41**, Grafana orange fill.
7. **or** divider — hairline rules on both sides of the word.
8. **Sign in with Apple** — secondary social CTA, same size as **Log in**.
9. Walkthrough caption under the Apple button (demo-only).

Keep password login first. Apple is an extra path under the divider, not a replacement for email and password.

## Sign in with Apple

Match the existing social-button row in `LoginServiceButtons.tsx`. The wireframe button is:

- **Label:** `Sign in with Apple` (service name **Apple**).
- **Size:** **328 × 41**, same as **Log in**.
- **Fill:** Apple black (`#000000`).
- **Text:** white / high-contrast on black.
- **Mark:** 16 × 16 light circle to the left of the label (wireframe placeholder for the Apple logo). Center the mark and label as one group.
- **Href:** `/login/apple` — still the incomplete OAuth stub. Do not add a backend connector from this handoff.

Hover may lighten the black fill slightly. Do not restyle the button as Grafana orange. Orange stays on **Log in** so the Apple path reads as a distinct provider.

## Dark theme

The wireframe is Grafana dark, not a light marketing page.

- **Page:** near-black canvas, no illustration.
- **Card:** slightly lighter dark surface, large corner radius, centered.
- **Wordmark:** Grafana orange on dark.
- **Headings:** white primary; the continue line is muted.
- **Fields:** darker than the card, full width, no extra chrome.
- **Primary CTA:** Grafana orange.
- **Apple CTA:** black on dark, separated by the **or** rule.

Reuse Grafana login tokens (`LoginLayout`, theme contrast on `#000000`) instead of inventing a second dark palette.

## Engineering map

This handoff does not change application code. Map the frame to the login stack already on `demo/sama-apple-auth`:

| Wireframe | Implementation |
| --- | --- |
| Page + centered card + titles | `LoginLayout.tsx` |
| Email, password, **Log in** | `LoginForm.tsx` |
| **or** + **Sign in with Apple** | `LoginServiceButtons.tsx` (`apple`: `bgColor: '#000000'`, `enabled: true`, `name: 'Apple'`) |

Leave the DEMO plants next to the Apple button (`dangerouslySetInnerHTML` of `apple_error`, hardcoded `APPLE_CLIENT_SECRET`). They are walkthrough findings. Do not remove them.

## Out of scope

These items are not part of the ACDY-18 stub:

- Official Apple logo SVG or Human Interface Guidelines legal artwork (the ellipse is enough for the demo).
- Production Sign in with Apple (Services ID, `.p8` key, identity token, nonce, or state).
- Light-theme or mobile-specific frames. Desktop dark is the walkthrough surface.
- Soft-spend, quota, or billing UI from older ticket text.

## Related resources

- [ACDY-18 in Jira](https://fe-anysphere-demo.atlassian.net/browse/ACDY-18)
- [Figma — Grafana Login — Sign in with Apple](https://www.figma.com/design/r0eXTXohXl3OqiL1euwWlH?node-id=1-8)
- [ACDY-18 agent plan](./ACDY-18-agent-plan.md)
- [PR #14 — Login: Apple authentication stub](https://github.com/yabrira/grafana/pull/14)
