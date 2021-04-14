## Deployment
This directory contains the kubernetes resources used to deploy selections.
This should make redeploying selections should the namespace be deleted trivial.

Note that the secrets are not committed to github. When deploying, you will need to ensure the following exist:

```yaml
apiVersion: v1
items:
- apiVersion: v1
  kind: ConfigMap
  # Define environment variables for dev
  metadata:
    name: dev
    namespace: selections
- apiVersion: v1
  kind: ConfigMap
  # Define environment variables for prod
  metadata:
    name: prod
    namespace: selections
- apiVersion: v1
  kind: Secret
  # Github webhook secret
  metadata:
    name: prod-github
    namespace: selections
- apiVersion: v1
  kind: Secret
  # Github webhook secret
  metadata:
    name: dev-github
    namespace: selections
kind: List
```

Creating the selections project should only require the following (after you've created either `secrets.yaml` or the secrets resources):
```bash
oc new-project selections
oc create -f *.yaml
```
