# K8s ClusterIP Services - because pods need to talk to each other

This section assumes that your `echo-frontend` app is deployed and scaled to 3 instances.

ClusterIP services are intended to establish dynamic communication channels between individual pods inside your cluster so, to see this in action, you first need to gain peer-level access to your workloads, just as you might do with a regular jumpbox (or bastion host) in the EC2 world.
With the `kubectl run` command we can conveniently deploy [nginx](https://www.nginx.com) as a standalone pod which can serve as a jumpbox.
```bash
kubectl run jumpbox --image=nginx                                # in default namespace
sleep 10 && kubectl exec -it jumpbox -- curl http://localhost:80 # <---- test the NGINX welcome page
```

Note that, in the absence of an associated deployment object, your single jumpbox pod will not be automatically replaced in the event of a failure.

Remote into nginx and attempt to demonstrate pod-to-pod communication via a service ... **which will fail** because no such service exists yet.
```bash
kubectl exec -it jumpbox -- curl http://echo-frontend-blue.demos.svc.cluster.local:80 # <---- FAILURE!
```

Upon creation, each service is allocated a long-term **internal** IP address which is scoped to the cluster and auto-registered within private [CoreDNS](https://coredns.io/) servers.
No service means no IP address and, hence, no DNS entry.

Now, using the same templating mechanism employed for the deployment, we can introduce our basic ClusterIP service and test again.
```bash
cat << EOF | tee ~/environment/echo-frontend/templates/echo-frontend-service.yaml | \
             sed "s/{{ .Values.color }}/blue/g" | \
             sed "s/{{ .Values.serviceType }}/ClusterIP/g" | \
             kubectl -n demos apply -f -
apiVersion: v1
kind: Service
metadata:
  name: echo-frontend-{{ .Values.color }}
  labels:
    app: echo-frontend-{{ .Values.color }}
spec:
  type: {{ .Values.serviceType }}
  ports:
  - port: 80
  selector:
    app: echo-frontend-{{ .Values.color }}
EOF
```

Inspect your first service.
```bash
kubectl -n demos get services
```

A private mapping from the DNS name of your service to its corresponding ClusterIP address is now in place so pods can now reach each other via DNS names.
Put the previous curl request in a loop.
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
