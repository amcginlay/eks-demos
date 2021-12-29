# Horizontal Pod Autoscaler - because demand for pods can grow

If you have **not** completed the earlier section on Services (Load Distribution) then you may not have an appropriate service manifest and corresponding service object in place.
If so, please return and complete the sections named **"K8s ClusterIP Services"** and **"K8s LoadBalancer Services"**.

When your workloads come under pressure their CPU consumption will rise.
Cloud native best practices suggest that the response to this situation is to increase the number of app replicas which spreads the load.
We have already seen how **manually** increasing the number of pod replicas will cause the Cluster Autoscaler (CA) to attempt to add new nodes to the cluster.
But what if your workloads come under CPU pressure and no one is present to make this adjustment?

The [Horizontal Pod Autoscaler (HPA)](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) aims to right-size the number of replicas in your deployments as a reaction to realtime changes in workload pressure.
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

One could generate an HPA using [kubectl autoscale](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/) command however, in the interests of maintaining focus on manifests, you will build one by hand.
```bash
cat << EOF | tee ~/environment/echo-frontend/templates/echo-frontend-hpa.yaml | \
             sed "s/{{ .Values.color }}/blue/g" | \
             sed "s/{{ .Values.minReplicas }}/3/g" | \
             sed "s/{{ .Values.maxReplicas }}/25/g" | \
             sed "s/{{ .Values.targetCPUUtilizationPercentage }}/50/g" | \
             kubectl apply -f -
apiVersion: autoscaling/v1
kind: HorizontalPodAutoscaler
metadata:
  name: echo-frontend-{{ .Values.color }}
spec:
  minReplicas: {{ .Values.minReplicas }}
  maxReplicas: {{ .Values.maxReplicas }}
  targetCPUUtilizationPercentage: {{ .Values.targetCPUUtilizationPercentage }}
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: echo-frontend-{{ .Values.color }}
EOF
```

Keep watching the k8s objects in a **dedicated** terminal window.
```bash
watch "kubectl get nodes; echo; kubectl -n demos get deployments,hpa,pods -o wide"
```

In another **dedicated** terminal window, use siege to put the app under heavy load.
```bash
clb_dnsname=$(kubectl -n demos get service -l app=echo-frontend-blue -o jsonpath='{.items[0].status.loadBalancer.ingress[0].hostname}')
siege -c 5 ${clb_dnsname} # simulate 5 concurrent users
```

This will cause the HPA to autoscale the pods upwards towards its `maxReplicas` setting.
Switch back to the terminal window displaying the `watch` results, paying attention to the HPA values shown under TARGETS and REPLICAS.
Also, observe the list of pods as it grows.
Under heavy load the TARGETS ratio will be high and the number of replicas will increase rapidly until the resources in the nodes are exhausted.
At this point you will observe that new pod IP addresses are unable to be allocated and a number of pods will remain in a Pending state.
As seen in an earlier section, the CA will step in to add new nodes and within a couple of minutes all the required pods will be Running as expected.

Switch back to the terminal window displaying the `siege` results.
Instead of stopping the load entirely, which is a little unnatural, we are just going to dial it back from high to low.
```bash
# ctrl+c to quit the running heavy-load siege command
siege -c 1 ${clb_dnsname} # simulate 1 concurrent users
```

The HPA TARGETS ratio will begin to drop.
After a couple of minutes, the values shown under REPLICAS will also drop.
Excess pods will be terminated and, correspondingly, the TARGETS ratio will stabilize.
The CA will eventually follow this pattern and underutilized nodes will be terminated.
However, as previously discussed, this will happen in a slower timeframe so feel free to quit `siege` and move on.

[Return To Main Menu](/README.md)
