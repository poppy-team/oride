# Visual Regression Test Report: {{Build.Identifier}}

## 1. Test Execution Metadata
- **Commit / PR**: `{{git_commit_sha}}` / PR #{{pr_number}}
- **Target Environment**: Docker Container (`{{docker_image_tag}}`)
- **Browser Engines Tested**: [Chromium | WebKit | Firefox]
- **Viewports Tested**: Desktop (1280x720), Tablet (768x1024), Mobile (375x667)
- **Date**: {{YYYY-MM-DD}}
- **Overall Status**: [PASS | VISUAL DIFF DETECTED - REVIEW REQUIRED]

---

## 2. Test Execution Summary

| Test Suite / View | Viewport | Theme | Mismatched Pixels | Diff Ratio (%) | Status |
| :--- | :---: | :---: | :---: | :---: | :---: |
| `Button.stories.tsx` | Desktop | Light | 0 | 0.00% | PASS |
| `Button.stories.tsx` | Desktop | Dark | 0 | 0.00% | PASS |
| `HeaderNavigation.tsx`| Mobile | Light | 42 | 0.03% | PASS (Within Tol.) |
| `DashboardOverview.tsx`| Desktop | Light | {{diff_pixels}} | {{diff_ratio}}% | {{status}} |

---

## 3. Visual Failure Artifacts & Triage

### Diff Failure #1: {{Test_Case_Name}}
- **Component / Route**: `{{component_route}}`
- **Engine / Viewport**: `{{browser}}` @ `{{viewport_dimensions}}`
- **Mismatched Pixels**: `{{pixel_count}}` (Allowed threshold: `{{max_diff_pixels}}`)
- **Failure Analysis**:
  - [ ] Intentional design update (requires baseline update).
  - [ ] Unintentional layout shift / CSS regression (defect).
  - [ ] Flaky dynamic content / unmasked timestamp (test defect).

#### Visual Comparison Triple
- **Baseline Image**: `{{artifact_path_baseline}}`
- **Actual Captured Image**: `{{artifact_path_actual}}`
- **Diff Image (Magenta Mismatch)**: `{{artifact_path_diff}}`

---

## 4. Sign-Off & Baseline Update Actions
- [ ] Visual diffs reviewed and approved by UI/UX Lead.
- [ ] Run `prumo test visual --update-snapshots` if intentional.
- [ ] PR committed with updated baseline images.
