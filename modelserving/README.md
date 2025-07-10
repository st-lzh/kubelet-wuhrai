# Model Serving

This directory provides components to build and deploy Large Language Model (LLM) serving endpoints.

- [`k8s/`](k8s/): Kubernetes manifests for model serving components.
- [`images/`](images/): Dockerfiles for building model serving container images.
- [`dev/tasks`](dev/tasks): Development-related scripts for model serving.
  - `download-model`: fetch the required model weights (e.g., Gemma 3 12B IT).
  - `build-images`: runs `download-model`, and then build the Docker image using the provided Dockerfile in `images/`.
  - `deploy-to-gke` or `dev/tasks/deploy-to-kind`: runs `build-images`, and then deploy the model serving Kubernetes manifests to Google Kubernetes Engine (GKE) or a local KinD cluster. Once deployed, the model server will be accessible via a Kubernetes Service defined in the manifest. You can use `kubectl get svc` to find the service details and access its endpoint.
  - `run-local`: run the model server locally for testing purposes, bypassing Kubernetes.