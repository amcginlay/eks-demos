# K8s ClusterIP Services - because pods need to talk to each other

This section assumes that the `EKS_APP_NAME` (i.e. `php-echo`) app is deployed and scaled to 3 instances.

To test ClusterIP services we first need to gain "private" access to our deployments, just as we might do with a regular EC2 jumpbox (or bastion host).
We can deploy [nginx](https://www.nginx.com) as a standalone pod in the default namespace which conveniently suits this purpose.
```bash
kubectl run jumpbox --image=nginx
sleep 10 && kubectl exec -it jumpbox -- curl localhost:80 # <---- test the NGINX welcome page
```

Remote into nginx and attempt to demonstrate pod-to-pod communication via a service ... **which will fail** because no such service exists yet.
```bash
kubectl exec -it jumpbox -- curl ${EKS_APP_NAME}.${EKS_APP_NS}.svc.cluster.local:80 # <---- FAILURE!
```

Upon creation, each service is allocated a long-term internal IP address which is auto-registered within namespace-scoped DNS servers.
No service means no IP address and no DNS entry.

Now we can introduce our basic ClusterIP service and test again.
```bash
kubectl -n ${EKS_APP_NS} get services
kubectl -n ${EKS_APP_NS} expose deployment ${EKS_APP_NAME} --port=80 --type=ClusterIP
kubectl -n ${EKS_APP_NS} get services
```

Note the `CLUSTER-IP` address, then perform a `dig` operation to test the private mapping from the DNS name of the service to its corresponding ClusterIP address.
```bash
kubectl exec -it jumpbox -- /bin/bash -c \
  "apt-get update && apt-get install dnsutils -y && \
  dig +short ${EKS_APP_NAME}.${EKS_APP_NS}.svc.cluster.local"
```

Now pods can reach each other via services.
```bash
kubectl exec -it jumpbox -- /bin/bash -c "while true; do curl ${EKS_APP_NAME}.${EKS_APP_NS}.svc.cluster.local:80; done"
# ctrl+c to quit loop
```

Observe the ec2IP and localhostIP changing with each of the invocations.
These requests were sent to the "jumpbox" pod, which exists on just one of the worker nodes.
However the `EKS_APP_NAME` replicas, which are spread across all the worker nodes, were all involved in servicing the requests.
That's netfilter/iptables at work.
When pods belonging to services are started/stopped, the k8s node-proxy agents on the worker nodes all simultaneously modify their routes, creating a kernel-level load balancer per service.
As a result, it doesn't matter which worker node receives the request, the routing behaviour is consistently well distributed.

We will use the jumpbox pod again so leave it in place for now.

[Return To Main Menu](/README.md)
