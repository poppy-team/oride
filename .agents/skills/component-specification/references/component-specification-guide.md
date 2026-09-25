# Component Specification Reference Guide

## Contract First

Define one responsibility and explicit non-goals before describing appearance. Name every prop with type, requiredness, default, constraints, and ownership. State whether each stateful prop is controlled, uncontrolled, or invalid when both value and default are supplied.

## State Matrix

Build the matrix from variants × sizes × interaction and data states. Include default, hover, focus-visible, active, loading, empty, error, disabled, and validation states where applicable. Every cell has behavior, visual tokens, accessibility semantics, and a test ID; unsupported combinations state why.

## Interaction Contract

Document pointer, keyboard, focus entry/return, and dismissal behavior. Prefer native elements before ARIA. Events carry serializable domain data, not DOM nodes. Slots have names and fallback content, and nested interactive content is prohibited unless explicitly designed.

## Accessibility Contract

Name the semantic element or role, accessible name source, description relationship, invalid state, required state, and live-region behavior. Define focus movement for opening, validation, deletion, and asynchronous completion.

## Versioning

Classify changes as patch, minor, or breaking. A new optional prop with a safe default can be minor; changing requiredness, meaning, focus order, or event payload is breaking. Preserve deprecated contracts for the stated migration window.

## Patterns and Anti-Patterns

| Do | Avoid |
|---|---|
| Specify observable behavior | Depend on implementation details |
| Use typed discriminated unions for variants | Accept arbitrary strings for every state |
| Make unsupported combinations explicit | Leave `TBD` matrix cells |
| Pair focus movement with state changes | Trap focus after close or deletion |
| Version the public contract | silently redefine an existing prop |

## Short Example

`variant: 'primary' | 'secondary' | 'danger'`; `size` defaults to `md`; `loading` disables activation but keeps the submit button focusable. `onSortChange` emits `{ columnId: string; direction: 'asc' | 'desc' }` and never the column DOM node.
