# Deploy From ECR To Kubernetes

Kubernetes objects, such as deployments, require manifests in order to be created. `kubectl` supports a number of `create` commands which can conveniently build simple objects whilst internalizing the complex detail of the manifest itself and obviating the need to have those manifests persisted to disk.

`kubectl` also supports dry runs, thereby enabling those `create` commands to double-up as manifest generators. This behaviour can be observed when executing the following non-destructive command. NOTE the command also makes use of the `kubectl neat` add-on which simplifies generated manifests down to their essential elements.
```bash
kubectl create deployment demo --image nginx --dry-run=client -o yaml | kubectl neat
```

Use `kubectl create deployment` to deploy the app from ECR to Kubernetes and scale to 3 pods.
```bash
kubectl create deployment ${EKS_APP_NAME} --image ${EKS_APP_ECR_REPO}:${EKS_APP_VERSION}
kubectl set resources deployment ${EKS_APP_NAME} --requests=cpu=200m # set a reasonable resource allocation (for scaling)
sleep 10 && kubectl get deployments,pods -o wide                     # one deployment, one pod
kubectl scale deployment ${EKS_APP_NAME} --replicas 3
kubectl get deployments,pods -o wide                                 # one deployment, three pods
```

Exec into the first pod to perform curl test.
```bash
kubectl exec -it $(kubectl get pods -l app=${EKS_APP_NAME} -o name | head -1) -- curl localhost:80
```

Do not delete this deployment. We will need it later.

[Return To Main Menu](/README.md)
