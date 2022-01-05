# K8s NodePort Services - because workloads outside the cluster need to talk to pods

NodePort services incorporate and extend the functionailty of the ClusterIP service.
They negotiate access to the underlying service functionality via an auto-assigned high-order port which every worker node agrees upon, hence the name.

We need to upgrade our ClusterIP service to a NodePort service.
There are a couple of ways we could achieve this, including the [kubectl patch](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/) command.
However, in the interests of maintaining focus on manifests, re-apply the service manifest adjusting for the new service type as follows.
```bash
kubectl -n demos get services # inspect service objects before upgrade
cat ~/environment/echo-frontend/templates/echo-frontend-service.yaml | \
    sed "s/{{ .*.color }}/blue/g" | \
    sed "s/{{ .*.serviceType }}/NodePort/g" | \
    kubectl -n demos apply -f -
kubectl -n demos get services # inspect service objects after upgrade
```

The type of your service has now been upgraded to NodePort.
The high-order ports exposed by each worker node will typically be in the 30000+ range.
This helps to minimize the risk of duplicate port usage.
Capture the private IP addresses of all the worker nodes and the high-order port for use in a moment.
```bash
node_ips=($(kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type=="InternalIP")].address}'))
node_port=$(kubectl -n demos get service -l app=echo-frontend-blue -o jsonpath='{.items[0].spec.ports[0].nodePort}')
```

All inbound requests sent to any worker node at `node_ip:node_port` will now be forwarded to the underlying ClusterIP service.
To demo the successful deployment of your NodePort service, `curl` from inside the jumpbox pod.
This access method conveniently avoids having to navigate any security restrictions such as security groups.
```bash
kubectl exec -it jumpbox -- /bin/bash -c "while true; do curl http://${node_ips[0]}:${node_port}; sleep 0.25; done"
# ctrl+c to quit loop
```

Other resources inside your VPC, such as regular EC2 instances or lambda functions, could now successfully communicate with your app.

[Return To Main Menu](/README.md)
