# Example 05: Utility-First CSS

## Overview

This example demonstrates a utility-first CSS approach (similar to Tailwind CSS) where single-purpose utility classes are composed to build designs. This shows how cssgen handles large numbers of utility classes and categorizes them effectively.

## CSS Input

**Files:**
- `input/spacing.css` - Margin, padding, and gap utilities (50+ classes)
- `input/typography.css` - Text styling utilities (60+ classes)
- `input/colors.css` - Color utilities for text, background, border (60+ classes)
- `input/layout.css` - Display, flexbox, grid, positioning (80+ classes)

### Utility-First Approach

**Philosophy:** Build designs by composing small, single-purpose utility classes instead of writing custom CSS.

**Traditional CSS:**
```css
.user-card {
  display: flex;
  padding: 1rem;
  background-color: white;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}
```

**Utility-First:**
```html
<div class="flex p-4 bg-white rounded-lg shadow">
  <!-- No custom CSS needed -->
</div>
```

**With cssgen:**
```go
<div class={ ui.Flex, ui.P4, ui.BgWhite, ui.RoundedLg, ui.Shadow }>
  <!-- Type-safe + autocomplete -->
</div>
```

### Utility Categories

#### Spacing (spacing.css)
- **Margin:** `.m-*`, `.mt-*`, `.mb-*`, `.mx-*`
- **Padding:** `.p-*`, `.px-*`, `.py-*`
- **Gap:** `.gap-*`
- **Scale:** 0, 1 (0.25rem), 2 (0.5rem), 4 (1rem), 8 (2rem), 16 (4rem)

#### Typography (typography.css)
- **Size:** `.text-xs` to `.text-4xl`
- **Weight:** `.font-thin` to `.font-black`
- **Alignment:** `.text-left`, `.text-center`, `.text-right`
- **Transform:** `.uppercase`, `.lowercase`, `.capitalize`
- **Decoration:** `.underline`, `.line-through`, `.no-underline`
- **Line height:** `.leading-tight` to `.leading-loose`
- **Spacing:** `.tracking-tight`, `.tracking-wide`
- **Utilities:** `.truncate`, `.whitespace-nowrap`

#### Colors (colors.css)
- **Text colors:** `.text-gray-*`, `.text-blue-*`, `.text-red-*`, etc.
- **Backgrounds:** `.bg-gray-*`, `.bg-blue-*`, `.bg-transparent`
- **Borders:** `.border-gray-*`, `.border-blue-*`
- **Border styles:** `.border-solid`, `.border-dashed`, `.border-none`
- **Border width:** `.border`, `.border-0`, `.border-2`, `.border-4`

#### Layout (layout.css)
- **Display:** `.block`, `.flex`, `.grid`, `.hidden`
- **Flex direction:** `.flex-row`, `.flex-col`
- **Flex wrap:** `.flex-wrap`, `.flex-nowrap`
- **Justify:** `.justify-start`, `.justify-center`, `.justify-between`
- **Align:** `.items-start`, `.items-center`, `.items-stretch`
- **Grid:** `.grid-cols-1` to `.grid-cols-4`
- **Width:** `.w-full`, `.w-auto`, `.w-1-2`, `.w-screen`
- **Height:** `.h-full`, `.h-auto`, `.h-screen`
- **Position:** `.relative`, `.absolute`, `.fixed`, `.sticky`
- **Inset:** `.inset-0`, `.top-0`, `.right-0`
- **Z-index:** `.z-0`, `.z-10`, `.z-50`
- **Overflow:** `.overflow-auto`, `.overflow-hidden`

## Generated Output

**File:** `output/styles.gen.go`

cssgen generates 250+ constants with property categorization:

```go
// **Layout:**
// - display: `flex`
const Flex = "flex"

// **Layout:**
// - padding: `1rem`
const P4 = "p-4"

// **Visual:**
// - background-color: `#ffffff`
const BgWhite = "bg-white"

// **Typography:**
// - font-size: `1.125rem`
// - line-height: `1.75rem`
const TextLg = "text-lg"
```

The property category annotations help understand what each utility affects.

## Key Takeaways

1. **Large Scale** - cssgen efficiently handles hundreds of utility classes
2. **Property Categories** - Generated comments show which properties each utility affects
3. **Systematic Naming** - Consistent patterns make utilities predictable
4. **Single Responsibility** - Each utility does one thing well
5. **Composition** - Complex designs built from simple utilities

## Usage Examples

### Card Component
```go
templ UserCard(user User) {
  <div class={ 
    ui.Flex, ui.FlexCol, ui.Gap4,
    ui.P6, ui.BgWhite,
    ui.BorderSolid, ui.Border, ui.BorderGray200,
    ui.RoundedLg,
  }>
    <div class={ ui.Flex, ui.ItemsCenter, ui.Gap3 }>
      <img class={ ui.W12, ui.H12, ui.RoundedFull } src={ user.Avatar }/>
      <div>
        <h3 class={ ui.TextLg, ui.FontSemibold }>{ user.Name }</h3>
        <p class={ ui.TextSm, ui.TextGray500 }>{ user.Email }</p>
      </div>
    </div>
    <p class={ ui.TextBase, ui.TextGray700 }>{ user.Bio }</p>
  </div>
}
```

### Alert Banner
```go
templ SuccessBanner(message string) {
  <div class={
    ui.Flex, ui.ItemsCenter, ui.Gap2,
    ui.Px4, ui.Py3,
    ui.BgGreen100, ui.BorderGreen500,
    ui.BorderSolid, ui.Border2,
    ui.RoundedMd,
  }>
    <span class={ ui.TextGreen600, ui.FontSemibold }>✓</span>
    <p class={ ui.TextGreen700 }>{ message }</p>
  </div>
}
```

### Responsive Grid
```go
templ ProductGrid(products []Product) {
  <div class={
    ui.Grid,
    ui.GridCols1,  // 1 column on mobile
    ui.SmGridCols2, // 2 columns on tablet (if responsive variants exist)
    ui.LgGridCols4, // 4 columns on desktop
    ui.Gap6,
  }>
    for _, product := range products {
      @ProductCard(product)
    }
  </div>
}
```

### Centered Hero Section
```go
templ Hero() {
  <section class={
    ui.Flex, ui.FlexCol,
    ui.ItemsCenter, ui.JustifyCenter,
    ui.MinHScreen,
    ui.BgGray900,
    ui.Px4,
  }>
    <h1 class={
      ui.Text4xl, ui.FontBold, ui.TextWhite,
      ui.TextCenter, ui.Mb4,
    }>Welcome to Our Product</h1>
    <p class={
      ui.TextLg, ui.TextGray300,
      ui.TextCenter, ui.MaxWProse,
    }>Build amazing things with type-safe CSS</p>
  </section>
}
```

## Utility-First Best Practices

### When to Use Utilities
✅ **Good for:**
- Layout and spacing
- Typography styles
- Colors and backgrounds
- Common patterns (flex centering, truncation)
- Prototyping and rapid development

❌ **Not ideal for:**
- Complex animations
- Component-specific interactions
- Heavily repeated combinations (extract to component)

### Naming Conventions
- **Scale-based:** Use consistent scales (1, 2, 4, 8, 16)
- **Abbreviated:** Short names for common use (`.p-4`, `.mt-2`)
- **Directional:** Use prefixes (`.px-*` = padding-x, `.mt-*` = margin-top)
- **Semantic:** Color names indicate purpose (gray, blue, red)

### Composition Patterns
```go
// Define common combinations as Go variables
var (
  CardBase = []string{ ui.BgWhite, ui.RoundedLg, ui.P6, ui.Shadow }
  FlexCenter = []string{ ui.Flex, ui.ItemsCenter, ui.JustifyCenter }
  SectionSpacing = []string{ ui.Py16, ui.Px4 }
)

templ Card() {
  <div class={ CardBase... }>Content</div>
}
```

## Regenerating

```bash
# From project root
cssgen -source ./examples/05-utility-first/input \
        -output-dir ./examples/05-utility-first/output \
        -package ui \
        -include "**/*.css"

# From this directory
cssgen -source ./input -output-dir ./output -package ui -include "**/*.css"
```

## Next Steps

- **06-complex-selectors** - Learn how cssgen handles advanced CSS selectors
- **01-basic** - Review simple BEM component patterns
