# Cluster Autoscaler - because no one likes a pending pod

This section assumes that the `echo-frontend` app is deployed and scaled to 3 instances.

Our two-node cluster cannot run an infinite number of pods.
If you scale your deployment beyond the capacity of the current nodes then any pods which cannot be scheduled will be set to a Pending state until circumstances change.
The [Cluster Autoscaler (CA)](https://docs.aws.amazon.com/eks/latest/userguide/cluster-autoscaler.html) monitors for Pending pods that failed to run due to insufficient resources and attempts to resolve the situation by scaling-up the number of available nodes and scaling them back when the pressure recedes.

Install the CA following documented best practices for EKS.
```bash
kubectl apply -f <( \
  curl --silent https://raw.githubusercontent.com/kubernetes/autoscaler/master/cluster-autoscaler/cloudprovider/aws/examples/cluster-autoscaler-autodiscover.yaml | \
    sed "s/v1\.[[:digit:]]*\.0/v${EKS_K8S_VERSION}.0/g" | \
    sed "s/<YOUR CLUSTER NAME>/${C9_PROJECT}\n            - --balance-similar-node-groups\n            - --skip-nodes-with-system-pods=false/g" \
)
```

In a moment you will submit a request to increase the number of pods in your existing deployment.
Before you do so, get ready to monitor what is happening inside your cluster.

In a **dedicated** terminal window prepare to observe the nodes and pods as their status changes.
```bash
watch "kubectl get nodes; echo; kubectl -n demos get pods -o wide"
```

In another **dedicated** terminal window, wait a moment before following the CA log to observe as it decides to intervene.
Scaling related events will be highlighted in red.
```bash
sleep 20 && kubectl logs deployment/cluster-autoscaler -n kube-system -f | grep 'scale-up\|scaleup\|scale up\|$' --color
```

In your original terminal window, re-scale your deployment to intentionally exceed the capacity of the nodes.
```bash
kubectl -n demos scale deployment echo-frontend-blue --replicas 25
```

Note how some pods start without an IP addresses because they're stuck in the Pending state and cannot be scheduled.
Once additional nodes get introduced (up to the current maximum of 6) the Pending pods will move to a Running state and an IP address will be allocated.
The CA will take **about 2 minutes** to complete the node scaling operation and thereby allow all the pods to start.

Once all the pods are successfully in a Running state and have IP addresses assigned the demo is complete.
Revert the replicaset to its previous size.
```bash
kubectl -n demos scale deployment echo-frontend-blue --replicas 3
```

If nodes are in Ready state and have spare capacity, new pods can be created in a matter of seconds.
In comparison, nodes take significantly longer to stand up so best practice indicates that the CA is to scale-out as rapidly as possible and scale-in slowly in order to cope with the possibility of spiky workloads.
With the number of pod replicas reduced, the CA will slowly scale-in the number of nodes and would eventually revert to its previous size.
This operation would take 10+ minutes so, to save time, manually revert the desired number of nodes and continue to monitor this to completion before moving on.
```bash
eksctl scale nodegroup --cluster ${C9_PROJECT} --name mng --nodes 2
```

Next: [Main Menu](/README.md) | [Horizonal Pod Autoscaler](../17-hpa/README.md)
