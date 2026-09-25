# Project Intelligence — Reference Guide

## 1. Core Concepts

### 1.1 The Intelligence Ledger
One JSONL line per closed task at evidence/intelligence.jsonl: task id, class, estimate
vs actual for hours and tokens, defect count within 7 days post-merge, reopen flag, and
review rounds. Append-only; corrections are new lines referencing the old one. The ledger
is the organization's memory of how long things really take.

### 1.2 Variance Formula
For each metric:

```
variance_pct = (actual - estimate) / estimate * 100
```

Positive means overrun, negative means underrun. Both are published. A team reporting
only overruns while pocketing underruns trains its calibration on half the data.

### 1.3 Class-Based Calibration
Group tasks into S (under 4 hours), M (4 to 16 hours), L (over 16 hours). The
calibration factor for class C is:

```
calibrated = raw * (1 + mean_variance_C)
```

With class M averaging +38% on hours, a raw 8-hour guess becomes a calibrated 11-hour
commitment. Recalibrate weekly; freeze factors during a release week to avoid planning
churn.

### 1.4 Quality Signals
Cost metrics alone invite speed at the expense of correctness. Every entry also records
defects found within 7 days, reopen flag, and review rounds. A class with low hour
variance but a 30% reopen rate is not predictable; it is rushed.

## 2. Patterns and Anti-Patterns

| Pattern (do this) | Anti-Pattern (never do this) |
|---|---|
| Estimate logged before work, timestamped | "Estimate" written at close-out to match actuals |
| Measured token spend from the ledger | Token guess recalled from memory |
| Signed variance published for every metric | Overruns reported, underruns quietly dropped |
| Calibrate from class means over 5+ tasks | Adjust from one dramatic outlier |
| Tasks measured, classes calibrated | Individuals ranked by variance |

## 3. Worked Example
Task T-342 (class M): estimated 6 hours and 25,000 tokens on 2026-09-20; actuals 8.5
hours and 31,000 tokens; variance +41.7% hours, +24.0% tokens; 1 review round, 0
reopens. Class M mean moves to +38% hours / +21% tokens over 9 tasks, so the next
class-M raw guess of 10 hours calibrates to 13.8 hours, recorded alongside the raw 10.
