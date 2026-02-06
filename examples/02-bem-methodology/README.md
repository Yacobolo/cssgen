# Example 02: BEM Methodology

## Overview

This example demonstrates comprehensive BEM (Block Element Modifier) patterns across multiple components. It shows how cssgen detects and documents the relationships between blocks, elements, and modifiers.

## CSS Input

**Files:**
- `input/card.css` - Card component with header, body, footer elements
- `input/button.css` - Button component with icon and text elements
- `input/navigation.css` - Navigation with complex item/link/badge structure

### BEM Pattern Explanation

**BEM** is a naming methodology that makes CSS more maintainable:
- **Block:** Standalone component (`.card`, `.btn`, `.nav`)
- **Element:** Part of a block (`.card__header`, `.btn__icon`, `.nav__link`)
- **Modifier:** Variation of block/element (`.card--featured`, `.btn--primary`, `.nav__item--active`)

**Naming convention:**
```
.block              /* Component */
.block__element     /* Part of component */
.block--modifier    /* Variation of component */
.block__element--modifier  /* Variation of element */
```

## Generated Output

**Files:**
- `output/styles.gen.go` - Main registry with all class mappings
- `output/styles_card.gen.go` - Card component constants
- `output/styles_button.gen.go` - Button component constants
- `output/styles_navigation.gen.go` - Navigation component constants

cssgen detects BEM relationships and adds context to generated constants:

```go
// Card block (base component)
const Card = "card"

// Card element - cssgen documents this is part of .card
// **Base:** .card
// **Context:** This is an element of .card
const CardHeader = "card__header"

// Card modifier - cssgen shows this modifies .card
// **Base:** .card  
// **Context:** Use with .card for proper styling
const CardFeatured = "card--featured"
```

## Key Takeaways

1. **Inheritance Detection** - cssgen automatically identifies BEM relationships
2. **Context Hints** - Generated comments explain when to use modifiers with base classes
3. **File Organization** - Each component gets its own generated file for clarity
4. **Element Documentation** - Element classes document which block they belong to
5. **Modifier Overrides** - Modifier constants list which properties they override

## BEM in Practice

### Card Component Usage
```go
// Base card
templ BasicCard() {
  <div class={ ui.Card }>
    <div class={ ui.CardHeader }>Title</div>
    <div class={ ui.CardBody }>Content</div>
    <div class={ ui.CardFooter }>Actions</div>
  </div>
}

// Featured card (combines base + modifier)
templ FeaturedCard() {
  <div class={ ui.Card, ui.CardFeatured }>
    <div class={ ui.CardHeader }>Featured!</div>
    <div class={ ui.CardBody }>Special content</div>
  </div>
}
```

### Button Component Usage
```go
templ IconButton(icon, text string) {
  <button class={ ui.Btn, ui.BtnPrimary }>
    <svg class={ ui.BtnIcon }>...</svg>
    <span class={ ui.BtnText }>{ text }</span>
  </button>
}
```

### Navigation Component Usage
```go
templ VerticalNav(items []NavItem) {
  <nav class={ ui.Nav, ui.NavVertical }>
    for _, item := range items {
      <div class={ ui.NavItem, templ.KV(ui.NavItemActive, item.Active) }>
        <a class={ ui.NavLink } href={ item.URL }>{ item.Label }</a>
        if item.Count > 0 {
          <span class={ ui.NavBadge }>{ strconv.Itoa(item.Count) }</span>
        }
      </div>
    }
  </nav>
}
```

## Regenerating

```bash
# Using config file (from this directory — reads .cssgen.yaml automatically)
cssgen

# Using CLI flags (from this directory)
cssgen generate --source ./input --output-dir ./output --package ui --include "**/*.css"

# From the project root
cssgen generate --source ./examples/02-bem-methodology/input \
        --output-dir ./examples/02-bem-methodology/output \
        --package ui --include "**/*.css"
```

## Next Steps

- **03-component-library** - See production-ready component examples
- **04-css-layers** - Learn about CSS cascade layers
