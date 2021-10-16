# Deploy From ECR To Kubernetes

### `kubectl` manifest generation (a side note)

Kubernetes objects, such as deployments, require manifests in order to be created. `kubectl` supports a number of convenient `create` commands which can build and apply k8s objects whilst avoiding the need to get hands-on with their underlying manifests.

`kubectl` also supports dry runs, thereby enabling those `create` commands to double-up as manifest generators. This behaviour can be observed when executing the following non-destructive command. NOTE the command also makes use of the `kubectl neat` add-on which reduces generated manifests down to their essential elements.
```bash
kubectl create deployment dummy-deployment --image dummy --dry-run=client -o yaml | kubectl neat
```

### Deploy our application at scale

Create a `blue` namespace which will host the first deployment then use `kubectl create deployment` to deploy the app from ECR to Kubernetes.
This deployment will start scaled down to zero so we can right-size the CPU requests setting before spinning up the 3 instances.
```bash
kubectl create namespace ${EKS_NS_BLUE}
kubectl -n ${EKS_NS_BLUE} create deployment ${EKS_APP_NAME} --replicas 0 --image ${EKS_APP_ECR_REPO}:${EKS_APP_VERSION} # begin with zero replicas
kubectl -n ${EKS_NS_BLUE} set resources deployment ${EKS_APP_NAME} --requests=cpu=200m,memory=200Mi                     # right-size the pods
kubectl -n ${EKS_NS_BLUE} scale deployment ${EKS_APP_NAME} --replicas 3                                                 # start 3 instances
sleep 10 && kubectl -n ${EKS_NS_BLUE} get all -o wide                                                                   # inspect objects
```

Exec into the first pod to perform curl test.
```bash
first_pod=$(kubectl -n ${EKS_NS_BLUE} get pods -l app=${EKS_APP_NAME} -o name | head -1)
kubectl -n ${EKS_NS_BLUE} exec -it ${first_pod} -- curl localhost:80
```

Do not delete this deployment. We will need it later.

[Return To Main Menu](/README.md)
