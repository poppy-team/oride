# Usability Testing Metrics & UXR Methodology: Technical Reference Guide

## 1. Quantitative Usability Metrics

### 1.1 System Usability Scale (SUS)
Created by John Brooke in 1986, SUS is the industry-standard benchmark for perceived usability.

#### The 10 Standard Questions (Rated 1 = Strongly Disagree to 5 = Strongly Agree):
1. I think that I would like to use this system frequently.
2. I found the system unnecessarily complex.
3. I thought the system was easy to use.
4. I think that I would need the support of a technical person to be able to use this system.
5. I found the various functions in this system were well integrated.
6. I thought there was too much inconsistency in this system.
7. I would imagine that most people would learn to use this system very quickly.
8. I found the system very cumbersome to use.
9. I felt very confident using the system.
10. I needed to learn a lot of things before I could get going with this system.

#### Mathematical Calculation Formula
For each user response:
- For **odd-numbered items** (positive statements: 1, 3, 5, 7, 9):
  $$\text{Item Score} = \text{Response} - 1$$
- For **even-numbered items** (negative statements: 2, 4, 6, 8, 10):
  $$\text{Item Score} = 5 - \text{Response}$$
- Sum all 10 item scores and multiply by 2.5:
  $$\text{SUS Score} = \left(\sum_{i=1}^{10} \text{Item Score}_i\right) \times 2.5$$

#### Interpretation Benchmarks (Sauro & Lewis)
- **Score $\ge 80.3$ (Grade A)**: Top 10% exceptional usability.
- **Score $68.0$ (Grade C)**: Global industry average.
- **Score $< 51.0$ (Grade F)**: Critical usability failure; high likelihood of user abandonment.

---

### 1.2 Single Ease Question (SEQ)
Administered immediately following the completion or abandonment of each discrete task:
> *"Overall, how easy or difficult was it to complete this task?"*
> `[1: Very Difficult]  [2]  [3]  [4: Neutral]  [5]  [6]  [7: Very Easy]`
- Global benchmark average: **5.5**. Tasks scoring below 5.0 indicate acute user friction.

### 1.3 Task Completion Rate (TCR)
$$\text{TCR} = \frac{\text{Successful Completions}}{\text{Total Task Attempts}} \times 100\%$$
- Benchmark standard: Target $\ge 80\%$ for unassisted core user workflows.

---

## 2. Qualitative Facilitation: Think-Aloud Protocol

The Think-Aloud method requires participants to verbalize their thoughts, expectations, and reasoning continuously while attempting tasks.

### Facilitator Rules
1. **Never Answer Interface Questions**: If the user asks *"Should I click here?"*, respond with: *"What would you expect to happen if you clicked there?"*
2. **Prompts for Silence**: If a participant remains silent for more than 5 seconds, use neutral prompts: *"Tell me what you're looking at right now"*, or *"What are you thinking?"*
3. **Avoid Leading Phrasing**: Do not mention UI element labels in the scenario description.
