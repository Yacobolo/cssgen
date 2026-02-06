<h1 align="center">cssgen</h1>

<p align="center">
  <strong>Type-safe CSS class constants for Go/templ projects.</strong>
  <br>
  <a href="https://pkg.go.dev/github.com/yacobolo/cssgen">
    <img src="https://img.shields.io/badge/go-reference-007d9c?logo=go&logoColor=white&style=flat-square" alt="Go Reference">
  </a>
  <a href="https://goreportcard.com/report/github.com/yacobolo/cssgen">
    <img src="https://goreportcard.com/badge/github.com/yacobolo/cssgen?style=flat-square" alt="Go Report Card">
  </a>
  <a href="https://opensource.org/licenses/MIT">
    <img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square" alt="License: MIT">
  </a>
  <a href="https://github.com/yacobolo/cssgen/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/yacobolo/cssgen/ci.yml?branch=main&style=flat-square&label=CI" alt="CI Status">
  </a>
</p>

<p align="center">
  <img src="assets/mascot.png" alt="cssgen mascot - a Go gopher in a construction outfit operating a CSS-to-Go conversion factory" width="600">
</p>

---

`cssgen` generates Go constants from your CSS files and provides a linter to eliminate hardcoded class strings and catch typos at build time.

## Table of Contents
- [Why cssgen?](#why-cssgen)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Examples](#examples)
- [Generated Output](#generated-output)
- [Linting Philosophy](#linting-philosophy)
- [Output Formats](#output-formats)
- [Usage Examples](#usage-examples)
- [How It Works](#how-it-works)
- [Configuration](#configuration)
- [FAQ](#faq)
- [License](#license)
- [Contributing](#contributing)

## Why cssgen?

In modern Go web development with [templ](https://templ.guide), CSS class names are just strings. This creates two problems:

1. **Typos are runtime errors** - `class="btn btn--primray"` fails silently
2. **No refactoring support** - Renaming `.btn-primary` to `.btn--brand` requires manual find/replace

`cssgen` solves this with **type-safe constants** and **build-time validation**:

```go
// Before: Runtime error waiting to happen
<button class="btn btn--primray">Click</button>

// After: Compile-time safety + IDE autocomplete
<button class={ ui.Btn, ui.BtnBrand }>Click</button>
```

## Features

- ✅ **1:1 CSS-to-Go mapping** - One CSS class = One Go constant
- ✅ **Rich IDE tooltips** - Hover over `ui.BtnBrand` to see CSS properties, layers, and inheritance
- ✅ **Smart linter** - Detects typos (errors) and hardcoded strings (warnings)
- ✅ **Multiple output formats** - Issues, JSON, Markdown reports
- ✅ **Zero runtime overhead** - Pure compile-time tool
- ✅ **Component-based generation** - Splits constants into logical files (buttons, cards, etc.)

## Installation

```bash
go install github.com/yacobolo/cssgen/cmd/cssgen@latest
```

**Requirements:** Go 1.21+

## Quick Start

### 1. Generate Constants

```bash
cssg -source ./web/styles -output-dir ./internal/ui -package ui
```

This scans your CSS files and generates:

```
internal/ui/
├── styles.gen.go              # Main file with AllCSSClasses registry
├── styles_buttons.gen.go      # Button constants
├── styles_cards.gen.go        # Card constants
└── ...                        # Other component files
```

### 2. Use in Templates

```go
import "yourproject/internal/ui"

templ Button(text string) {
    <button class={ ui.Btn, ui.BtnBrand, ui.BtnLg }>
        { text }
    </button>
}
// Produces: <button class="btn btn--brand btn--lg">
```

### 3. Lint Your Code

```bash
# Default: Show errors and warnings (golangci-lint style)
cssg -lint-only

# CI mode: Fail on any issue
cssg -lint-only -strict

# Full report with statistics and Quick Wins
cssg -lint-only -output-format full
```

## Examples

The [examples/](./examples/) directory contains comprehensive examples showing how cssgen transforms CSS into type-safe Go constants:

| Example | Focus | Complexity |
|---------|-------|------------|
| [01-basic](./examples/01-basic/) | Simple button styles with BEM modifiers | ⭐ Beginner |
| [02-bem-methodology](./examples/02-bem-methodology/) | Comprehensive BEM patterns | ⭐⭐ Intermediate |
| [03-component-library](./examples/03-component-library/) | Production-ready UI components | ⭐⭐⭐ Advanced |
| [04-css-layers](./examples/04-css-layers/) | CSS cascade layers (@layer) | ⭐⭐ Intermediate |
| [05-utility-first](./examples/05-utility-first/) | Utility class patterns (Tailwind-style) | ⭐⭐ Intermediate |
| [06-complex-selectors](./examples/06-complex-selectors/) | Advanced CSS selector handling | ⭐⭐⭐ Advanced |

**New to cssgen?** Start with [examples/01-basic](./examples/01-basic/) for a gentle introduction.

Each example includes:
- **Input CSS** - Production-quality, well-commented CSS files
- **Output Go** - Pre-generated constants with rich documentation
- **README** - Detailed explanation of patterns and usage

See the [examples README](./examples/README.md) for a complete guide.

## Generated Output

Each constant includes rich metadata as Go comments:

```go
const BtnBrand = "btn--brand"

// @layer components
//
// **Base:** .btn
// **Context:** Use with .btn for proper styling
// **Overrides:** 2 properties (background, color)
//
// **Visual:**
// - background: `var(--ui-color-brand)` 🎨
// - color: `var(--ui-color-brand-on)` 🎨
//
// **Pseudo-states:** :hover, :focus, :active
```

Your IDE shows this when you hover over `ui.BtnBrand`, giving instant CSS context without leaving your editor.

## Linting Philosophy

### Soft Gate (Default)

- **Errors** → Exit code 1 (typos, invalid classes)
- **Warnings** → Exit code 0 (hardcoded strings that should use constants)

This allows gradual migration without blocking development.

```bash
cssg -lint-only
# ✓ Passes CI if no typos (warnings are informational)
```

### Strict Mode (Enforce Adoption)

```bash
cssg -lint-only -strict
# ✗ Fails CI on any issue (errors OR warnings)
```

Use strict mode once you've migrated critical templates.

## Output Formats

`cssgen` supports five output formats via `-output-format`:

| **Format** | **Best For...** | **Visual Detail** |
|------------|-----------------|-------------------|
| `issues` | Local development | High (Inline code pointers) |
| `summary` | Quick health checks | Low (Aggregated stats only) |
| `full` | Deep-dive audits | Maximum (Everything) |
| `markdown` | PR Comments / CI | Medium (Formatted for web) |
| `json` | Custom Tooling | Machine-readable |

### `issues` (default)

Golangci-lint style - errors and warnings only:

```
internal/web/components/button.templ:12:8: invalid CSS class "btn--primray" (csslint)
    <button class="btn btn--primray">
                   ^
internal/web/components/card.templ:5:8: hardcoded CSS class "card" should use ui.Card constant (csslint)
    <div class="card">
               ^

12 issues (1 errors, 11 warnings):
* csslint: 12

Hint: Run with -output-format full to see statistics and Quick Wins
```

### `summary`

Statistics and Quick Wins only (no individual issues):

```
CSS Constant Usage Statistics
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Total constants:         232
Actually used:           8 (3.4%)
Available for migration: 95 (41.0%)
Completely unused:       129 (55.6%)

Top Migration Opportunities
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. "btn" → ui.Btn (23 occurrences)
2. "card" → ui.Card (18 occurrences)
...
```

### `full`

Everything (issues + statistics + Quick Wins):

```
[All issues listed]

[Statistics summary]

[Quick Wins]
```

### `json`

Machine-readable JSON for tooling integration:

```json
{
  "issues": [...],
  "stats": {...},
  "quickWins": [...]
}
```

### `markdown`

Shareable reports for GitHub issues, wikis, or documentation:

```markdown
# CSS Linting Report

## Summary
- **Total Issues:** 225
- **Errors:** 12
- **Warnings:** 213
...
```

## Usage Examples

### Basic Workflows

```bash
# Generate constants from CSS
cssg

# Generate + lint in one pass
cssg -lint

# Lint only (no generation)
cssg -lint-only

# Quiet mode (exit code only, for pre-commit hooks)
cssg -lint-only -quiet

# Weekly adoption report
cssg -lint-only -output-format summary

# Export Markdown report
cssg -lint-only -output-format markdown > css-report.md
```

### Advanced Options

```bash
# Custom source/output directories
cssg -source ./assets/css -output-dir ./pkg/styles -package styles

# Specific file patterns
cssg -include "components/**/*.css,utilities.css"

# Limit linting scope
cssg -lint-only -lint-paths "internal/views/**/*.templ"

# Limit output (CI performance)
cssg -lint-only -max-issues-per-linter 50
```

### CI Integration

#### GitHub Actions

```yaml
name: Lint
on: [push, pull_request]

jobs:
  css-lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      
      - name: Install cssg
        run: go install github.com/yacobolo/cssgen/cmd/cssgen@latest
      
      - name: Lint CSS classes
        run: cssg -lint-only
```

#### Taskfile (task-runner)

```yaml
# Taskfile.yml
tasks:
  css:gen:
    desc: Generate CSS constants
    cmds:
      - cssg
  
  css:lint:
    desc: Lint CSS usage (fast - issues only)
    cmds:
      - cssg -lint-only
  
  css:report:
    desc: Weekly CSS adoption report
    cmds:
      - cssg -lint-only -output-format summary
  
  check:
    desc: Run all checks (Go + CSS)
    cmds:
      - task: test
      - task: css:lint
      - golangci-lint run
```

#### Makefile

```makefile
.PHONY: css-gen css-lint check

css-gen:
	cssg

css-lint:
	cssg -lint-only

check: test css-lint
	golangci-lint run
```

## How It Works

### Generation Process

1. **Scan** - Find CSS files matching glob patterns
2. **Parse** - Extract classes using native CSS parser ([tdewolff/parse](https://github.com/tdewolff/parse))
3. **Analyze** - Detect BEM patterns, build inheritance tree
4. **Generate** - Write Go constants with rich comments

> **Note:** The parser supports **standard CSS** only. For Tailwind/PostCSS-specific syntax (like `@apply` or nested selectors), ensure your build process outputs standard CSS before running `cssgen`.

### Linting Process

1. **Load** - Parse generated `styles*.gen.go` files to build class registry
2. **Scan** - Find all `class=` attributes in `.templ` and `.go` files
3. **Match** - Check each class against registry (with greedy token matching)
4. **Report** - Output issues in golangci-lint format

### 1:1 Mapping Philosophy

`cssgen` uses **pure 1:1 mapping** between CSS classes and Go constants:

```css
.btn { }           →  const Btn = "btn"
.btn--brand { }    →  const BtnBrand = "btn--brand"
.card__header { }  →  const CardHeader = "card__header"
```

**NOT** joined constants:

```go
// ❌ WRONG - Creates pollution and false positives
const BtnBrand = "btn btn--brand"

// ✅ CORRECT - Pure 1:1 mapping
const Btn = "btn"
const BtnBrand = "btn--brand"
```

This ensures:
- **Zero false positives** - Linter suggestions are always accurate
- **Composability** - Mix and match any classes: `{ ui.Btn, ui.BtnBrand, ui.Disabled }`
- **Clear intent** - Each constant represents exactly one CSS class

### Smart Token Matching

When the linter sees `class="btn btn--brand"`:

1. Check if exact match exists for `"btn btn--brand"` → No
2. Split into tokens: `["btn", "btn--brand"]`
3. Match each token: `btn` → `ui.Btn`, `btn--brand` → `ui.BtnBrand`
4. Suggest: `{ ui.Btn, ui.BtnBrand }`

This produces accurate, predictable suggestions.

## Configuration

### Default Behavior

Without flags, `cssgen` uses these defaults:

- **Source:** `web/ui/src/styles`
- **Output:** `internal/web/ui`
- **Package:** `ui`
- **Includes:** `layers/components/**/*.css`, `layers/utilities.css`, `layers/base.css`
- **Lint paths:** `internal/web/features/**/*.{templ,go}`

### Common Flags

**Generation:**
- `-source DIR` - CSS source directory
- `-output-dir DIR` - Go output directory
- `-package NAME` - Go package name
- `-include PATTERNS` - Comma-separated glob patterns

**Linting:**
- `-lint` - Run linter after generation
- `-lint-only` - Run linter without generation
- `-lint-paths PATTERNS` - Files to scan
- `-strict` - Exit 1 on any issue (CI mode)

**Output:**
- `-output-format MODE` - `issues` (default), `summary`, `full`, `json`, `markdown`
- `-quiet` - Suppress all output (exit code only)
- `-max-issues-per-linter N` - Limit issues shown
- `-color` - Force color output

Run `cssg -h` for complete flag documentation.

## FAQ

### Why not use a CSS-in-JS library?

`cssgen` works with **existing CSS files** and standard build tools. No runtime overhead, no new syntax to learn, works with any CSS framework.

### What about utility classes like Tailwind?

`cssgen` generates constants for **any** CSS class, including utilities:

```go
const FlexFill = "flex-fill"
const TextCenter = "text-center"
```

Use them just like component classes: `class={ ui.Flex, ui.FlexFill }`

### Why split into multiple files?

Generated files can be large (1000+ constants). Splitting by component improves:
- IDE performance (faster autocomplete)
- Code navigation (logical grouping)
- Readability (buttons.gen.go vs 3000-line styles.gen.go)

### Can I customize the generated code?

The generator supports two formats:
- `markdown` (default) - Rich comments with hierarchies and diffs
- `compact` - Minimal comments for smaller files

Use `-format compact` for a lighter output.

### Does cssgen work with plain Go `html/template`?

The linter currently targets `templ` and Go files. Support for `html/template` could be added - contributions welcome!

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Issues and pull requests welcome at https://github.com/yacobolo/cssgen

---

**Built with:**
- [tdewolff/parse](https://github.com/tdewolff/parse) - CSS parser
- [bmatcuk/doublestar](https://github.com/bmatcuk/doublestar) - Glob matching
- [fatih/color](https://github.com/fatih/color) - Terminal colors
