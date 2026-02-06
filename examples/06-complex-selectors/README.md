# Example 06: Complex CSS Selectors

## Overview

This example demonstrates how cssgen handles various CSS selector types. It shows which selectors generate Go constants and which are documented in other ways (like pseudo-states in comments).

## CSS Input

**File:** `input/advanced.css`

A comprehensive CSS file showing:
- Simple class selectors
- Pseudo-classes (:hover, :focus, :active)
- Pseudo-elements (::before, ::after)
- Combinators (descendant, child, sibling)
- Attribute selectors
- Complex combinations
- Media queries
- Special cases

## What Gets Extracted

### ✓ Simple Class Selectors

**CSS:**
```css
.button { padding: 0.5rem 1rem; }
```

**Generated:**
```go
const Button = "button"
```

**Result:** Direct 1:1 mapping.

---

### ✓ Pseudo-Classes (Documented)

**CSS:**
```css
.link { color: #3b82f6; }
.link:hover { color: #2563eb; }
.link:focus { outline: 2px solid #3b82f6; }
```

**Generated:**
```go
// **Visual:**
// - color: `#3b82f6`
// - text-decoration: `none`
//
// **Interactions:**
// - `:hover`: Changes color to `#2563eb`, text-decoration to `underline`
// - `:focus`: Changes outline to `2px solid #3b82f6`, outline-offset to `2px`
const Link = "link"
```

**Result:** Base class extracted, pseudo-states documented in comments.

---

### ✓ Pseudo-Elements (Documented)

**CSS:**
```css
.quote { padding-left: 1.5rem; }
.quote::before { content: '"'; }
.quote::after { content: '"'; }
```

**Generated:**
```go
// **Layout:**
// - padding-left: `1.5rem`
// - position: `relative`
//
// Note: Uses ::before and ::after pseudo-elements
const Quote = "quote"
```

**Result:** Base class extracted, pseudo-elements mentioned.

---

### ✓ Descendant Combinators (Class Parts)

**CSS:**
```css
.card .card__header .card__title {
  font-size: 1.25rem;
}
```

**Generated:**
```go
const Card = "card"
const CardHeader = "card__header"
const CardTitle = "card__title"
```

**Result:** All class parts extracted separately. Combinator relationship preserved in BEM naming.

---

### ✓ Child Combinator (>)

**CSS:**
```css
.menu > .menu__item {
  margin-bottom: 0.5rem;
}
```

**Generated:**
```go
const Menu = "menu"
const MenuItem = "menu__item"
```

**Result:** Both classes extracted independently.

---

### ✓ Sibling Combinator (+, ~)

**CSS:**
```css
.label + .input {
  margin-top: 0.25rem;
}
```

**Generated:**
```go
const Label = "label"
const Input = "input"
```

**Result:** Both classes extracted separately.

---

### ✓ Media Queries (Classes Extracted)

**CSS:**
```css
.responsive { display: block; }

@media (min-width: 768px) {
  .responsive { display: flex; }
}
```

**Generated:**
```go
// **Layout:**
// - display: `block`
//
// Note: Has responsive styles in media queries
const Responsive = "responsive"
```

**Result:** Class extracted once, media query noted.

---

## What Doesn't Get Extracted

### ✗ Attribute Selectors

**CSS:**
```css
[data-theme="dark"] { background-color: #111827; }
input[type="text"] { border: 1px solid #d1d5db; }
```

**Result:** No constants generated. Attribute selectors are not class-based.

---

### ✗ Element Selectors

**CSS:**
```css
body { margin: 0; }
h1 { font-size: 2rem; }
```

**Result:** No constants. cssgen only extracts class selectors.

---

### ✗ ID Selectors

**CSS:**
```css
#header { height: 4rem; }
```

**Result:** Not extracted. IDs are discouraged in favor of classes.

---

### ✗ Universal Selector

**CSS:**
```css
* { box-sizing: border-box; }
```

**Result:** No constant. Universal selector has no specific name.

---

## Generated Output

**File:** `output/styles.gen.go`

cssgen extracts class selectors and documents their context:

- **Simple classes** → Direct constants
- **Pseudo-classes** → Documented in "Interactions" section
- **Pseudo-elements** → Noted in comments
- **Combined selectors** → Class parts extracted individually
- **Media queries** → Noted as "responsive styles"
- **BEM patterns** → Relationships documented in "Base" and "Context"

## Key Takeaways

1. **Class-Only Extraction** - cssgen only generates constants for class selectors
2. **Pseudo-State Documentation** - :hover, :focus, etc. documented in generated comments
3. **Combinator Handling** - Class parts from complex selectors extracted separately
4. **Context Preservation** - BEM relationships and usage hints included
5. **Focused Scope** - No constants for elements, IDs, or attributes

## Usage Patterns

### Simple Classes
```go
templ Button() {
  <button class={ ui.Button }>Click</button>
}
```

### Pseudo-States (Handled by CSS)
```go
templ Link(href string) {
  // :hover and :focus handled automatically by CSS
  <a class={ ui.Link } href={ href }>Link</a>
}
```

### BEM Components
```go
templ Card() {
  <div class={ ui.Card }>
    <div class={ ui.CardHeader }>
      <h3 class={ ui.CardTitle }>Title</h3>
    </div>
  </div>
}
```

### Complex Combinations
```go
templ Dropdown(open bool) {
  <div class={ ui.Dropdown, templ.KV("is-open", open) }>
    <div class={ ui.DropdownMenu }>Items</div>
  </div>
}
// Note: State classes like "is-open" may not have constants
// if they're only used in combinators
```

### Internal Classes
```go
// Classes starting with _ are marked as internal
// They generate constants but are documented as internal-use
templ Component() {
  <div class={ ui.Internal }>Hidden utility</div>
}
```

## Understanding cssgen's Extraction Rules

### Always Extracted
- ✅ `.classname` - Simple class selector
- ✅ `.class1, .class2` - Multiple selectors
- ✅ `.parent .child` - Descendant (both classes)
- ✅ `.parent > .child` - Child combinator (both classes)
- ✅ `.prev + .next` - Adjacent sibling (both classes)

### Documented, Not Extracted
- 📝 `:hover`, `:focus`, `:active` - Pseudo-classes
- 📝 `::before`, `::after` - Pseudo-elements
- 📝 `@media` queries - Responsive styles

### Never Extracted
- ❌ `[attribute]` - Attribute selectors
- ❌ `element` - Element selectors
- ❌ `#id` - ID selectors
- ❌ `*` - Universal selector

## Best Practices

1. **Use Classes** - Prefer class selectors for all styled elements
2. **Avoid IDs** - Use classes instead of IDs for styling
3. **BEM Naming** - Use BEM for component relationships
4. **State Classes** - Use explicit state classes (`.is-open`, `.is-active`)
5. **Trust CSS** - Pseudo-states work automatically, no special handling needed

## Regenerating

```bash
# Using config file (from this directory — reads .cssgen.yaml automatically)
cssgen

# Using CLI flags (from this directory)
cssgen generate --source ./input --output-dir ./output --package ui --include "**/*.css"

# From the project root
cssgen generate --source ./examples/06-complex-selectors/input \
        --output-dir ./examples/06-complex-selectors/output \
        --package ui --include "**/*.css"
```

## Next Steps

- **01-basic** - Return to simple patterns
- **02-bem-methodology** - Deep dive into BEM
- **examples/README.md** - Overview of all examples
