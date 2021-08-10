# K8s NodePort Services - because workloads outside the cluster need to talk to pods

Upgrade to a NodePort service which also makes deployment accessible via ANY worker node, then check the services.
```bash
kubectl delete service ${EKS_APP_NAME}
kubectl expose deployment ${EKS_APP_NAME} --port=80 --type=NodePort # this will auto-assign a high-order port on ALL worker nodes
kubectl get services
```

Capture the private IP addresses of the worker nodes and the designated node port for later use (this high-order port this will be in the 30000+ range).
```bash
worker_nodes=($(kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type=="InternalIP")].address}'))
node_port=$(kubectl get service -l app=${EKS_APP_NAME} -o jsonpath='{.items[0].spec.ports[0].nodePort}')
```

Worker nodes will now distribute inbound requests to underlying pods. We curl from inside the jumpbox pod to avoid having to update security groups w.r.t the high-order node port.
```bash
echo ${worker_nodes[0]}:${node_port}
kubectl exec -it jumpbox -- /bin/bash -c "while true; do curl ${worker_nodes[0]}:${node_port}; done"
```

Observe the ec2IP and localhostIP changing with each of the invocations ... these requests were sent to just one of the worker nodes, yet serviced by both of them? That's netfilter/iptables at work. When pods belonging to services are started/stopped, the k8s node-proxy agent on each worker node modifies its routes, creating a kernel-level load balancer per service

[Return To Main Menu](/README.md)
