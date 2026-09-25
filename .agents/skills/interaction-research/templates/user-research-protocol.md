# Usability Testing Protocol & Study Plan: {{Study.Title}}

## 1. Study Overview & Goals
- **Product Area**: `{{product_area}}`
- **Primary Research Questions**:
  1. {{Question 1: e.g. Can new users successfully complete first-time onboarding without documentation?}}
  2. {{Question 2: e.g. What causes drop-off in the checkout payment step?}}
- **Format**: [Moderated Remote | Moderated In-Person | Unmoderated Asynchronous]
- **Sample Size**: {{participant_count}} participants (Target: 5-8 per cohort)

---

## 2. Participant Screener & Demographics
- **Target Persona**: `{{persona_title}}`
- **Inclusion Criteria**:
  - [ ] {{Requirement 1: e.g. Employs Go or Rust in production}}
  - [ ] {{Requirement 2: e.g. Familiar with CI/CD tools}}
- **Exclusion Criteria**:
  - [ ] Exclude employees, contractors, or direct competitors.

---

## 3. Test Scenarios & Task Scripts

### Task 1: {{Task Name - e.g. Project Initialization}}
- **User Scenario**:
  > "{{Context: Imagine you just joined a new engineering team. You want to initialize a new project workspace named 'billing-service'.}}"
- **Success Criteria**: {{User locates 'New Project' and successfully creates repository}}
- **Post-Task Metric**: Single Ease Question (SEQ 1-7 scale)
- **Max Time Allotted**: 3 minutes

### Task 2: {{Task Name - e.g. Configure Pipeline}}
- **User Scenario**:
  > "{{Context: Your project requires an automated testing step before merge. Set up the default test runner.}}"
- **Success Criteria**: {{User attaches test action to pipeline}}
- **Post-Task Metric**: Single Ease Question (SEQ 1-7 scale)
- **Max Time Allotted**: 5 minutes

---

## 4. Quantitative Usability Scorecard

| Participant ID | Task 1: Complete? | Task 1: SEQ (1-7) | Task 2: Complete? | Task 2: SEQ (1-7) | Overall SUS Score (0-100) |
| :---: | :---: | :---: | :---: | :---: | :---: |
| **P1** | [Yes/No] | {{seq_p1_t1}} | [Yes/No] | {{seq_p1_t2}} | {{sus_p1}} |
| **P2** | [Yes/No] | {{seq_p2_t1}} | [Yes/No] | {{seq_p2_t2}} | {{sus_p2}} |
| **Average** | **{{tcr_t1}}%** | **{{avg_seq_t1}}** | **{{tcr_t2}}%** | **{{avg_seq_t2}}** | **{{avg_sus}} (Grade: {{grade}})** |

---

## 5. Qualitative Findings & Affinity Clusters
- **Theme 1**: {{Summary of primary user confusion or pattern}}
  - *Observation*: {{Observed behavior in 4/5 participants}}
  - *User Quote*: *"{{Representative direct quote}}"*
  - *Recommended Fix*: {{Concrete design or technical modification}}

---

## 6. Action Items & Sprint Backlog
- [ ] **P0**: {{Immediate critical usability fix}}
- [ ] **P1**: {{Secondary workflow improvement}}
