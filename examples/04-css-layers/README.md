# Example 04: CSS Cascade Layers

## Overview

This example demonstrates modern CSS `@layer` directive usage and how cssgen handles it. CSS cascade layers provide explicit control over the cascade, allowing you to organize styles by purpose and control specificity.

## CSS Input

**Files:**
- `input/base.css` - `@layer base` - Resets and foundational styles
- `input/components.css` - `@layer components` - UI component styles
- `input/utilities.css` - `@layer utilities` - Utility classes
- `input/theme.css` - CSS custom properties (design tokens)

### CSS Layers Explained

**CSS Cascade Layers** (`@layer`) allow explicit ordering of CSS by purpose, not source order:

```css
@layer base {
  .btn { /* base button styles */ }
}

@layer utilities {
  .hidden { display: none !important; }
}
```

**Layer Priority** (lowest to highest):
1. `@layer base` - Resets, defaults
2. `@layer components` - Component styles
3. `@layer utilities` - Utility overrides
4. Unlayered styles - Highest priority

**Benefits:**
- **Predictable Cascade** - Explicit specificity control
- **No !important Needed** - Layers handle priority
- **Better Organization** - Styles grouped by purpose
- **Safer Overrides** - Utilities always win

### File Organization

#### base.css - Foundation
- `.reset` - Box model reset
- `.text-base` - Default typography
- `.container` - Content wrapper
- `.link` - Default link styles

#### components.css - UI Components
- `.btn-primary` - Primary button
- `.card`, `.card__title`, `.card__content` - Card component
- `.input`, `.input--error` - Form input

#### utilities.css - Helpers
- Spacing: `.mt-*`, `.mb-*`
- Display: `.hidden`, `.flex`, `.grid`
- Typography: `.text-center`, `.font-bold`
- Colors: `.text-gray-*`, `.text-blue-*`
- Width: `.w-full`, `.w-auto`

#### theme.css - Design Tokens
- Color palette (`--color-*`)
- Spacing scale (`--spacing-*`)
- Border radius (`--radius-*`)
- Shadows (`--shadow-*`)
- Typography (`--font-*`, `--text-*`)
- Theme classes using custom properties

## Generated Output

**File:** `output/styles.gen.go`

cssgen generates constants with `@layer` annotations in comments:

```go
// @layer base
//
// **Layout:**
// - margin-left: `auto`
// - margin-right: `auto`
// - max-width: `72rem`
// ...
const Container = "container"

// @layer components
//
// **Visual:**
// - background-color: `#3b82f6`
// - color: `white`
// ...
const BtnPrimary = "btn-primary"

// @layer utilities
//
// **Layout:**
// - display: `none`
const Hidden = "hidden"
```

The layer annotation helps developers understand cascade priority when combining classes.

## Key Takeaways

1. **Layer Annotations** - cssgen preserves `@layer` information in generated comments
2. **Cascade Control** - Layers provide predictable style priority
3. **Design Tokens** - CSS custom properties work alongside typed constants
4. **Organization** - Layers help organize large stylesheets by purpose
5. **Modern CSS** - cssgen supports contemporary CSS features

## Usage Examples

### Combining Layers
```go
// Utility layer always overrides component layer
templ HiddenCard() {
  // The .hidden utility will override .card display property
  <div class={ ui.Card, ui.Hidden }>
    Never shown
  </div>
}

templ VisibleCard() {
  <div class={ ui.Card }>
    <h3 class={ ui.CardTitle }>Title</h3>
    <p class={ ui.CardContent }>Content</p>
  </div>
}
```

### Base + Components
```go
templ Page() {
  <div class={ ui.Container }>
    <a class={ ui.Link } href="/home">Home</a>
    <button class={ ui.BtnPrimary }>Click me</button>
  </div>
}
```

### Utilities for Layout
```go
templ Form() {
  <form>
    <input class={ ui.Input, ui.WFull, ui.Mb4 } type="text"/>
    <input class={ ui.Input, ui.InputError, ui.WFull, ui.Mb4 } type="email"/>
    <button class={ ui.BtnPrimary, ui.WFull }>Submit</button>
  </form>
}
```

### Theme Classes with Custom Properties
```go
templ ThemedSection() {
  <section class={ ui.ThemePrimary }>
    <h2 class={ ui.TextCenter, ui.FontBold }>Primary Theme</h2>
    <p class={ ui.TextCenter }>Uses var(--color-primary)</p>
  </section>
}
```

## Layer Best Practices

### Organization
1. **Base Layer** - Normalize, resets, element defaults
2. **Component Layer** - Reusable UI components
3. **Utility Layer** - Single-purpose helpers (highest priority)
4. **Unlayered** - Third-party CSS, overrides

### Specificity
- Layers make specificity less important
- Utility layer classes always override component layer
- No need for `!important` when using layers properly

### Migration
- Existing projects can adopt layers incrementally
- Start with utilities in `@layer utilities`
- Gradually move components to `@layer components`
- Keep unlayered styles until migrated

## CSS Custom Properties

While cssgen generates constants for class names, CSS custom properties (variables) complement this approach:

**Class constants** → Type-safe class references  
**CSS variables** → Dynamic theming, runtime changes

```go
// Type-safe classes + CSS variables = powerful combination
templ DynamicCard(primary bool) {
  <div class={ 
    ui.Card,
    templ.KV(ui.ThemePrimary, primary),
    templ.KV(ui.ThemeSurface, !primary),
  }>
    Content adapts to theme via CSS variables
  </div>
}
```

## Regenerating

```bash
# Using config file (from this directory — reads .cssgen.yaml automatically)
cssgen

# Using CLI flags (from this directory)
cssgen generate --source ./input --output-dir ./output --package ui --include "**/*.css"

# From the project root
cssgen generate --source ./examples/04-css-layers/input \
        --output-dir ./examples/04-css-layers/output \
        --package ui --include "**/*.css"
```

## Next Steps

- **05-utility-first** - Comprehensive utility class patterns
- **06-complex-selectors** - Advanced CSS selector handling
