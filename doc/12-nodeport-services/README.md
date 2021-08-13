# K8s NodePort Services - because workloads outside the cluster need to talk to pods

NodePort services incorporate and extend the functionailty of the ClusterIP service.
NodePort services provide access to the underlying ClusterIP service via a designated high-order port on every worker node.

Upgrade the ClusterIP service to a NodePort service then check the services.
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

So now resources outside our cluster, such as regular EC2 instances inside our VPC, can successfully communicate with our underlying ClusterIP services.

[Return To Main Menu](/README.md)
