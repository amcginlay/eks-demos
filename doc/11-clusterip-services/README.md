# K8s ClusterIP Services - because pods need to talk to each other

This section assumes that the `EKS_APP_NAME` (i.e. `php-echo`) app is deployed and scaled to 3 instances.

To test ClusterIP services we first need to gain "private" access to our deployments, just as we might do with a regular EC2 jumpbox (or bastion host).
We can deploy [nginx](https://www.nginx.com) as a standalone pod which conveniently suits this purpose.
```bash
kubectl -n ${EKS_APP_NAME} run jumpbox --image=nginx
sleep 5 && kubectl -n ${EKS_APP_NAME} exec -it jumpbox -- curl localhost:80 # <---- test the NGINX welcome page
```

Remote into nginx and attempt to demonstrate pod-to-pod communication via a service ... **which will fail** because no such service exists yet.
```bash
kubectl -n ${EKS_APP_NAME} exec -it jumpbox -- curl ${EKS_APP_NAME}:80 # <---- FAILURE!
```

Upon creation, each service is allocated a long-term internal IP address which is auto-registered within namespace-scoped DNS servers.
No service means no IP address and no DNS entry.

Now we can introduce our basic ClusterIP service and test again.
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

Observe from the responses how the service distributes requests across all of the active pod replicas (see `localhostIP`)

We will use the jumpbox pod again so leave it in place for now.

[Return To Main Menu](/README.md)
