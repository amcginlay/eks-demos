# K8s LoadBalancer Services - because the world needs to talk to our cluster

LoadBalancer services incorporate and extend the functionailty of the NodePort service.
They provide access to the underlying NodePort service via a classic AWS load balancer.

The NodePort service solved one problem but exposed another.
If our EC2 instances belong to a functioning auto-scaling group then how can we know how many are currently active and what their private IP addresses are.
A load balancer is designed to solve this exact type of problem.

Kubernetes components, known as controllers, can consume applied manifests and perform whatever tasks are necessary to reconcile that which is desired against that which currently exists. Sometimes the required reconciliation actions require the controller to reach beyond the cluster and in the host environment. LoadBalancer services are a classic example of this pattern which is designed to keep the developer focused on one set of tools to deploy both their applications and the wider infrastructure components used.

Upgrade the NodePort service to a LoadBalancer service, then check the services.
```bash
kubectl -n ${EKS_APP_NAME} delete service ${EKS_APP_NAME}
kubectl -n ${EKS_APP_NAME} expose deployment ${EKS_APP_NAME} --port=80 --type=LoadBalancer
kubectl -n ${EKS_APP_NAME} get service
```

External port 80 requests are now load balanced across the node port of all worker nodes. Grab the load balancer DNS name and put the following curl command in a loop as the AWS resource will not be immediately resolved (2-3 mins). If you receive any `curl` errors, just wait a little longer.
```bash
lb_dnsname=$(kubectl -n ${EKS_APP_NAME} get service -l app=${EKS_APP_NAME} -o jsonpath='{.items[0].status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${lb_dnsname}; done
```

[Return To Main Menu](/README.md)
