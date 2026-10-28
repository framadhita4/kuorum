---
trigger: model_decision
description: Design system guidelines, semantic color palette, typography scale, spacing, and visual component specifications.
---

# Rule: Design System & Visual Guidelines

## Overview
Applies a "Warm Enterprise Productivity" aesthetic that pairs neo-grotesk typographic precision (`TWK Lausanne`) with a light font weight (*light* 300) on large headings, high contrast between a clean white canvas (`#FFFFFF`) and deep ink black (`#0D0E10`), and energetic accent colors such as *Light Coral* (`#F06A6A`), *Medium Slate Blue* (`#796EFF`), and *Claret* (`#690031`) to distinguish *Agentic Work Management* features from traditional corporate software.

---

## 01. Colors (Semantic Palette)

### Surface & Canvas
- **Canvas Primary (`--color-bg-primary`):** `#FFFFFF` — Primary page background.
- **Surface Subtle (`--color-bg-subtle`):** `#F6F8F9` — Secondary section background, feature cards, and block dividers.
- **Surface Neutral (`--color-bg-neutral`):** `#F3F3F3` — Inactive tab pills and secondary containers.
- **Surface Dark Charcoal (`--color-bg-charcoal`):** `#222B37` — Dark mastheads and product highlight blocks.
- **Surface Cool Near Black (`--color-bg-dark`):** `#151B26` — Dark sections (dark mode sections) and enterprise card backgrounds.
- **Surface Deep Navy (`--color-bg-footer`):** `#0D1B2F` — Enterprise hero background and global footer.

### Ink (Text & Iconography)
- **Ink Primary (`--color-ink-primary`):** `#0D0E10` — Primary text color, headings, and dark primary buttons (AAA contrast `19.31:1` on `#FFFFFF`).
- **Ink Secondary (`--color-ink-secondary`):** `#4A5464` — Supporting paragraphs, card descriptions, and sub-labels (AAA contrast `7.65:1` on `#FFFFFF`).
- **Ink Muted (`--color-ink-muted`):** `#646F79` — Placeholder text, captions, metadata, and secondary icons.
- **Ink Inverse (`--color-ink-inverse`):** `#FFFFFF` — Text on dark buttons and dark-surfaced sections.

### Brand & Semantic Accents
- **Brand Coral (`--color-brand-coral`):** `#F06A6A` — Primary brand color, hover state for primary/secondary buttons.
- **Brand Red Orange (`--color-brand-vibrant`):** `#FF584A` — High-intensity brand glyph accent.
- **Brand Claret (`--color-brand-claret`):** `#690031` — Deep maroon accent for Enterprise themes and statistic cards.
- **Accent Slate Purple (`--color-accent-purple`):** `#796EFF` — AI / Agentic Workflows indicator, active navigation, and gradient starting point.
- **Accent Slate Purple Hover (`--color-accent-purple-dark`):** `#635AC7` — Border and hover state for purple components.
- **Accent Lilac (`--color-accent-lilac`):** `#FFC9FF` — Surface highlight for AI Studio & AI Teammates features.
- **Accent Deep Sky Blue (`--color-accent-blue`):** `#00B2FF` — Interactive video player controls and timeline accents.
- **Accent Midnight Blue (`--color-accent-midnight`):** `#222875` — Active tab background (`.tab_button--blue`).
- **Accent Gold (`--color-accent-gold`):** `#FCBD01` — Customer marquee cards and spotlight badges.
- **Semantic Danger / Active Red (`--color-danger`):** `#CD4848` — Active sidebar indicator and warning/error messages.

### Gradients
- **Surface Fade Gradient:** `linear-gradient(180deg, #FFFFFF 0%, #F6F8F9 100%)`
- **Signature AI / Masthead Gradient:** `linear-gradient(90deg, #796EFF 0%, #D472C5 45%, #DF70B8 65%, #FF6D92 100%)`
- **Media Overlay Scrim:** `linear-gradient(180deg, rgba(0, 0, 0, 0) 0%, rgba(0, 0, 0, 0.75) 100%)`

---

## 02. Typography (Type System)

### Font Families
- **Primary Sans (`--font-sans`):** `'TWK Lausanne', 'Helvetica Neue', 'Helvetica', Arial, sans-serif`
  - *Self-hosted `@font-face`*, loaded weights: `250`, `300`, `350`, `400`, `450`, `500`, `550`.
- **Display Accent (`--font-display`):** `'Ghost', 'TWK Lausanne', sans-serif` (Weight `500`, reserved for specific editorial accents).
- **Fallback / International (`--font-cjk`):** `'Hiragino Kaku Gothic ProN', sans-serif`

### Type Scale (Desktop 1440px → Mobile 390px)
| Role | Font Family | Size (Desktop) | Size (Mobile) | Line Height | Weight | Letter Spacing |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Display / Hero H1** | `TWK Lausanne` | `102px` (`6.375rem`) | `38px` (`2.375rem`) | `0.90` (`92px`) / `1.0` | `300` (Light) | `-0.03em` |
| **Heading 2 (Section)** | `TWK Lausanne` | `54px` (`3.375rem`) | `30px` (`1.875rem`) | `1.00` (`54px`) / `1.1` | `300` (Light) | `-0.02em` |
| **Heading 3 (Feature)** | `TWK Lausanne` | `36px` (`2.25rem`) | `23px` (`1.4375rem`) | `1.30` (`47px`) / `1.2` | `300`–`400` | `-0.015em` |
| **Heading 4 (Card Title)** | `TWK Lausanne` | `30px` (`1.875rem`) | `20px` (`1.25rem`) | `1.20` (`36px`) / `1.4` | `400` (Regular) | `-0.01em` |
| **Heading 5 (Subhead)** | `TWK Lausanne` | `24px` (`1.5rem`) | `18px` (`1.125rem`) | `1.20` (`29px`) / `1.4` | `500` (Medium) | `0em` |
| **Heading 6 (Overline)** | `TWK Lausanne` | `12px` (`0.75rem`) | `12px` (`0.75rem`) | `1.50` (`18px`) | `500` (Medium) | `0.08em` (Uppercase) |
| **Body Large (Lead)** | `TWK Lausanne` | `20px` (`1.25rem`) | `16px` (`1rem`) | `1.50` (`30px`) | `300` (Light) | `0em` |
| **Body Base** | `TWK Lausanne` | `16px` (`1rem`) | `14px` (`0.875rem`) | `1.50` (`24px`) | `300`–`400` | `0em` |
| **Button / CTA** | `TWK Lausanne` | `16px` (`1rem`) | `14px` (`0.875rem`) | `1.50` (`24px`) | `500` (Medium) | `0.01em` |
| **Inline Link** | `TWK Lausanne` | `14px` (`0.875rem`) | `14px` (`0.875rem`) | `1.50` (`21px`) | `500` (Medium) | `0em` |

---

## 03. Spacing, Shape, and Depth

### Spacing Scale (Base Unit: `4px`)
- `space-1`: `4px` (`0.25rem`)
- `space-2`: `8px` (`0.5rem`)
- `space-3`: `12px` (`0.75rem`)
- `space-4`: `16px` (`1rem`)
- `space-6`: `24px` (`1.5rem`) — Standard gutter between cards.
- `space-8`: `32px` (`2rem`) — Card internal padding.
- `space-12`: `48px` (`3rem`) — Spacing between sub-sections.
- `space-16`: `64px` (`4rem`) — Vertical section padding on mobile.
- `space-20`: `80px` (`5rem`) — Standard vertical section padding on desktop.
- `space-30`: `120px` (`7.5rem`) — Macro divider between major sections.

### Border Radius
- `radius-none`: `0px`
- `radius-sm`: `4px` — Compact buttons, checkboxes, and small input elements.
- `radius-md`: `8px` — Standard CTA buttons (`button.-dark`, `button.-light`) and input fields.
- `radius-lg`: `12px` — Secondary feature cards and dropdown menus.
- `radius-xl`: `16px` — Bento cards, product preview containers, and video wrappers.
- `radius-2xl`: `24px` — Large hero showcase containers.
- `radius-pill`: `9999px` — Category filter tabs (`.tab_button`), badges, and rounded navigation pills.

### Elevation & Shadows
- `shadow-sm`: `0 2px 6px rgba(13, 14, 16, 0.04)` — Flat cards with subtle separation.
- `shadow-md`: `0 8px 24px rgba(13, 14, 16, 0.08)` — Card hover state and sticky header components.
- `shadow-lg`: `0 16px 40px rgba(13, 14, 16, 0.12)` — Mega-menu navigation, modals, and product UI demonstration frames.
- `border-subtle`: `1px solid rgba(100, 111, 121, 0.2)` (`#E2E6E9`) — Standard card border line.

---

## 04. Page Composition (Macro Layout Map)

1. **01 Sticky Global Navigation (`#FFFFFF`):** Centered navigation bar (`max-width: 1280px`), height `72px`, featuring Coral/Black logo on the left, dropdown menus (*Products, Solutions, Resources, Pricing*), language selector (Globe icon), *Sign in* link, and two CTA buttons (*Contact Sales* & *Get Started*).
2. **02 Hero Section (`linear-gradient(180deg, #FFFFFF 0%, #F6F8F9 100%)`):** Centered single-column layout for a prominent headline (`TWK Lausanne` 300) with two side-by-side action buttons, followed by an interactive canvas / interface showcase (using Three.js / interactive tabs).
3. **03 Enterprise Trust & Customer Marquee (`#FFFFFF`):** Partner company logo row alongside colorful metric cards (*Gold* `#FCBD01`, *Coral* `#F06A6A`, *Lilac* `#FFC9FF`).
4. **04 Interactive Workflow Tabs (`#F6F8F9`):** Horizontal cluster of pill tabs (`.tab_button`) toggling product preview viewports across various departments (*Agentic Work Management, IT, Marketing, Operations*).
5. **05 Bento Grid & AI Features (`#151B26` / `#690031`):** High-contrast dark section with 2 to 3 columns on desktop (stacked vertically on mobile) spotlighting *AI Studio* and no-code automation.
6. **06 Enterprise Proof & Stats (`#FFFFFF`):** 3-column grid layout for key impact metrics (e.g., `100k+`, `200k+`, `85%`) using `54px`–`72px` figures at font weight `300`.
7. **07 Global Footer (`#0D1B2F`):** Multi-column grid (5 link columns on desktop, 1 accordion column on mobile) over a deep navy background with `#FFFFFF` and `#646F79` typography.

---

## 05. Component Specifications

### Buttons
- **Primary Dark (`button.-dark`):**
  - Background: `#0D0E10` | Text: `#FFFFFF`
  - Hover: Background changes to `#F06A6A` (*Light Coral*), Text changes to `#0D0E10`
  - Padding: `14px 24px` (Desktop), `10px 18px` (Compact)
  - Border Radius: `6px`–`8px` | Font: `16px`, weight `500`
- **Secondary Outlined (`button.-border`):**
  - Background: `transparent` | Text: `#0D0E10` | Border: `1px solid #0D0E10`
  - Hover: Background `#F06A6A`, Border `#F06A6A`, Text `#0D0E10`
- **Dark Surface Outlined (`button.-dark.-border`):**
  - Background: `#151B26` | Text: `#FFFFFF` | Border: `1px solid rgba(255, 255, 255, 0.3)`
  - Hover: Background `#FFFFFF`, Text `#0D0E10`

### Navigation & Filter Tabs (`.tab_button`)
- **Default:** Background `#F3F3F3`, Text `#4A5464`, Radius `9999px`, Padding `10px 20px`, Font `14px` weight `500`.
- **Active State:** Background `#222875` (*Midnight Blue*) or `#796EFF` (*Medium Slate Blue*), Text `#FFFFFF`.

### Feature & Bento Cards
- **Surface:** `#FFFFFF` (on `#F6F8F9` sections) or `#222B37` (on dark `#151B26` sections).
- **Padding:** `32px` (`2rem`) | **Radius:** `16px` | **Border:** `1px solid rgba(100, 111, 121, 0.16)`.
- **Internal Rhythm:** Category label (`12px` uppercase weight `500`) → Card title (`24px`–`30px` weight `400`) → Description (`16px` weight `300` `#4A5464`) → Action link with arrow (`Learn more →`).

---

## 06. Do's and Don'ts for AI Agents

- **DO:** Use light font weight (`font-weight: 300`) for all large headings (`H1`, `H2`, `H3`) and lead paragraphs to preserve the signature refined elegance of `TWK Lausanne`.
- **DO:** Apply immediate hover color shifts on primary buttons: transitioning directly from solid ink black (`#0D0E10`) to Light Coral (`#F06A6A`) with black text (`#0D0E10`).
- **DO:** Maintain generous vertical whitespace (`80px`–`120px` between sections) and limit paragraph text measure to a maximum of `65ch`.
- **DON'T:** Do not use bold font weight (`font-weight: 700`) on Hero Headings (`H1`/`H2`); use large scale sizing (`54px`–`102px`) with light weight `300` and tight letter-spacing (`-0.02em` to `-0.03em`).
- **DON'T:** Do not use generic corporate blue (`#0000FF` or `#3B82F6`) for primary buttons; use `#0D0E10`, `#F06A6A`, or `#796EFF`.
