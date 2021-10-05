# K8s LoadBalancer Services - because the world needs to talk to our cluster

LoadBalancer services incorporate and extend the functionailty of the NodePort service which, in turn, extends the functionality of the ClusterIP service.
They provide access to the underlying NodePort service via a classic AWS load balancer.

The NodePort service solved one problem but exposed another.
If our EC2 instances belong to a functioning auto-scaling group then this raises some important questions:

- How many EC2 instances are currently active?
- What their private IP addresses are?
- What if the EC2 instance we're currently targeting is terminated?

A load balancer is designed to solve this exact type of problem by providing a single point of access which guards us from the underlying complexity.

Kubernetes components, known as controllers, can react to the presence of specific kinds of object instances and perform whatever tasks are necessary to reconcile the system state which is desired against that which currently exists. Sometimes the reconciliation actions require the controller to reach out beyond the cluster and into the host environment itself - in this case AWS. LoadBalancer services are a classic example of this pattern which is designed to help the developer remain focused on one set of tools to deploy both their applications and the wider set infrastructure components used.

Upgrade the NodePort service to a LoadBalancer service, then check the services.
```bash
kubectl -n ${EKS_NS_BLUE} delete service ${EKS_APP_NAME}
kubectl -n ${EKS_NS_BLUE} expose deployment ${EKS_APP_NAME} --port=80 --type=LoadBalancer
sleep 5 && kubectl -n ${EKS_NS_BLUE} get service
```

External port 80 requests are now load balanced across the negotiated node port of all worker nodes. Grab the load balancer DNS name and put the following curl command in a loop as the AWS resource will not be immediately resolved (2-3 mins). If you receive any `curl` errors, just wait a little longer.
```bash
lb_dnsname=$(kubectl -n ${EKS_NS_BLUE} get service ${EKS_APP_NAME} -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${lb_dnsname}; sleep 0.25; done
```

The [AWS Classic Load Balancer](https://aws.amazon.com/elasticloadbalancing/classic-load-balancer) is being used here but it's a little too basic for our needs so, before we move on, **downgrade** back to a NodePort service then check the services.
This may take a few seconds to complete.
```bash
kubectl -n ${EKS_NS_BLUE} delete service ${EKS_APP_NAME}
kubectl -n ${EKS_NS_BLUE} expose deployment ${EKS_APP_NAME} --port=80 --type=NodePort
kubectl -n ${EKS_NS_BLUE} get services
```

[Return To Main Menu](/README.md)
