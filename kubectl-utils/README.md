kubectl-utils contains some experimental kubectl extensions that should help us write simpler evals for kubectl-ai

They may one day be useful in their own right, but that is not the current goal.

# kubectl expect

kubectl expect polls an object, waiting for a CEL expression to be true.

Example usage: `kubectl expect StatefulSet/mysql 'self.status.replicas >= 1'`
