#!/usr/bin/env bash
# Wait for deployment to scale to 2 replicas with kubectl wait
if kubectl wait --for=condition=Available=True --timeout=30s deployment/web-app -n scale-test; then
    # Verify the replica count is exactly 2
    if [ "$(kubectl get deployment web-app -n scale-test -o jsonpath='{.status.availableReplicas}')" = "2" ]; then
        exit 0
    fi
fi

# If we get here, deployment didn't scale up correctly in time
exit 1 