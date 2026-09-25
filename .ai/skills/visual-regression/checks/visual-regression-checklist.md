# Automated Visual Regression Testing Checklist

- [ ] Snapshot tests executed in deterministic environment / container
- [ ] Dynamic elements (dates, avatars, counters) masked out
- [ ] CSS animations and transitions disabled during snapshot capture
- [ ] Snapshots taken only after fonts.ready and images finish loading
- [ ] Strict pixel-diff thresholds configured (maxDiffPixelRatio <= 0.002)
- [ ] Visual failure reports generate baseline, current, and diff images
- [ ] Baseline updates require explicit git commit and review approval
