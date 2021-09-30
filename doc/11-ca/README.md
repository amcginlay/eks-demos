# Cluster Autoscaler - because no one likes a pending pod

These steps assume you have already deployed the the PHP app we built previously.

Our two-node cluster cannot run an infinite number of pods.
If we scale our deployment beyond the capacity of the current nodes then any pods which cannot be scheduled will be set to a Pending state until circumstances change.
The Cluster Autoscaler monitors for Pending pods that failed to run due to insufficient resources and attempts to resolve the situation by scaling-up the number of available nodes and scaling them back when the pressure recedes.

Install the Cluster Autoscaler.
```bash
kubectl apply -f <( \
  curl --silent https://raw.githubusercontent.com/kubernetes/autoscaler/master/cluster-autoscaler/cloudprovider/aws/examples/cluster-autoscaler-autodiscover.yaml | \
    sed "s/<YOUR CLUSTER NAME>/${EKS_CLUSTER_NAME}\n            - --balance-similar-node-groups\n            - --skip-nodes-with-system-pods=false/g" \
)
```

In a moment we're going to request to increase the number of pods in our existing deployment.
Before we do so, get ready to monitor what is happening.

In a dedicated terminal window prepare to observe the nodes and pods as their status changes.
```bash
watch "kubectl get nodes; echo; kubectl -n ${EKS_NS_BLUE} get pods -o wide"
```

In another dedicated terminal window, begin tailing the Cluster Autoscaler log file to observe as it decides to intervene.
The log is quite noisy so to help pick out the key events the phrase "Scale-up" will be highlighted in red.
```bash
sleep 20 && kubectl logs deployment/cluster-autoscaler -n kube-system -f | grep 'Scale-up\|$' --color
```

Re-scale our deployment to intentionally exceed the capacity of the nodes.
```bash
kubectl -n ${EKS_NS_BLUE} scale deployment ${EKS_APP_NAME} --replicas 30
```

Note how some pods start without an IP addresses because they're stuck in the Pending state and cannot be scheduled.
Once more nodes get added (maximum of 6) the Pending pods will move to a Running state and an IP address will be allocated.
The Cluster Autoscaler will take about 2 minutes to scale-out the nodes and thereby allow all the pods to reach a Running state.

Once all the pods are running, revert their replicset to its previous size.
```bash
kubectl -n ${EKS_NS_BLUE} scale deployment ${EKS_APP_NAME} --replicas 3
```

Best practice suggests that scale-out operations should occur quickly whilst scale-in operations should be slow and graceful.
Eventually, the Cluster Autoscaler will revert to the previous number of nodes.
To save time, manually revert the desired number of nodes and continue to monitor this to completion before moving on.
```bash
eksctl scale nodegroup --cluster ${EKS_CLUSTER_NAME} --name ng-${EKS_CLUSTER_NAME} --nodes 2
```

[Return To Main Menu](/README.md)
