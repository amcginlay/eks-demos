# Deploy From ECR To Kubernetes

### `kubectl` manifest generation (a side note)

Kubernetes objects, such as deployments, require manifests in order to be created. `kubectl` supports a number of `create` commands which can conveniently construct and apply simple k8s objects whilst obviating the need to observe or persist their manifests.

`kubectl` also supports dry runs, thereby enabling those `create` commands to double-up as manifest generators. This behaviour can be observed when executing the following non-destructive command. NOTE the command also makes use of the `kubectl neat` add-on which simplifies generated manifests down to their essential elements.
```bash
kubectl create deployment dummy-deployment --image dummy --dry-run=client -o yaml | kubectl neat
```

### Deploy our application at scale

Create a `blue` namespace which will host the first deployment.
```bash
kubectl create namespace ${EKS_APP_NAME}-blue
```

Use `kubectl create deployment` to deploy the app from ECR to Kubernetes and scale to 3 pods.
```bash
kubectl -n ${EKS_APP_NAME}-blue create deployment ${EKS_APP_NAME} --image ${EKS_APP_ECR_REPO}:${EKS_APP_VERSION}
sleep 10 && kubectl -n ${EKS_APP_NAME}-blue get all -o wide               # one deployment, one pod
kubectl -n ${EKS_APP_NAME}-blue scale deployment ${EKS_APP_NAME} --replicas 3
kubectl -n ${EKS_APP_NAME}-blue get all -o wide                           # one deployment, three pods
```

Exec into the first pod to perform curl test.
```bash
first_pod=$(kubectl -n ${EKS_APP_NAME}-blue get pods -l app=${EKS_APP_NAME} -o name | head -1)
kubectl -n ${EKS_APP_NAME}-blue exec -it ${first_pod} -- curl localhost:80
```

Do not delete this deployment. We will need it later.

[Return To Main Menu](/README.md)
