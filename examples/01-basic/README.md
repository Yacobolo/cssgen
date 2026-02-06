# Example 01: Basic Button Styles

## Overview

This example demonstrates the fundamental CSS-to-Go transformation that cssgen performs. It shows how simple button styles with BEM modifiers are converted into type-safe Go constants with rich documentation.

## CSS Input

**File:** `input/buttons.css`

A simple button component with:
- `.btn` - Base button class with core styles
- `.btn--primary` - Primary color variant
- `.btn--secondary` - Secondary color variant  
- `.btn--large` - Large size modifier
- `.btn--small` - Small size modifier

This follows the BEM (Block Element Modifier) naming pattern where:
- **Block:** `btn` (the component)
- **Modifiers:** `--primary`, `--secondary`, `--large`, `--small` (variations)

## Generated Output

**File:** `output/styles.gen.go`

cssgen generates:
- One Go constant per CSS class (1:1 mapping)
- Rich comments documenting each constant including:
  - Property categories (visual, layout, typography, etc.)
  - Actual CSS property values
  - Pseudo-state information (:hover, :active)
  - BEM relationships and context hints

Example generated constant:
```go
// **Base:** .btn
// **Context:** Use with .btn for proper styling
// **Overrides:** 2 properties (background-color, color)
//
// **Visual:**
// - background-color: `#3b82f6` 🎨
// - color: `white` 🎨
//
// **Pseudo-states:** :hover
const BtnPrimary = "btn--primary"
```

## Key Takeaways

1. **1:1 Mapping** - Each CSS class becomes exactly one Go constant
2. **Type Safety** - No more typos like `"btn--primray"` - the compiler catches errors
3. **Auto-complete** - Your IDE suggests available button classes
4. **Documentation** - Every constant has rich comments explaining usage and properties
5. **BEM Support** - cssgen detects modifier relationships and documents them

## Regenerating

To regenerate the output file from CSS:

```bash
# Using config file (from this directory — reads .cssgen.yaml automatically)
cssgen

# Using CLI flags (from this directory)
cssgen generate --source ./input --output-dir ./output --package ui --include "**/*.css"

# From the project root
cssgen generate --source ./examples/01-basic/input \
        --output-dir ./examples/01-basic/output \
        --package ui --include "**/*.css"
```

## Usage in Go/templ

```go
import "yourproject/ui"

// Type-safe button classes with autocomplete
templ PrimaryButton(text string) {
  <button class={ ui.Btn, ui.BtnPrimary }>
    { text }
  </button>
}

// Produces: <button class="btn btn--primary">Click me</button>
```

## Next Steps

- **02-bem-methodology** - Learn comprehensive BEM patterns with blocks, elements, and modifiers
- **03-component-library** - See real-world component examples
