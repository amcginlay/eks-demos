# Deploy From ECR To Kubernetes

Kubernetes objects, such as deployments, require manifests in order to be created. `kubectl` supports a number of `create` commands which internalize the manifest creation so that, for the simplest of use cases, we never need to see the manifest. `kubectl` also supports dry runs and the ability to export those manifests such that `kubectl create` commands double-up as manifest generators, as demonstrated by the following command. NOTE the `kubectl neat` add-on reduces the generated manifest down to its essential elements.
```bash
kubectl create deployment demo --image nginx --dry-run=client -o yaml | kubectl neat
```

Use `kubectl create deployment` to deploy the app from ECR to Kubernetes and scale to 3 pods.
```bash
kubectl create deployment ${APP_NAME} --image ${APP_ECR_REPO}:${APP_VERSION}
kubectl set resources deployment ${APP_NAME} --requests=cpu=200m # set a reasonable resource allocation (for scaling)
kubectl get deployments,pods -o wide                           # one deployment, one pod
kubectl scale deployment ${app_name} --replicas 3
kubectl get deployments,pods -o wide                           # one deployment, three pods
```

Exec into the first pod to perform curl test.
```bash
kubectl exec -it $(kubectl get pods -l app=${APP_NAME} -o jsonpath='{.items[0].metadata.name}') -- curl localhost:80
```

[Return To Main Menu](/README.md)
