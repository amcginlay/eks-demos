# Cluster Autoscaler - because no one likes a pending pod

Our two-node cluster cannot run an infinite number of pods.
If we scale our deployment beyond the capacity of the current nodes then any pods which cannot be scheduled will be set to a Pending state until circumstances change.
The Cluster Autoscaler monitors for pods that failed to run due to insufficient resources and attempts to resolve the situation by scaling-up the number of available nodes and scaling them back when the pressure recedes.

Install the Cluster Autoscaler:
```bash
kubectl apply -f <( \
  curl --silent https://raw.githubusercontent.com/kubernetes/autoscaler/master/cluster-autoscaler/cloudprovider/aws/examples/cluster-autoscaler-autodiscover.yaml | \
    sed "s/<YOUR CLUSTER NAME>/${EKS_CLUSTER_NAME}\n            - --balance-similar-node-groups\n            - --skip-nodes-with-system-pods=false/g" \
)
```

TODO ...

# in a dedicated terminal window, tail the logs to witness the EC2 scaling
kubectl logs deployment/cluster-autoscaler -n kube-system -f | grep 'Scale-up\|$' --color

# in another dedicated terminal window, tail the logs to witness the EC2 scaling
watch kubectl get pods -n ${EKS_NS_BLUE} -o wide

# now re-scale our deployment beyond the capacity of the nodes
kubectl -n ${EKS_NS_BLUE} scale deployment ${EKS_APP_NAME} --replicas 30

# notice how some pods start without an ip addresses because they're stuck in the Pending state.
# once more nodes get added (maximum of 6) the pending pods move to a running state

# revert the changes
kubectl -n ${EKS_NS_BLUE} scale deployment ${EKS_APP_NAME} --replicas 3
eksctl scale nodegroup --cluster ${EKS_CLUSTER_NAME} --name ng-${EKS_CLUSTER_NAME} --nodes 2


[Return To Main Menu](/README.md)
