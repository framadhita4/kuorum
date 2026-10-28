---
trigger: model_decision
description: This guide is **MANDATORY** for all AI Agents implementing or modifying code in the **Frontend (React + Vite + TypeScript + TanStack)**.
---

# Rule: Frontend Architecture & Conventions

This guide is **MANDATORY** for all AI Agents implementing or modifying code in the **Frontend (React + Vite + TypeScript + TanStack)**.

---

## 1. Frontend Stack & Tooling
- **Framework & Bundler:** React 18/19 + Vite + TypeScript (Strict Mode).
- **Styling Engine:** `@emotion/react` + `@emotion/styled` with a single centralized theme (`#/-components/ui/theme`).
- **Routing:** TanStack Router (`@tanstack/react-router`) with file-based routing.
- **Server State & Data Fetching:** TanStack Query (`@tanstack/react-query`).
- **Form & Validation:** `react-hook-form` + `@hookform/resolvers/zod` + `zod`.
- **Path Alias:** Must use `#/` pointing to `src/` (e.g., `import { Button } from '#/-components/ui/Button'`).

---

## 2. Directory Structure

```
frontend/src/
├── -components/                 # GLOBAL COMPONENTS (Cross-feature)
│   ├── ui/                      # Primitive UI (Button, TextInput, Modal, Badge, etc.)
│   ├── providers/               # Context Providers (QueryClient, Theme, etc.)
│   └── layout/                  # Shell / AppLayout / Header / Sidebar
├── constant/                    # Global constants
├── integrations/
│   └── tanstack-query/          # QueryClient setup & default options
├── lib/                         # External client helpers & configurations
├── routes/                      # ROUTE PAGES (TanStack File-Based Routing)
│   ├── __root.tsx               # Root route layout
│   ├── index.tsx                # Home / Landing page
│   └── <feature-route>/         # Feature page route
│       ├── index.tsx            # Main route component
│       ├── -components/         # Private components ONLY for this page
│       ├── -schema/             # Zod validation schemas & form types (co-located with form)
│       └── -utils/              # Route-specific utility helpers
├── services/                    # GLOBAL SERVICES & DATA FETCHING
│   └── <feature>/               # Domain services (TanStack Query hooks)
│       ├── use<Action>.ts       # Custom useMutation / useQuery hooks
│       └── <feature>Storage.ts  # Local storage helpers if needed
├── types/                       # GLOBAL DOMAIN TYPES
│   └── <feature>.type.ts        # Global entity interfaces & domain types
└── utils/                       # Shared utility functions (fetchHelper, formatting)
```

---

## 3. Strict Component & Folder Rules

### A. Dash Prefix (`-`) for Non-Route Directories
- Any directory inside `src/routes/` that is NOT a route page **MUST** start with a dash prefix (`-`):
  - `-components/`
  - `-schema/`
  - `-utils/`
- This ensures TanStack Router generator ignores them when generating the route tree.
- The global component directory in `src/` also uses `src/-components/`.

### B. Co-location of Form Schemas (`-schema/`)
- Form validation schemas (Zod) and inferred form types **MUST be placed in `-schema/` adjacent to their form component**.
- **STRICTLY PROHIBITED** to place feature-specific form validation schemas in global `src/types/`. Global `src/types/` is reserved for backend entity data contracts.

---

## 4. Services & Data Fetching (TanStack Query)

1. **One Hook Per Operation:**
   - Each endpoint/use-case must have its own dedicated file in `src/services/<feature>/`, e.g.:
     - `src/services/exploration/useDecomposeConcept.ts`
     - `src/services/exploration/useGetAngles.ts`
2. **Standard HTTP Wrapper:**
   - Use the `fetchHelper` utility for all REST requests:
     ```ts
     const response = await fetchHelper<ApiResponse<ResultType>, PayloadType>(
       '/api/v1/exploration/decompose',
       { method: 'POST', body: payload }
     );
     ```
3. **Response Envelope Handling:**
   - All server responses are wrapped in a generic envelope:
     ```ts
     export type ApiResponse<T> = {
       success: boolean;
       message?: string;
       data: T;
       error?: string | null;
     };
     ```
4. **No Server State Duplication:**
   - Data returned from the server **MUST NEVER** be synchronized into local `useState`. Always consume TanStack Query values directly (`data`, `isLoading`, `error`).

---

## 5. TypeScript Strict Standards
- **STRICTLY FORBIDDEN** to use the `any` type. Use `unknown` with explicit type guards/narrowing.
- All component props must be explicitly typed using `interface` or `type`.
- Avoid `React.FC`. Use explicit functional signatures:
  ```tsx
  interface ConceptCardProps {
    title: string;
    description: string;
    onSelect: () => void;
  }

  export const ConceptCard = ({ title, description, onSelect }: ConceptCardProps) => {
    return <div>...</div>;
  };
  ```

---

## 6. Styling Engine & Architecture Rules (@emotion/react)

Styling in the frontend is strictly powered by **`@emotion/react`** and typed via Emotion Theme context.

### A. One Main Theme File as Single Source of Truth
- **Single Source of Truth File:** `frontend/src/-components/ui/theme.ts` (aliased as `#/-components/ui/theme`).
- All color values, font sizes, font families, line heights, letter spacings, border radii, shadows, and spacing steps MUST be defined in this one theme file.
- `src/emotion.d.ts` extends `@emotion/react`'s `Theme` interface (`export interface Theme extends CustomTheme {}`) to ensure 100% type-safety in `styled`, `css={(theme) => ...}`, and `useTheme()`.
- **STRICTLY PROHIBITED** to hardcode arbitrary hex colors (e.g. `#ffffff`, `#121212`, `#f06a6a`) or ad-hoc typography values directly across pages/routes. Always consume tokens from `theme`.

### B. Core Styling Separation Rule: Component Visuals vs. Layout Orchestration

1. **Component Internal Styling (Controlled Strictly by the Component):**
   - **Properties Governed:**
     - Colors (`color`, `background-color`, `border-color`, `fill`, `stroke`)
     - Typography (`font-size`, `font-family`, `font-weight`, `line-height`, `letter-spacing`, `text-transform`)
     - Surface appearance (`border-radius`, `box-shadow`, `border-width`, `border-style`)
     - Internal padding (the component's internal breathing room)
     - Interactive states (`:hover`, `:active`, `:focus-visible`, `:disabled`)
   - **Rule:**
     - These visual attributes **MUST be controlled by the component itself** using Emotion (`styled` or `css`), referencing the central `theme`.
     - External callers and parent layouts must **NOT** reach into a component to override its internal colors, fonts, or text size via ad-hoc style overrides.
     - If a component requires visual variation (e.g., subtle vs solid, small vs large), it must expose explicit, typed semantic props (e.g., `variant="primary" | "secondary" | "outline" | "ghost"`, `size="sm" | "md" | "lg"`).

2. **Layout Styling (Adjusted Directly by Parent / Consumer):**
   - **Properties Governed:**
     - Display & Flexbox (`display: flex | inline-flex`, `flex-direction`, `align-items`, `justify-content`, `flex-wrap`, `flex-grow`, `flex-shrink`, `gap`)
     - CSS Grid (`display: grid`, `grid-template-columns`, `grid-template-rows`, `grid-column`, `grid-row`, `gap`)
     - Positioning (`position: relative | absolute | fixed | sticky`, `top`, `right`, `bottom`, `left`, `z-index`)
     - Spatial dimensions & layout margins (`width`, `height`, `min-width`, `max-width`, `margin`, `overflow`)
   - **Rule:**
     - Spatial arrangement, positioning, and layout orchestration **CAN and SHOULD be adjusted directly** at the call site or container level (e.g. using Emotion's `css` prop, layout wrapper components like `BentoGrid` / `Flex` / `Box`, or direct layout styles).
     - UI primitive components (such as `Button`, `Badge`, `Input`, `Card`) **MUST NOT** include baked-in outer margins or fixed page-positioning assumptions. They are layout-agnostic building blocks positioned directly by their parent layout containers.

```tsx
// ✅ CORRECT: Layout adjusted directly, component visuals controlled by component & theme
<div
  css={(theme) => ({
    display: 'flex',
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: theme.spacing['4'],
    width: '100%',
  })}
>
  <Button variant="primary" size="md">
    Save Exploration
  </Button>
  <Badge variant="success" dot size="sm">
    Synced
  </Badge>
</div>

// ❌ WRONG: Overriding component's internal color & font size from the outside
<Button style={{ backgroundColor: '#ff0000', fontSize: '20px', margin: '20px' }}>
  Click Me
</Button>
```

