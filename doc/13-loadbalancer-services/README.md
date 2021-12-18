# K8s LoadBalancer Services - because the world needs to talk to our cluster

The NodePort service solved one problem but exposed another.
If our worker nodes (i.e. EC2 instances) belong to an auto-scaling group then some questions remain:

- How many nodes are active at any given moment and how do we track their IP addresses?
- Are these nodes reachable by your clients?
- What if the node we were targeting a moment ago has now been terminated?

In general, load balancers are designed to negate these questions by providing a single point of access which guards us from the underlying complexity and volatility.
Kubernetes services of type LoadBalancer incorporate and extend the functionailty of NodePort services.
In EKS, services of type LoadBalancer provide access to the underlying functionality via an [AWS Classic Load Balancer (CLB)](https://aws.amazon.com/elasticloadbalancing/classic-load-balancer).

Kubernetes clusters provide extensibilty via pluggable components known as [controllers](https://kubernetes.io/docs/concepts/architecture/controller/) which can be customized for a variety of purposes.
Controllers, which implement the [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/), take their cues from type-matched object definitions.
Whenever these objects get applied to the cluster the associated controller will react by taking whatever action necessary to reconcile.
It is the same pattern of behaviour you have seen for all native Kubernetes objects so this reconciliation model (i.e. desired vs. actual) should feel familiar.
Some controllers have the capability to reach out beyond their cluster and into their surrounding environment (e.g. AWS) to satisfy the desired state.
Use of these controllers can help the developer remain focused on a single set of tools (i.e. kubectl and manifests) to deploy both their applications and the wider set of associated infrastructure resources.
In EKS, services of type LoadBalancer are supported by this type of custom controller implementation.

Upgrade the NodePort service to a LoadBalancer service, then check the running services.
```bash
kubectl -n demos get services # inspect services before upgrade
sed -i "s/NodePort/LoadBalancer/g" ~/environment/echo-frontend-1.0/manifests/echo-frontend-service.yaml
kubectl apply -f ~/environment/echo-frontend-1.0/manifests/echo-frontend-service.yaml
sleep 5 && kubectl -n demos get services # inspect services after upgrade
```

The `EXTERNAL-IP`, which was previously set as `<none>`, now contains the publicly addressible DNS name for an AWS Classic Load Balancer (CLB).
Port 80 requests arriving at this endpooint are now evenly distributed across all worker nodes using their node port as before.
Grab the CLB DNS name and put the following `curl` command in a loop as the AWS resource will not be immediately resolved (2-3 mins).
If you receive any errors, just wait a little longer.
```bash
clb_dnsname=$(kubectl -n demos get service echo-frontend -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${clb_dnsname}; sleep 0.25; done
# ctrl+c to quit loop
```

It is important to recognise the shared DNA that runs through the Kubernetes service types.
Services of type **LoadBalancer** inherit from **NodePort** services which, in turn, inherit from **ClusterIP** services.

[Return To Main Menu](/README.md)
