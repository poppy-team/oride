# Component Specification Verification Checklist

## Responsibility and Scope
- [ ] The component has one stated responsibility and explicit non-goals.
- [ ] Ownership is clear for adjacent behavior such as pagination, selection, or persistence.
- [ ] Public names match existing component-system conventions.

## Props and State
- [ ] Every prop has a type, requiredness, default, constraint, and ownership rule.
- [ ] Controlled and uncontrolled state cannot be active simultaneously.
- [ ] Every variant, size, and state-matrix cell is resolved or justified as unsupported.

## Interaction and Accessibility
- [ ] Mouse or pointer actions have keyboard equivalents or a documented exception.
- [ ] Focus entry, movement, return, dismissal, and deletion behavior are explicit.
- [ ] Semantic element or role, accessible name, description, state, and live-region behavior are defined.

## Events and Slots
- [ ] Events carry serializable domain data and define cancellation or error behavior.
- [ ] Named slots define permitted content, fallback content, and nesting constraints.
- [ ] Public props and events include a versioning and deprecation policy.

## Fixtures and Sign-Off
- [ ] Fixtures cover default, hover, focus-visible, loading, empty, error, and disabled states.
- [ ] Responsive, zoom, and high-contrast behavior is specified.
- [ ] `scripts/verify.sh` exits with code 0 from the repository root.
