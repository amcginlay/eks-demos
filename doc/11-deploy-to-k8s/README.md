# Deploy From ECR To Kubernetes

### `kubectl` manifest generation (a side note)

Kubernetes objects, such as deployments, require manifests in order to be created. `kubectl` supports a number of convenient `create` commands which can build and apply k8s objects whilst avoiding the need to get hands-on with their underlying manifests.

`kubectl create` also supports dry runs, enabling `create` commands to double-up as manifest generators. This behaviour can be observed when executing the following non-destructive command. NOTE the command also makes use of the `kubectl neat` add-on which reduces generated manifests down to their essential elements.
```bash
kubectl create deployment dummy-deployment --image dummy --dry-run=client -o yaml | kubectl neat
```

### Deploy our application at scale

Create a namespace which will host the first deployment then use `kubectl create deployment` to deploy the app from ECR to Kubernetes.
This deployment will start scaled down to zero so we can right-size the CPU requests setting before spinning up the 3 instances.
```bash
kubectl create namespace ${EKS_APP_NS}
kubectl -n ${EKS_APP_NS} create deployment ${EKS_APP_FE} --replicas 0 --image ${EKS_APP_FE_ECR_REPO}:${EKS_APP_FE_VERSION} # begin with zero replicas
kubectl -n ${EKS_APP_NS} set resources deployment ${EKS_APP_FE} --requests=cpu=200m,memory=200Mi                           # right-size the pods
kubectl -n ${EKS_APP_NS} patch deployment ${EKS_APP_FE} --patch="{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"${EKS_APP_FE}\",\"imagePullPolicy\":\"Always\"}]}}}}"
kubectl -n ${EKS_APP_NS} scale deployment ${EKS_APP_FE} --replicas 3                                                       # start 3 instances
sleep 10 && kubectl -n ${EKS_APP_NS} get deployments,pods -o wide                                                          # inspect objects
```

Exec into the first pod to perform curl test.
```bash
first_pod=$(kubectl -n ${EKS_APP_NS} get pods -l app=${EKS_APP_FE} -o name | head -1)
kubectl -n ${EKS_APP_NS} exec -it ${first_pod} -- curl localhost:80
```

Do not delete this deployment. We will need it later.

[Return To Main Menu](/README.md)