---
name: design-system
description: Semantic design tokens, component contract specifications, Storybook documentation, responsive typography scales, and a11y tokens
---
# Design System Architecture & Tokens

## 1. Semantic Design Tokens
Structure design tokens in three tiers: Global (raw values: blue-500: #3b82f6), Semantic (intent-based: color-primary: var(--blue-500)), and Component-scoped (button-bg: var(--color-primary)).

## 2. Format-Agnostic Token Distribution
Maintain tokens in JSON format and compile them to platform targets: CSS custom properties, SCSS variables, JavaScript/TypeScript objects, and iOS/Android tokens via Style Dictionary.

## 3. Component API Contract Specification
Define strict props, slots, and events for every design system component. Enforce variant constraints via TypeScript union types (e.g. variant: 'primary' | 'secondary' | 'ghost').

## 4. Accessibility (A11y) Token Integration
Ensure color tokens strictly satisfy WCAG 2.2 AA contrast ratios (4.5:1 for normal text, 3:1 for large text). Provide high-contrast and reduced-motion token overrides.

## 5. Spacing & Fluid Typography Scales
Use a 4px or 8px baseline grid system for all margin, padding, and layout dimensions. Implement fluid typography using clamp() to scale text smoothly across viewports.

## 6. Component Composition & Headless Architecture
Separate component logic/state from styling. Build on accessible headless primitives (Radix UI, Headless UI, React Aria) to guarantee robust keyboard navigation and ARIA patterns.

## 7. Living Documentation & Storybook Catalog
Document every component in Storybook with interactive states (default, hover, active, focus, disabled, error). Document component usage guidelines and anti-patterns.

## 8. Visual Regression Testing Gates
Automate visual regression tests on Storybook stories in CI (using Chromatic or Playwright) to catch unintended styling regressions before merging.

## 9. Semantic Versioning & Deprecation Lifecycle
Follow strict semver for design system releases. Provide codemods and migration guides whenever component API signatures change or deprecations occur.

## 10. Theme Switching & Dark Mode Support
Implement dark mode and custom white-label themes via CSS class or data-theme attributes, switching semantic token variables with zero page refresh.
