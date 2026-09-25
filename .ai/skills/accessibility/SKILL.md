---
name: accessibility
description: WCAG 2.2 AA complete checklist, keyboard paths, visible focus, semantic naming, contrast, screen reader, zoom/reflow, reduced motion
---
# Accessibility Engineering

## 1. Semantic HTML
Use semantic HTML elements (nav, main, article, button) instead of generic divs. This provides inherent meaning to assistive technologies. Avoid using aria-roles to fix poor markup; fix the markup instead.

## 2. Keyboard Navigation
Ensure all interactive elements are reachable and operable via keyboard alone. Verify logical tab order matching the visual flow. Avoid keyboard traps where focus gets stuck in a component.

## 3. Visible Focus
Provide highly visible focus indicators for all interactive elements. Do not remove outlines (`outline: none`) without providing a strong visual alternative. Focus should be evident without relying solely on color.

## 4. Contrast Ratios
Maintain strict WCAG 2.2 AA contrast ratios. Text requires a contrast ratio of at least 4.5:1 against its background. Large text requires 3:1. UI components and graphical objects require a 3:1 contrast ratio.

## 5. Screen Reader Support
Test interfaces with popular screen readers (NVDA, VoiceOver). Ensure dynamic content updates (like notifications or loading states) are announced using `aria-live` regions appropriately.

## 6. Zoom and Reflow
Ensure content supports zooming up to 200% without loss of functionality. Implement responsive designs that reflow content into a single column at 400% zoom without requiring horizontal scrolling.

## 7. Reduced Motion
Respect user preferences for reduced motion (`@media (prefers-reduced-motion)`). Disable non-essential animations, parallax effects, and smooth scrolling for these users to prevent vestibular disorders.

## 8. Forms and Labels
Every input must have a programmatic association with a visible label using the `for` and `id` attributes. Error messages must be clearly linked to the specific input field that caused the error.

## 9. Alternative Text
Provide meaningful `alt` text for informational images. Use empty `alt=""` for purely decorative images. Ensure complex charts or graphs have comprehensive text alternatives.

## 10. Cognitive Accessibility
Write clear, simple copy. Avoid jargon. Provide predictable navigation patterns. Ensure critical actions can be easily reversed or confirmed to minimize user error anxiety.

