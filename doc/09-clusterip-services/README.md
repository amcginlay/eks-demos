# K8s ClusterIP Services - because pods need to talk to each other

This section assumes that the `php-echo` app is deployed and scaled to 3 instances.

Create a second deployment (nginx) for ClusterIP demo purposes.
NOTE we will utilize our nginx deployment as a form of jumpbox, to gain "private" access to our app
```bash
kubectl create deployment nginx --image nginx
kubectl get deployments,pods -o wide                           # two deployments, four pods
sleep 5 && kubectl exec -it $(kubectl get pods -l app=nginx -o jsonpath='{.items[0].metadata.name}') -- curl localhost:80
```

Introduce ClusterIP service (NOTE we remote into nginx here to demonstrate pod-to-pod communication).
This fails, because no such service exists yet ...
```bash
kubectl exec -it $(kubectl get pods -l app=nginx -o jsonpath='{.items[0].metadata.name}') -- curl ${EKS_APP_NAME}:80 # <---- FAILURE!
kubectl get services                                           # our service should not currently exist so delete if present
kubectl expose deployment ${EKS_APP_NAME} --port=80 --type=ClusterIP
kubectl get services
```

Now pods can reach each other via services.
```bash
kubectl exec -it $(kubectl get pods -l app=nginx -o jsonpath='{.items[0].metadata.name}') -- /bin/bash -c "while true; do curl ${EKS_APP_NAME}:80; done"
# ctrl+c to quit loop
```

[Return To Main Menu](/README.md)
