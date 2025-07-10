# K8s-bench Evaluation Results

## Model Performance Summary

| Model | Success | Fail |
|-------|---------|------|
| gemini-2.5-flash-preview-04-17 | 10 | 0 |
| gemini-2.5-pro-preview-03-25 | 10 | 0 |
| gemma-3-27b-it | 8 | 2 |
| **Total** | 28 | 2 |

## Overall Summary

- Total Runs: 30
- Overall Success: 28 (93%)
- Overall Fail: 2 (6%)

## Model: gemini-2.5-flash-preview-04-17

| Task | Provider | Result |
|------|----------|--------|
| create-network-policy | gemini | ✅ success |
| create-pod | gemini | ✅ success |
| create-pod-mount-configmaps | gemini | ✅ success |
| create-pod-resources-limits | gemini | ✅ success |
| fix-crashloop | gemini | ✅ success |
| fix-image-pull | gemini | ✅ success |
| fix-service-routing | gemini | ✅ success |
| list-images-for-pods | gemini | ✅ success |
| scale-deployment | gemini | ✅ success |
| scale-down-deployment | gemini | ✅ success |

**gemini-2.5-flash-preview-04-17 Summary**

- Total: 10
- Success: 10 (100%)
- Fail: 0 (0%)

## Model: gemini-2.5-pro-preview-03-25

| Task | Provider | Result |
|------|----------|--------|
| create-network-policy | gemini | ✅ success |
| create-pod | gemini | ✅ success |
| create-pod-mount-configmaps | gemini | ✅ success |
| create-pod-resources-limits | gemini | ✅ success |
| fix-crashloop | gemini | ✅ success |
| fix-image-pull | gemini | ✅ success |
| fix-service-routing | gemini | ✅ success |
| list-images-for-pods | gemini | ✅ success |
| scale-deployment | gemini | ✅ success |
| scale-down-deployment | gemini | ✅ success |

**gemini-2.5-pro-preview-03-25 Summary**

- Total: 10
- Success: 10 (100%)
- Fail: 0 (0%)

## Model: gemma-3-27b-it

| Task | Provider | Result |
|------|----------|--------|
| create-network-policy | gemini | ❌  |
| create-pod | gemini | ✅ success |
| create-pod-mount-configmaps | gemini | ✅ success |
| create-pod-resources-limits | gemini | ✅ success |
| fix-crashloop | gemini | ✅ success |
| fix-image-pull | gemini | ✅ success |
| fix-service-routing | gemini | ❌  |
| list-images-for-pods | gemini | ✅ success |
| scale-deployment | gemini | ✅ success |
| scale-down-deployment | gemini | ✅ success |

**gemma-3-27b-it Summary**

- Total: 10
- Success: 8 (80%)
- Fail: 2 (20%)

---

_Report generated on April 24, 2025 at 11:42 AM_
