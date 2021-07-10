# K8s LoadBalancer Services - because the world needs to talk to our cluster

Upgrade to a LoadBalancer service which also supports external request distribution over an AWS ELB, then check the services.
```bash
kubectl delete service ${EKS_APP_NAME}
kubectl expose deployment ${EKS_APP_NAME} --port=80 --type=LoadBalancer # note this new service will automatically re-assign the high-order node port
kubectl get service
```

External port 80 requests are now load balanced across the node ports of all worker nodes. Grab the load balancer DNS name and put the following curl command in a loop as the AWS resource will not be immediately resolved (2-3 mins)
```bash
lb_dnsname=$(kubectl get service -l app=${EKS_APP_NAME} -o jsonpath='{.items[0].status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${lb_dnsname}; done                 # ... now try this HTTP URL from a browser and ctrl+c to quit loop
```

[Return To Main Menu](/README.md)
