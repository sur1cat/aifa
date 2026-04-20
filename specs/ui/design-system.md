# Design System

## Philosophy

**"Calm Productivity"** — HabitFlow should feel like a peaceful space, not another demanding app. Every interaction should be effortless.

## Color Palette

### Light Mode

| Name | Hex | Usage |
|------|-----|-------|
| Background | `#FFFFFF` | Main background |
| Surface | `#F8F9FA` | Cards, inputs |
| Border | `#DEE2E6` | Dividers, outlines |
| Text Primary | `#212529` | Main text |
| Text Secondary | `#6C757D` | Secondary text |
| Text Tertiary | `#ADB5BD` | Hints, placeholders |

### Dark Mode

| Name | Hex | Usage |
|------|-----|-------|
| Background | `#121212` | Main background |
| Surface | `#1E1E1E` | Cards, inputs |
| Border | `#2D2D2D` | Dividers, outlines |
| Text Primary | `#FFFFFF` | Main text |
| Text Secondary | `#A0A0A0` | Secondary text |
| Text Tertiary | `#6C6C6C` | Hints, placeholders |

### Accent Colors

| Name | Hex | Usage |
|------|-----|-------|
| Primary | `#339AF0` | Primary actions, links |
| Success | `#51CF66` | Completions, positive |
| Warning | `#FFA94D` | Warnings, attention |
| Error | `#FF6B6B` | Errors, destructive |

### Habit Colors

```swift
enum HabitColor: String, CaseIterable {
    case red = "#FF6B6B"
    case orange = "#FFA94D"
    case yellow = "#FFD43B"
    case green = "#51CF66"
    case teal = "#20C997"
    case cyan = "#22B8CF"
    case blue = "#339AF0"
    case indigo = "#5C7CFA"
    case violet = "#845EF7"
    case pink = "#F06595"
    case gray = "#868E96"
    case dark = "#495057"
}
```

## Typography

Using **SF Pro** (system font) for optimal readability.

| Style | Weight | Size | Line Height | Usage |
|-------|--------|------|-------------|-------|
| Large Title | Bold | 34 | 41 | Screen titles |
| Title 1 | Bold | 28 | 34 | Section headers |
| Title 2 | Bold | 22 | 28 | Card titles |
| Title 3 | Semibold | 20 | 25 | Subsections |
| Headline | Semibold | 17 | 22 | Important text |
| Body | Regular | 17 | 22 | Main content |
| Callout | Regular | 16 | 21 | Secondary content |
| Subheadline | Regular | 15 | 20 | Metadata |
| Footnote | Regular | 13 | 18 | Captions, hints |
| Caption | Regular | 12 | 16 | Small labels |

### SwiftUI Implementation

```swift
extension Font {
    static let hf = HFFont()

    struct HFFont {
        let largeTitle = Font.largeTitle.weight(.bold)
        let title1 = Font.title.weight(.bold)
        let title2 = Font.title2.weight(.bold)
        let title3 = Font.title3.weight(.semibold)
        let headline = Font.headline
        let body = Font.body
        let callout = Font.callout
        let subheadline = Font.subheadline
        let footnote = Font.footnote
        let caption = Font.caption
    }
}
```

## Spacing

Based on 4pt grid.

| Token | Value | Usage |
|-------|-------|-------|
| `xs` | 4 | Tight spacing |
| `sm` | 8 | Related elements |
| `md` | 16 | Standard spacing |
| `lg` | 24 | Section spacing |
| `xl` | 32 | Large gaps |
| `xxl` | 48 | Screen padding |

### SwiftUI Implementation

```swift
enum Spacing {
    static let xs: CGFloat = 4
    static let sm: CGFloat = 8
    static let md: CGFloat = 16
    static let lg: CGFloat = 24
    static let xl: CGFloat = 32
    static let xxl: CGFloat = 48
}
```

## Corner Radius

| Token | Value | Usage |
|-------|-------|-------|
| `sm` | 8 | Small buttons, badges |
| `md` | 12 | Cards, inputs |
| `lg` | 16 | Large cards, modals |
| `full` | 9999 | Pills, avatars |

## Shadows

| Level | Shadow | Usage |
|-------|--------|-------|
| `sm` | 0 1 2 rgba(0,0,0,0.1) | Subtle elevation |
| `md` | 0 4 6 rgba(0,0,0,0.1) | Cards |
| `lg` | 0 10 15 rgba(0,0,0,0.1) | Modals, popovers |

## Components

### Button

```swift
struct HFButton: View {
    enum Style {
        case primary    // Filled, accent color
        case secondary  // Outlined
        case ghost      // Text only
        case destructive // Red
    }

    enum Size {
        case small   // Height 32
        case medium  // Height 44
        case large   // Height 52
    }
}
```

**States**: default, pressed, disabled, loading

### Card

```swift
struct HFCard<Content: View>: View {
    // Background: Surface color
    // Padding: md (16)
    // Corner radius: md (12)
    // Shadow: md
}
```

### Input

```swift
struct HFTextField: View {
    // Background: Surface color
    // Border: 1px Border color
    // Height: 48
    // Corner radius: md (12)
    // Padding horizontal: md (16)
}
```

**States**: default, focused, error, disabled

### Toggle

```swift
struct HFToggle: View {
    // Uses system toggle
    // Tint: Primary color
}
```

### Progress Ring

```swift
struct HFProgressRing: View {
    var progress: Double  // 0.0 - 1.0
    var color: Color
    var lineWidth: CGFloat = 6

    // Animated fill
    // Shows percentage in center
}
```

### Habit Card

```swift
struct HabitCard: View {
    let habit: Habit
    let isCompleted: Bool

    // Icon (colored circle with SF Symbol)
    // Title
    // Streak badge
    // Completion indicator
    // Swipe to complete gesture
}
```

## Icons

Using **SF Symbols** exclusively.

| Usage | Symbol | Weight |
|-------|--------|--------|
| Add | `plus` | medium |
| Settings | `gearshape` | medium |
| Back | `chevron.left` | medium |
| Close | `xmark` | medium |
| Check | `checkmark` | bold |
| Delete | `trash` | medium |
| Edit | `pencil` | medium |
| Complete | `checkmark.circle.fill` | medium |
| Streak | `flame.fill` | medium |

## Animation

### Principles

1. **Quick but not instant** — 200-300ms for most transitions
2. **Spring for delight** — Bouncy animations for completions
3. **Respect Reduce Motion** — Fallback to fade for accessibility

### Timing

| Animation | Duration | Curve |
|-----------|----------|-------|
| Button tap | 100ms | easeOut |
| Card appear | 300ms | spring(0.6) |
| Modal present | 350ms | spring(0.8) |
| Completion | 400ms | spring(0.5, bounce: 0.3) |
| Page transition | 300ms | easeInOut |

### Completion Animation

```swift
withAnimation(.spring(response: 0.4, dampingFraction: 0.5)) {
    isCompleted = true
}

// Confetti for milestones
if streak.isMultiple(of: 7) {
    triggerConfetti()
}
```

## Haptics

| Action | Haptic |
|--------|--------|
| Button tap | `.light` |
| Toggle | `.medium` |
| Complete habit | `.success` |
| Delete | `.warning` |
| Error | `.error` |
| Pull to refresh | `.impact(style: .medium)` |

## Dark Mode

- Follows system setting by default
- Manual override in settings
- All colors adapt automatically
- Test both modes for every screen

## Accessibility

### Requirements

- [ ] VoiceOver labels on all interactive elements
- [ ] Dynamic Type support (up to xxxLarge)
- [ ] Minimum touch target 44x44
- [ ] Sufficient color contrast (4.5:1 minimum)
- [ ] Reduce Motion support
- [ ] No information conveyed by color alone

### Implementation

```swift
// VoiceOver
.accessibilityLabel("Complete Morning Run habit")
.accessibilityHint("Double tap to mark as done")

// Dynamic Type
.font(.body)  // Scales automatically

// Reduce Motion
@Environment(\.accessibilityReduceMotion) var reduceMotion

if reduceMotion {
    // Use simple fade instead of spring
}
```
