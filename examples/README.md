# cssgen Examples

This directory contains comprehensive examples demonstrating how cssgen transforms CSS into type-safe Go constants. Each example focuses on a specific CSS pattern or use case.

## Quick Start

**New to cssgen?** Start with [01-basic](./01-basic/) for a simple introduction.

**Want to learn BEM?** Check out [02-bem-methodology](./02-bem-methodology/).

**Building a component library?** See [03-component-library](./03-component-library/).

## Examples Overview

| Example | Focus | Classes | Complexity |
|---------|-------|---------|------------|
| [01-basic](./01-basic/) | Simple button styles with BEM modifiers | 5 | ⭐ Beginner |
| [02-bem-methodology](./02-bem-methodology/) | Comprehensive BEM patterns across components | 20 | ⭐⭐ Intermediate |
| [03-component-library](./03-component-library/) | Production-ready UI components | 50 | ⭐⭐⭐ Advanced |
| [04-css-layers](./04-css-layers/) | CSS cascade layers (@layer directive) | 40 | ⭐⭐ Intermediate |
| [05-utility-first](./05-utility-first/) | Utility class patterns (Tailwind-style) | 186 | ⭐⭐ Intermediate |
| [06-complex-selectors](./06-complex-selectors/) | Advanced CSS selector handling | 31 | ⭐⭐⭐ Advanced |

## Learning Path

### 1. Fundamentals
Start here to understand cssgen basics:
- **01-basic** - Learn the 1:1 CSS-to-Go mapping
- **06-complex-selectors** - Understand what gets extracted and how

### 2. Component Patterns
Learn how to structure CSS for components:
- **02-bem-methodology** - Master BEM naming conventions
- **03-component-library** - See production patterns in action

### 3. Organizational Approaches
Explore different CSS organization strategies:
- **04-css-layers** - Modern CSS with cascade layers
- **05-utility-first** - Utility-first CSS approach

## What's Inside Each Example

Each example directory contains:

```
example-name/
├── .cssgen.yaml       # Config file — run `cssgen` from here to regenerate
├── README.md          # Detailed explanation of the example
├── input/             # CSS source files
│   └── *.css         # Well-commented CSS demonstrating patterns
└── output/            # Generated Go files (pre-generated)
    └── *.gen.go      # Type-safe constants with rich documentation
```

### Input Directory
Contains CSS files showing various patterns. These are:
- **Production-quality** - Realistic, well-structured CSS
- **Heavily commented** - Explains design decisions
- **Best practices** - Follows modern CSS conventions

### Output Directory
Contains pre-generated Go files so you can:
- **Browse immediately** - No need to run cssgen first
- **See generated output** - Understand what cssgen produces
- **Learn patterns** - Study the generated comments and structure

## How to Use These Examples

### 1. Browse and Learn
Simply read the input CSS and generated output Go files. Each example's README explains the patterns and concepts.

### 2. Regenerate Output
Try regenerating the output yourself:

```bash
# From an example directory (each has a .cssgen.yaml config file)
cssgen

# Or using explicit CLI flags
cssgen generate --source ./input --output-dir ./output --package ui --include "**/*.css"

# From project root
cssgen generate --source ./examples/01-basic/input \
        --output-dir ./examples/01-basic/output \
        --package ui --include "**/*.css"
```

### 3. Experiment
Modify the CSS files and regenerate to see how changes affect the output:
- Add new classes
- Change property values
- Try different BEM patterns
- Add pseudo-classes

### 4. Copy Patterns
Use these examples as templates for your own projects:
- Copy CSS structure
- Adopt naming conventions
- Replicate component patterns

## Key Concepts Demonstrated

### 1:1 Mapping (01-basic)
Every CSS class becomes exactly one Go constant:
```css
.btn--primary { background: blue; }
```
```go
const BtnPrimary = "btn--primary"
```

### BEM Patterns (02-bem-methodology)
Block, Element, Modifier methodology:
```css
.card                    /* Block */
.card__header            /* Element */
.card--featured          /* Modifier */
```

### Component Architecture (03-component-library)
Production components with variants:
```css
.avatar                  /* Base component */
.avatar--lg              /* Size variant */
.avatar__status--online  /* Element modifier */
```

### CSS Layers (04-css-layers)
Modern cascade control:
```css
@layer base { .container { ... } }
@layer components { .btn { ... } }
@layer utilities { .hidden { ... } }
```

### Utility-First (05-utility-first)
Single-purpose utilities:
```css
.flex { display: flex; }
.p-4 { padding: 1rem; }
.text-center { text-align: center; }
```

### Selector Extraction (06-complex-selectors)
What cssgen extracts:
- ✅ Class selectors → Constants
- ✅ Pseudo-classes → Documented
- ✅ Combinators → Class parts extracted
- ❌ Attributes → Not extracted
- ❌ Elements → Not extracted

## Generated Output Features

cssgen generates Go files with rich documentation:

### Property Categorization
```go
// **Visual:**
// - background-color: `#3b82f6`
// - color: `white`
//
// **Layout:**
// - padding: `0.625rem 1.25rem`
//
// **Typography:**
// - font-weight: `500`
```

### BEM Relationships
```go
// **Base:** .btn
// **Context:** Use with .btn for proper styling
// **Overrides:** 2 properties (background-color, color)
const BtnPrimary = "btn--primary"
```

### Pseudo-State Documentation
```go
// **Interactions:**
// - `:hover`: Changes background-color to `#2563eb`
// - `:focus`: Changes box-shadow to `0 0 0 3px rgba(59, 130, 246, 0.1)`
const Button = "button"
```

### Layer Annotations
```go
// @layer components
//
// **Visual:**
// - background-color: `white`
const Card = "card"
```

## Common Patterns Index

Looking for a specific pattern? Find it here:

### Component Patterns
- **Simple button** → 01-basic/input/buttons.css
- **Card component** → 02-bem-methodology/input/card.css
- **Navigation** → 02-bem-methodology/input/navigation.css
- **Avatar** → 03-component-library/input/avatar.css
- **Badge** → 03-component-library/input/badge.css
- **Alert** → 03-component-library/input/alert.css
- **Modal** → 03-component-library/input/modal.css

### CSS Organization
- **Layer-based** → 04-css-layers/input/
- **Utility-first** → 05-utility-first/input/

### Selector Techniques
- **Pseudo-classes** → 06-complex-selectors (`:hover`, `:focus`)
- **Pseudo-elements** → 06-complex-selectors (`::before`, `::after`)
- **Combinators** → 06-complex-selectors (descendant, child)

### Naming Conventions
- **BEM blocks** → 02-bem-methodology
- **BEM elements** → 02-bem-methodology
- **BEM modifiers** → 01-basic, 02-bem-methodology
- **Utility naming** → 05-utility-first
- **Semantic colors** → 03-component-library/input/badge.css

### Size Scales
- **T-shirt sizes** → 03-component-library/input/avatar.css (xs, sm, md, lg, xl)
- **Numeric scales** → 05-utility-first/input/spacing.css (0, 1, 2, 4, 8)

## Usage in Go/templ

All examples show how to use generated constants in templ templates:

```go
import "yourproject/ui"

// Simple usage
templ Button() {
  <button class={ ui.Btn, ui.BtnPrimary }>
    Click me
  </button>
}

// Conditional classes
templ Card(featured bool) {
  <div class={ ui.Card, templ.KV(ui.CardFeatured, featured) }>
    Content
  </div>
}

// Utility composition
templ Section() {
  <section class={ 
    ui.Flex, ui.FlexCol, ui.ItemsCenter,
    ui.Py16, ui.Px4,
    ui.BgGray50,
  }>
    Content
  </section>
}
```

## Tips for Your Own Projects

### 1. Start Small
Begin with component-based CSS (like 01-basic or 02-bem-methodology) rather than large utility libraries.

### 2. Choose a Strategy
Pick one organizational approach:
- **Component-focused** → BEM methodology (examples 01-03)
- **Layer-based** → CSS layers (example 04)
- **Utility-first** → Utility classes (example 05)

Don't mix approaches unless you understand the tradeoffs.

### 3. Use Consistent Naming
Pick a naming convention and stick to it:
- **BEM:** `.block__element--modifier`
- **Utilities:** `.property-value` (`.text-lg`, `.p-4`)
- **Semantic:** `.btn-primary`, `.alert-success`

### 4. Leverage Generated Comments
cssgen's generated comments are educational:
- See which properties each class sets
- Understand BEM relationships
- Know when to combine classes

### 5. Run cssgen Often
Regenerate after CSS changes:
```bash
# With a .cssgen.yaml config file in your project root
cssgen

# Or with explicit flags
cssgen generate --source ./styles --output-dir ./ui --package ui --include "**/*.css"
```

Add to your build process or use a file watcher.

## Regenerating All Examples

To regenerate all examples at once:

```bash
# Using config files (each example has .cssgen.yaml)
for dir in examples/*/; do
  (cd "$dir" && cssgen)
done

# Or using explicit CLI flags
for dir in examples/*/; do
  cssgen generate --source "$dir/input" \
          --output-dir "$dir/output" \
          --package ui --include "**/*.css"
done
```

Or run cssgen individually for each example (see each example's README).

## Getting Help

- **Main Documentation:** [../README.md](../README.md)
- **CLI Usage:** [../cmd/cssgen/README.md](../cmd/cssgen/README.md)
- **Contributing:** [../CONTRIBUTING.md](../CONTRIBUTING.md)
- **Issues:** [GitHub Issues](https://github.com/yacobolo/cssgen/issues)

## Next Steps

1. **Read 01-basic** → Understand the basics
2. **Explore other examples** → Find patterns relevant to your project
3. **Try regenerating** → Run cssgen yourself
4. **Experiment** → Modify CSS and see what changes
5. **Apply to your project** → Use these patterns in your own code

Happy coding with type-safe CSS!
