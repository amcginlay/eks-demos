# K8s ClusterIP Services - because pods need to talk to each other

This section assumes that the `EKS_APP_NAME` (i.e. `php-echo`) app is deployed and scaled to 3 instances.

To test ClusterIP services we first need to gain "private" access to our deployments, just as we might do with a regular EC2 jumpbox (or bastion host).
We can deploy [nginx](https://www.nginx.com) as a standalone pod which conveniently suits this purpose.
```bash
kubectl -n ${EKS_APP_NAME} run jumpbox --image=nginx
sleep 5 && kubectl -n ${EKS_APP_NAME} exec -it jumpbox -- curl localhost:80 # <---- NGINX welcome page
```

Remote into nginx to demonstrate pod-to-pod communication ... which fails, because no such service exists yet, therefore the DNS lookup will fail.
```bash
kubectl -n ${EKS_APP_NAME} exec -it jumpbox -- curl ${EKS_APP_NAME}:80 # <---- FAILURE!
```

Introduce the service.
```bash
kubectl -n ${EKS_APP_NAME} get services
kubectl -n ${EKS_APP_NAME} expose deployment ${EKS_APP_NAME} --port=80 --type=ClusterIP
kubectl -n ${EKS_APP_NAME} get services
```

Now pods can reach each other via services.
```bash
kubectl -n ${EKS_APP_NAME} exec -it jumpbox -- /bin/bash -c "while true; do curl ${EKS_APP_NAME}:80; done"
# ctrl+c to quit loop
```

We will use the jumpbox pod again so leave it in place for now.

[Return To Main Menu](/README.md)
