# Example 03: Component Library

## Overview

This example showcases production-ready UI components with real-world complexity. These components demonstrate advanced BEM patterns, multiple size variants, semantic color schemes, and interactive states.

## CSS Input

**Files:**
- `input/avatar.css` - User profile pictures with sizes, shapes, and status indicators
- `input/badge.css` - Status badges with semantic colors and styles
- `input/alert.css` - Notification boxes with icons, actions, and variants
- `input/modal.css` - Dialog overlays with animations and multiple sizes

### Components Overview

#### Avatar Component
- **Base:** Profile picture container with circular shape
- **Elements:** `__img`, `__initials`, `__status`
- **Sizes:** `--xs`, `--sm`, `--md`, `--lg`, `--xl`
- **Shapes:** `--square`, `--rounded`
- **Status:** `__status--online`, `__status--busy`, `__status--away`

#### Badge Component
- **Base:** Small status indicator or label
- **Elements:** `__dot` (visual indicator)
- **Variants:** `--success`, `--warning`, `--error`, `--info`, `--neutral`
- **Sizes:** `--sm`, `--lg`
- **Styles:** `--outline`, `--pill`

#### Alert Component
- **Base:** Notification message box
- **Elements:** `__icon`, `__content`, `__title`, `__description`, `__actions`, `__close`
- **Types:** `--success`, `--warning`, `--error`, `--info`
- **Styles:** `--outline`

#### Modal Component
- **Base:** Full-screen overlay backdrop
- **Elements:** `__dialog`, `__header`, `__title`, `__close`, `__body`, `__footer`
- **Sizes:** `__dialog--sm`, `__dialog--lg`, `__dialog--xl`, `__dialog--full`
- **States:** `--hidden`, `--closing`

## Generated Output

**Files:**
- `output/styles.gen.go` - Main registry
- `output/styles_avatar.gen.go` - Avatar constants (15+ classes)
- `output/styles_badge.gen.go` - Badge constants
- `output/styles_alert.gen.go` - Alert constants
- `output/styles_modal.gen.go` - Modal constants

Each generated constant includes:
- Property categorization (visual, layout, typography, effects)
- Actual CSS values for reference
- BEM relationships and usage context
- Pseudo-state information (:hover, animations)
- Override information for modifiers

## Key Takeaways

1. **Complex Component Patterns** - Real components have many variants and states
2. **Semantic Naming** - Color variants use semantic names (success, error) not colors (green, red)
3. **Size Systems** - Consistent sizing scales (xs, sm, md, lg, xl) across components
4. **Element Composition** - Complex components compose multiple child elements
5. **State Management** - Interactive states (hover, focus) documented in constants

## Usage Examples

### Avatar with Status
```go
templ UserAvatar(user User) {
  <div class={ ui.Avatar, ui.AvatarLg }>
    <img class={ ui.AvatarImg } src={ user.AvatarURL } alt={ user.Name }/>
    <span class={ ui.AvatarStatus, ui.AvatarStatusOnline }></span>
  </div>
}
```

### Badge Variants
```go
templ StatusBadge(status string) {
  <span class={ 
    ui.Badge, 
    templ.KV(ui.BadgeSuccess, status == "active"),
    templ.KV(ui.BadgeWarning, status == "pending"),
    templ.KV(ui.BadgeError, status == "failed"),
  }>
    <span class={ ui.BadgeDot }></span>
    { status }
  </span>
}
```

### Alert with Actions
```go
templ SuccessAlert(message string, onClose func()) {
  <div class={ ui.Alert, ui.AlertSuccess }>
    <svg class={ ui.AlertIcon }>...</svg>
    <div class={ ui.AlertContent }>
      <p class={ ui.AlertTitle }>Success</p>
      <p class={ ui.AlertDescription }>{ message }</p>
      <div class={ ui.AlertActions }>
        <button onclick={ onClose }>Dismiss</button>
      </div>
    </div>
    <button class={ ui.AlertClose } onclick={ onClose }>×</button>
  </div>
}
```

### Modal Dialog
```go
templ ConfirmModal(open bool, onConfirm, onCancel func()) {
  <div class={ ui.Modal, templ.KV(ui.ModalHidden, !open) }>
    <div class={ ui.ModalDialog, ui.ModalDialogSm }>
      <div class={ ui.ModalHeader }>
        <h3 class={ ui.ModalTitle }>Confirm Action</h3>
        <button class={ ui.ModalClose } onclick={ onCancel }>×</button>
      </div>
      <div class={ ui.ModalBody }>
        <p>Are you sure you want to proceed?</p>
      </div>
      <div class={ ui.ModalFooter }>
        <button onclick={ onCancel }>Cancel</button>
        <button onclick={ onConfirm }>Confirm</button>
      </div>
    </div>
  </div>
}
```

## Design Patterns

### Semantic Colors
Instead of specific colors, use intent-based names:
- `--success` (green) - Positive actions/states
- `--warning` (yellow) - Caution/attention needed
- `--error` (red) - Errors/destructive actions
- `--info` (blue) - Informational messages
- `--neutral` (gray) - Default/inactive states

### Size Scales
Consistent sizing across components:
- `--xs` - Extra small (1.5rem)
- `--sm` - Small (2rem)
- `--md` - Medium/default (2.5rem)
- `--lg` - Large (3rem)
- `--xl` - Extra large (4rem)

### Element Naming
Elements describe function, not appearance:
- `__header`, `__body`, `__footer` - Sections
- `__title`, `__description` - Content
- `__icon`, `__close` - Interactive elements
- `__actions` - Button containers

## Regenerating

```bash
# Using config file (from this directory — reads .cssgen.yaml automatically)
cssgen

# Using CLI flags (from this directory)
cssgen generate --source ./input --output-dir ./output --package ui --include "**/*.css"

# From the project root
cssgen generate --source ./examples/03-component-library/input \
        --output-dir ./examples/03-component-library/output \
        --package ui --include "**/*.css"
```

## Next Steps

- **04-css-layers** - Learn about CSS cascade layers for organizing styles
- **05-utility-first** - See utility class patterns
