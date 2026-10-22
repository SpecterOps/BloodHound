# Color Token Experiment: Badge

This document describes an experiment to restyle the `Badge` component with component-scoped CSS variables. The rest of the library still uses the existing theme setup.

## What changed

| File                                   | Change                                                                                                                      |
| -------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| `src/tailwind/plugin.ts`               | Added `--badge-*` CSS variables to both the light (`:root`) and dark (`.dark`) blocks.                                      |
| `src/tailwind/preset.ts`               | Removed the `badgeColors` import. Badge colors are no longer part of the Tailwind theme.                                    |
| `src/components/Badge/Badge.styles.ts` | Removed the `badgeColors` hex map and every `dark:` class. Variants now read CSS variables through arbitrary-value classes. |

### Before

Badge colors were raw hex values exported from `Badge.styles.ts`, each with a separate `-dark` copy. They were registered in the Tailwind theme through `preset.ts`. Every variant and color pair had its own rule in `compoundVariants`, with a `dark:` override:

```ts
{ variant: 'fill', color: 'primary', className: 'bg-badge-primary-surface dark:bg-badge-primary-surface-dark' },
```

### After

`plugin.ts` defines each badge color once per theme:

```ts
// :root
'--badge-primary-surface': light.badge.primary.fill,
// .dark
'--badge-primary-surface': dark.badge.primary.fill,
```

Each `color` variant sets two local variables on the badge element, `--badge-surface` and `--badge-accent`. Each `variant` reads those two variables:

```ts
variant: {
    fill: 'bg-[color:var(--badge-surface)]',
    outline: [
        'border border-[color:var(--badge-accent)] bg-[color:var(--badge-grey-surface)]',
        '[&>span>svg]:text-[color:var(--badge-accent)]',
    ],
},
color: {
    primary: '[--badge-surface:var(--badge-primary-surface)] [--badge-accent:var(--badge-primary-accent)]',
    // ...
},
```

The `error`, `warning`, and `success` surfaces are badge variables whose values point at the status variables. This keeps the original look: the status `main` color in light mode and the status `fill` color in dark mode.

The 22 compound variants were reduced to one. The remaining rule is `[&>span>svg]:text-inherit` for grey outline badges. It keeps the original rendering, in which the icon inherits its color.

## Why

-   **Theme switching in one place.** The light and dark values live in the plugin, so components no longer need `dark:` classes.
-   **One direction of dependency.** The theme setup (`preset.ts`) no longer imports from a component file. Colors now flow one way: `colors.ts` → `plugin.ts` → component styles.
-   **A smaller shared theme.** Badge colors are used only by the Badge component. Registering them as Tailwind colors added `bg-`, `text-`, `ring-`, `fill-`, `stroke-`, and other badge classes to every app that uses the preset.
-   **Simpler variants.** The variant (fill or outline) and the color are now separate: a color sets the variables, and a variant uses them.

## Pros

-   No `dark:` classes in Badge styles. The theme switches automatically through CSS variables.
-   Badge-only colors don't appear in the shared Tailwind theme or in editor autocomplete for consumers.
-   `preset.ts` no longer depends on `Badge.styles.ts`.
-   `Badge.styles.ts` went from 221 lines to 65.
-   Adding a color takes one line in `color` plus its variables in `plugin.ts`, with no new compound variants.
-   The CSS variables still support runtime theming and rebranding by overriding them at the `:root` or `.dark` level.
-   No visual change. Light and dark values match the previous implementation.

## Cons

-   **No editor autocomplete** for `bg-[color:var(--badge-surface)]`-style classes.
-   **No typo protection.** A misspelled variable name still produces a valid class and silently renders no color. A test that checks referenced variables against both themes would cover this.
-   **More verbose class syntax.** The `color:` type hint is required so that Tailwind does not read a value like `border-[var(--x)]` as a width.
-   **Two patterns in the library.** Badge works differently from the components that still use theme-registered tokens, until the approach is adopted or reverted.
-   **Indirection.** The color comes from a chain (color variant → local variable → theme variable → `colors.ts`), which is harder to trace in dev tools.
-   **Breaking change for external class usage.** The `bg-badge-*`, `border-badge-*`, `text-badge-*`, and `*-dark` utility classes are no longer generated. No usages exist in this repository, but downstream consumers should be checked.
-   **Nested overrides are limited.** Variables that reference other variables are resolved where they are declared (`:root` / `.dark`). Overriding a `--status-*` variable on a nested element does not change the badge colors inside it.

## Guideline going forward

-   **Add colors to the Tailwind theme** when they are shared across components or used by consumers (for example `primary`, `text-*`, `status-*`, `neutral-*`).
-   **Use CSS variables only** for colors private to a single component, and consume them in that component's styles with arbitrary-value classes.

## Rolling this out to other components

Because of the PR size limit, migrate **two or three components per PR**. Do the components with exported color maps first, since they currently feed `preset.ts`:

1. `Switch` (`switchBackgroundColors`)
2. `Menu` (`menuBackgroundColors`, `menuBorderColors`, `menuTextColors`)
3. `Button` (`buttonBackgroundColors`, `buttonBorderColors`)

After those, migrate the components that use `dark:` classes or component-specific theme colors (for example `Checkbox`, `Input`, `Select`, `Toggle`, `Textarea`, `DataTable`).

### Agent prompt

Copy the prompt below and replace `<COMPONENTS>` with the components for this PR.

```text
Migrate the following Doodle UI components to the component-scoped color variable pattern: <COMPONENTS>.

Context:
- Read packages/javascript/doodle-ui/COLOR_TOKEN_EXPERIMENT.md first. Badge is the reference implementation:
  src/components/Badge/Badge.styles.ts, the "Badge" sections of src/tailwind/plugin.ts, and src/tailwind/preset.ts.
- Only change the listed components. Keep the PR small; do not migrate anything else.

For each component:
1. Find every color it uses: exported color maps in *.styles.ts, component-specific colors registered in
   src/tailwind/preset.ts, hex arbitrary values (e.g. bg-[#0D0A30]), and dark: classes in *.styles.ts or *.tsx.
2. In src/tailwind/plugin.ts, add a "// // Components/<Name>" section to BOTH the ' :root' and '.dark' blocks,
   defining --<component>-<role> variables (e.g. --switch-track-checked). Use values from src/tailwind/colors.ts.
   When the light and dark modes used different shared tokens (e.g. light X, dark:Y), create one component variable
   whose value is var(--X) in :root and var(--Y) in .dark.
3. Remove the component's exported color map and its import/spread in preset.ts. Remove theme color entries
   in preset.ts only if no other component, bh-shared-ui, or cmd/ui uses them.
4. Rewrite the styles to use arbitrary-value classes with a type hint, e.g. bg-[color:var(--switch-track)],
   border-[color:var(--x)], text-[color:var(--x)]. Remove every dark: color class. Where a variant and a color
   combine, prefer local variables set by the color variant (see Badge) over per-combination compoundVariants.
5. Keep shared semantic theme classes (primary, secondary, text-*, status-*, neutral-*) as they are.

Rules:
- Visual output must not change in light or dark mode. Before swapping a class, confirm it actually generates CSS.
  Known trap: the theme key 'text-main' produces text-text-main, so `text-main` is not a real class and only inherits.
  Preserve the existing rendering and report such cases instead of silently fixing them.
- Follow the repo's AGENTS.md. Match existing code style and comment density; keep license headers.
- Do not add dependencies.
- Search packages/javascript/bh-shared-ui and cmd/ui for any removed class names and report usages.

Verify (in packages/javascript/doodle-ui):
- npx tsc --noEmit -p . ; npx eslint src ; npx prettier --check src ; yarn vitest run
- Update existing tests that assert removed classes (e.g. dark:... or hex arbitrary values).
- Build CSS: npx tailwindcss -i ./src/input.css -o /tmp/doodle.css. Confirm every new arbitrary class is generated
  and every --<component>-* variable is defined in both :root and .dark.
- For every variant combination, confirm each var(--...) used is either set locally on the element or
  defined in both themes.
- Run `just prepare-for-codereview` before committing.

Finish by appending a short entry to COLOR_TOKEN_EXPERIMENT.md under "Migration log": the components migrated,
any classes removed from the shared theme, and any rendering quirks found and preserved.
```

## Migration log

-   **Badge:** reference implementation. Removed the `badge-*` theme colors. The grey outline icon keeps inherited color (`text-inherit`) because the original `text-main` class did not exist.
