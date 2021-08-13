# K8s NodePort Services - because workloads outside the cluster need to talk to pods

NodePort services incorporate and extend the functionailty of the ClusterIP service.
NodePort services provide access to the underlying ClusterIP service via a designated high-order port on every worker node.

Upgrade to a NodePort service then check the services.
```bash
kubectl -n ${EKS_APP_NAME} delete service ${EKS_APP_NAME}
kubectl -n ${EKS_APP_NAME} expose deployment ${EKS_APP_NAME} --port=80 --type=NodePort # this will auto-assign a high-order port on ALL worker nodes
kubectl -n ${EKS_APP_NAME} get services
```

Capture the private IP addresses of the worker nodes and the designated port for later use (this high-order port this will be in the 30000+ range).
```bash
worker_nodes=($(kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type=="InternalIP")].address}'))
node_port=$(kubectl -n ${EKS_APP_NAME} get service -l app=${EKS_APP_NAME} -o jsonpath='{.items[0].spec.ports[0].nodePort}')
```

All worker nodes will now forward inbound requests on the designated port to the underlying ClusterIP service.
We `curl` from inside the jumpbox pod (created when we built the original ClusterIP service) to avoid having to update security groups in respect of the node port.
```bash
echo ${worker_nodes[0]}:${node_port}
kubectl -n ${EKS_APP_NAME} exec -it jumpbox -- /bin/bash -c "while true; do curl ${worker_nodes[0]}:${node_port}; done"
```

Observe the ec2IP and localhostIP changing with each of the invocations ... these requests were sent to just one of the worker nodes, yet serviced by both of them? That's netfilter/iptables at work. When pods belonging to services are started/stopped, the k8s node-proxy agent on each worker node modifies its routes, creating a kernel-level load balancer per service

[Return To Main Menu](/README.md)
