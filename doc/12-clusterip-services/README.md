# K8s ClusterIP Services - because pods need to talk to each other

This section assumes that your `echo-frontend` app is deployed and scaled to 3 instances.

### Using `nginx` as a "jumpbox"

ClusterIP services are intended to establish dynamic communication channels between individual pods inside your cluster.
To see this in action and troubleshoot problems it often helps to gain peer-level access to your workloads, just as you might do with a regular [jumpbox](https://en.wikipedia.org/wiki/Jump_server) (or bastion host) in the EC2 world.
With the `kubectl run` command we can conveniently deploy [nginx](https://www.nginx.com) as a standalone pod which will serve as your "jumpbox".
```bash
kubectl run jumpbox --image=nginx                                # in default namespace
sleep 10 && kubectl exec -it jumpbox -- curl http://localhost:80 # <---- test the NGINX welcome page
```

Note that, in the absence of an associated deployment object, your single "jumpbox" pod will not be automatically replaced in the event of a failure.

### Creating a basic service object

Remote into your "jumpbox" and attempt to demonstrate pod-to-pod communication via a service ... **which will fail** because no such service exists yet.
```bash
kubectl exec -it jumpbox -- curl http://echo-frontend-blue.demos.svc.cluster.local:80 # <---- FAILURE!
```

Upon creation, each service is allocated a long-term **internal** IP address which is scoped to the cluster and auto-registered within private [CoreDNS](https://coredns.io/) servers.
No service means no IP address and, hence, no DNS entry.

Before you ask your cluster to deploy the first incarnation of your service object, download its manifest to your Cloud9 environment.
```bash
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-frontend/templates/echo-frontend-service.yaml \
     --directory-prefix ~/environment/echo-frontend/templates/
```

Open `~/environment/echo-frontend/templates/echo-frontend-service.yaml` in Cloud9 IDE to review the code.
Observe that this is using the same templating mechanism employed for the deployment

Now, we can introduce our basic ClusterIP service.
```bash
cat ~/environment/echo-frontend/templates/echo-frontend-service.yaml | \
    sed "s/{{ .Values.color }}/blue/g" | \
    sed "s/{{ .Values.serviceType }}/ClusterIP/g" | \
    kubectl -n demos apply -f -
```

Inspect your first service object.
```bash
kubectl -n demos get services
```

A private mapping from the DNS name of your service to its corresponding ClusterIP address is now in place so pods can now reach each other via DNS names.
Take the previous curl request which failed and place it in a loop.
```bash
kubectl exec -it jumpbox -- /bin/bash -c "while true; do curl http://echo-frontend-blue.demos.svc.cluster.local:80; sleep 0.25; done"
# ctrl+c to quit loop
```

Observe the ec2IP and hostname changing with each of the invocations.
These requests were sent to the jumpbox pod which, as a singleton, exists on just one of the worker nodes.
However the 3 **echo-frontend** replicas, which are spread across all the worker nodes, were each involved in servicing the requests.
That's [netfilter/iptables](https://netfilter.org/) at work.
When pods belonging to services are started/stopped, the **node-proxy** components on the worker nodes all simultaneously modify their routes, creating a consistent kernel-level load balancer per service.
As a result, it doesn't matter which worker node receives the request, the routing behaviour is consistently well distributed.

We will use the jumpbox pod again so leave it in place for now.

[Return To Main Menu](/README.md)
