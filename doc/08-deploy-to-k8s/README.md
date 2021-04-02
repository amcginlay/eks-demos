# Deploy From ECR To Kubernetes

Kubernetes objects, such as deployments, require manifests in order to be created. 

Deploy the app to Kubernetes and scale to 3 pods.
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
