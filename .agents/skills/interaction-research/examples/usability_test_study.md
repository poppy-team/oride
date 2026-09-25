# Sample Usability Study: API Key Generation & Scoped Tokens

## 1. Study Overview
- **Product Area**: Developer Platform Settings (`/settings/api-tokens`)
- **Research Goal**: Evaluate if backend developers can successfully generate a scoped read-only API token and configure expiration without assistance.
- **Participants**: 5 Senior Backend Developers (P1 - P5)
- **Format**: Moderated Remote Think-Aloud (45-minute sessions)

---

## 2. Quantitative Results & Usability Metrics

### Summary Scorecard

| Participant | Task 1: Generate Scoped Token | Task 1: SEQ (1-7) | Task 2: Revoke Inactive Token | Task 2: SEQ (1-7) | SUS Score (0-100) |
| :---: | :---: | :---: | :---: | :---: | :---: |
| **P1** | Yes (42s) | 6 | Yes (18s) | 7 | 82.5 |
| **P2** | Yes (75s) | 5 | Yes (24s) | 6 | 77.5 |
| **P3** | Yes (55s) | 6 | Yes (15s) | 7 | 85.0 |
| **P4** | No (Timeout) | 3 | Yes (30s) | 5 | 62.5 |
| **P5** | Yes (60s) | 5 | Yes (20s) | 6 | 75.0 |
| **Average** | **80% Completion** | **5.0 / 7.0** | **100% Completion** | **6.2 / 7.0** | **76.5 (Grade B+)** |

*SUS Calculation Note*: Overall average of 76.5 exceeds the global industry average of 68.0, indicating solid baseline usability with specific scope selection friction.

---

## 3. Key Usability Insights & Affinity Clusters

### Finding 1: Scope Checkbox Hierarchy Causes Decision Paralysis (Task 1 Friction)
- **Observation**: 4 out of 5 participants hesitated when selecting permissions because 32 individual granular scopes were shown in an alphabetical unorganized list without grouped categories.
- **User Quote (P4)**: *"I just want read-only access to deployments, but I'm afraid I'll accidentally check a write scope or miss a necessary dependency."*
- **Recommendation**:
  Introduce pre-packaged permission bundles: `Read-Only`, `Deployment Admin`, and `Full Admin` as 1-click presets, with an expandable "Custom Scopes" progressive disclosure accordion.

### Finding 2: Token Secret Copy Confirmation is Highly Reassuring
- **Observation**: 5 out of 5 participants praised the full-screen modal showing the token secret with a prominent "Copy to Clipboard" button and clear warning that the secret will never be displayed again.
- **User Quote (P1)**: *"The immediate copy feedback and masked display gives me confidence I didn't lose the key."*
