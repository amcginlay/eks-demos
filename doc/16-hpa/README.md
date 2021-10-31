# Horizontal Pod Autoscaler - because demand for pods can grow

This next section assumes that you you have completed the earlier section on **LoadBalancer services** and have a load balancer available.

When your workloads come under pressure their CPU consumption will rise.
Cloud native best practices suggest that the response to this situation is to increase the number of workload replicas which spreads the load.
We have already seen how **manually** increasing the number of pod replicas will cause the Cluster Autoscaler (CA) to attempt to add new nodes to to the cluster.
But what if your workloads come under CPU pressure and no one is present to make this adjustment?

The [Horizontal Pod Autoscaler (HPA)](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) aims to right-size the number of replicas in your deployments as a  reaction to realtime changes in workload pressure.
For a given deployment, you can create an HPA object which specifies the min/max limits for the `replicas` attribute and a target CPU percentage.
From that point the HPA will, within its limits, continually evaluate the CPU pressure on your deployment and adjust the `replicas` attribute to suit.

Although the HPA is natively installed, it depends upon the [Kubernetes Metrics Server](https://github.com/kubernetes-sigs/metrics-server) which is not.
A good way to test if the Metrics Server is missing is to run the following `top` command.
Right now this **will fail**.
```bash
kubectl top nodes
```

Fix the error by installing the Metrics Server.
Note you may need to adjust `latest` for a more specific release, but this usually works. 
```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

Confirm the Metrics Server was installed.
```bash
kubectl get deployment metrics-server -n kube-system
```

The newly installed Metrics Server may take a minute to begin producing results.
Use a `watch` command to observe as the Metrics Server comes online.
```bash
watch "kubectl top nodes; echo; kubectl top pods --all-namespaces"
# ctrl+c to quit watch command
```

The command to activate an HPA for an existing deployment is `autoscale`.
Activate the Horizontal Pod Autoscaler for our running deployment.
```bash
kubectl -n ${EKS_APP_NS} autoscale deployment ${EKS_APP} --cpu-percent=50 --min=3 --max=25
```

Keep watching the k8s objects in a **dedicated** terminal window.
```bash
watch "kubectl get nodes; echo; kubectl -n ${EKS_APP_NS} get deployments,hpa,pods -o wide"
```

In another **dedicated** terminal window, use siege to put the app under heavy load.
```bash
clb_dnsname=$(kubectl -n ${EKS_APP_NS} get service -l app=${EKS_APP} -o jsonpath='{.items[0].status.loadBalancer.ingress[0].hostname}')
siege -c 5 ${clb_dnsname}                                    # simulate 5 concurrent users
```

This will cause the HPA to autoscale the pods upwards towards its max setting.
Switch back the terminal window displaying the `watch` results, paying attention to the HPA values shown under TARGETS / REPLICAS.
Also, observe the list of pods as it grows.
Under heavy load the TARGET ratio will be high and the number of replicas will increase rapidly until the resources in the nodes are exhausted.
At this point you will observe that new pod IP addresses are unable to be allocated and a number of pods will remain in a Pending state.
As seen in an earlier section, the CA will shortly step in to add new nodes and within a couple of minutes all 25 pods will be Running as expected.

Switch back the terminal window displaying the `siege` results.
Instead of stopping the load entirely, which is a little unnatural, we are just going to dial it back from high to low.
```bash
# ctrl+c to quit the high-load siege command
siege -c 1 ${clb_dnsname}                                    # simulate 1 concurrent users
```

The HPA TARGET ratio will start to drop and, after a couple of minutes, the values shown under REPLICAS will drop and pods will be terminated.
Eventually the CA will follow but this will not happen for over 10 minutes.
Feel free to move on if you are not prepared to wait to see any of this.

[Return To Main Menu](/README.md)
