---
name: interaction-research
description: User experience research (UXR), usability testing protocols, qualitative/quantitative metrics (SUS, SEQ, Task Completion Rate), think-aloud facilitation, and thematic insight synthesis.
---

# Interaction & User Experience Research (UXR) Contract

## 1. Purpose
Plan, conduct, analyze, and synthesize empirical user experience research (UXR) and usability testing studies. Formulate non-leading task scenarios, administer standardized metrics (System Usability Scale - SUS, Single Ease Question - SEQ, Task Completion Rate - TCR), extract behavioral friction points, and deliver data-driven recommendations that guide product architecture.

---

## 2. Use When
- Formulating usability testing protocols and interview scripts for prototype or feature evaluation.
- Benchmarking software usability before and after major redesigns or architecture migrations.
- Measuring user task completion efficiency, error rates, and cognitive friction in core funnels.
- Synthesizing qualitative feedback from user testing sessions into prioritized product insights.

---

## 3. Do Not Use When
- Designing visual layout tokens or component styling (use `design-tokens` or `ui-implementation`).
- Conducting automated headless pixel comparison (use `visual-regression`).
- Writing production application code.

---

## 4. Required Context
Before formulating a research study, verify:
1. **Research Objectives & Hypotheses**: What specific behaviors, assumptions, or risks is this study designed to test?
2. **Target User Cohorts**: Screener criteria, domain expertise level, customer segment.
3. **Stimulus Material**: Figma prototype, staging deployment, production interface, or wireframe.
4. **Study Format**: Moderated synchronous (remote/in-person) vs Unmoderated asynchronous (UserTesting, Maze).

---

## 5. Procedure

### Step 1: Study Design & Participant Screening
1. Define primary research questions and falsifiable hypotheses.
2. Draft a screening survey selecting 5 to 8 participants per distinct user segment (Nielsen's rule: 5 users uncover 85% of usability issues).
3. Ensure participant diversity and exclude internal stakeholders or biased individuals.

### Step 2: Task Scenario & Protocol Authoring
1. Design realistic, goal-oriented scenarios that avoid leading phrasing:
   - *Good*: "You need to find out how much you spent on compute in August. Show me how you'd find that."
   - *Bad (Leading)*: "Go to the Billing tab and click on the August invoice button."
2. Define explicit success criteria, maximum time allotment, and allowed assistance thresholds for each task.

### Step 3: Session Facilitation & Think-Aloud Execution
1. Obtain informed consent and record permissions (GDPR/privacy compliance).
2. Facilitate sessions using the **Think-Aloud Protocol**: encourage participants to verbalize their thoughts, expectations, hesitations, and surprises in real time.
3. Maintain neutral facilitation: never defend the design, explain the UI, or instruct the user unless they are irrecoverably blocked.

### Step 4: Quantitative Metrics Collection
Administer standardized post-task and post-study questionnaires:
1. **Single Ease Question (SEQ)**: 1-item 7-point scale administered immediately after each task:
   *"Overall, how easy or difficult was it to complete this task?"* (1 = Very Difficult, 7 = Very Easy; target: $\ge 5.5$).
2. **Task Completion Rate (TCR)**: Binary measure (0 = Failed/Abandoned, 1 = Succeeded). Target: $\ge 80\%$.
3. **System Usability Scale (SUS)**: 10-item standard questionnaire administered at the conclusion of the study, scored from 0 to 100 (industry benchmark average: 68; target: $\ge 75$ for high satisfaction).

### Step 5: Thematic Synthesis & Insight Prioritization
1. Group observed behavioral friction into thematic clusters using affinity mapping.
2. Quantify issue severity by frequency (how many users struggled) and impact (blocking vs minor delay).
3. Deliver clear, actionable design and engineering recommendations.

---

## 6. Decision Rules
1. **Zero Leading Prompts**: Prompts must describe user goals, never UI mechanisms or specific menu names.
2. **Observe Behavior Over Stated Preference**: What users do is authoritative over what users say. If a user states a task was "easy" but took 8 minutes and 3 errors, record it as a high-friction task.
3. **Standardized SUS Scoring**: The System Usability Scale must be calculated using the exact standard odd/even normalization formula; ad-hoc modifications invalidate benchmarks.
4. **Privacy & Data Redaction**: User PII, credentials, or sensitive organizational data must be scrubbed from transcripts and video clips.

---

## 7. Evidence Required
- **Research Protocol Document**: Complete test plan with screeners, task scenarios, and metrics.
- **Raw Observations & Metric Logs**: Task completion matrix, ToT, SEQ scores, and SUS score calculations.
- **Synthesized Findings Report**: Thematic findings linked to timestamped/recorded behavioral evidence.

---

## 8. Output Contract
A production UXR deliverable must contain:
1. User Research Protocol & Plan (`templates/user-research-protocol.md`).
2. Quantitative scorecard (TCR, SEQ, SUS).
3. Key insights matrix with prioritized design/engineering recommendations.

---

## 9. Stop Conditions
- Required participant cohort sessions completed.
- All task scenarios scored for TCR and SEQ.
- Overall SUS score computed with standard deviation.

---

## 10. Escalation Rules
- Escalate to Product Management if usability testing reveals fundamental mismatch between user mental models and core product value proposition.
- Escalate to Compliance/Legal if session recordings capture unauthorized customer data or PII.
