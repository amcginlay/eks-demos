# Deploy From ECR To Kubernetes

Kubernetes objects, such as deployments, require manifests in order to be created. `kubectl` supports a number of `create` commands which internalize the creation of manfests such that, for simple use cases, we do not need to see the manifest in order to apply one. `kubectl` also supports dry runs and the ability to export those manifests thereby enabling `kubectl create` commands double-up as manifest generators. This behaviour can be observed when executing the following non-destructive command. NOTE the command also makes use of the `kubectl neat` add-on which simplifies manifests down to their essential elements.
```bash
kubectl create deployment demo --image nginx --dry-run=client -o yaml | kubectl neat
```

Use `kubectl create deployment` to deploy the app from ECR to Kubernetes and scale to 3 pods.
```bash
kubectl create deployment ${EKS_APP_NAME} --image ${EKS_APP_ECR_REPO}:${EKS_APP_VERSION}
kubectl set resources deployment ${EKS_APP_NAME} --requests=cpu=200m # set a reasonable resource allocation (for scaling)
kubectl get deployments,pods -o wide                           # one deployment, one pod
kubectl scale deployment ${EKS_APP_NAME} --replicas 3
kubectl get deployments,pods -o wide                           # one deployment, three pods
```

Exec into the first pod to perform curl test.
```bash
kubectl exec -it $(kubectl get pods -l app=${EKS_APP_NAME} -o jsonpath='{.items[0].metadata.name}') -- curl localhost:80
```

[Return To Main Menu](/README.md)
