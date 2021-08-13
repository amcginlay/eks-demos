# K8s LoadBalancer Services - because the world needs to talk to our cluster

The NodePort service solves one problem but exposes another.
If our worker nodes belong to a functioning auto-scaling group then we can never really know for sure how many there are or what their private IP addresses are.
If this was a regular fleet of EC2s we would now introduce a load balancer, and this situation is no different.

Upgrade the NodePort service to a LoadBalancer service which supports external request distribution over an AWS ELB, then check the services.
```bash
kubectl -n ${EKS_APP_NAME} delete service ${EKS_APP_NAME}
kubectl -n ${EKS_APP_NAME} expose deployment ${EKS_APP_NAME} --port=80 --type=LoadBalancer
kubectl -n ${EKS_APP_NAME} get service
```

External port 80 requests are now load balanced across the node ports of all worker nodes. Grab the load balancer DNS name and put the following curl command in a loop as the AWS resource will not be immediately resolved (2-3 mins). If you receive `curl: (3) Bad URL` just wait a little longer.
```bash
lb_dnsname=$(kubectl get service -l app=${EKS_APP_NAME} -o jsonpath='{.items[0].status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${lb_dnsname}; done
```

[Return To Main Menu](/README.md)
