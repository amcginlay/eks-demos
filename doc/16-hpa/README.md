# Horizontal Pod Autoscaler - because demand for pods can grow

# TODO this all need tidying up ...

# see that metrics server is missing
```bash
kubectl top nodes
```

# install the metrics server (NOTE "latest" may not be appropriate)
```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

# see that metrics server is now ready
```bash
kubectl describe deployment metrics-server -n kube-system
kubectl top nodes                                              # this may take a minute to produce results
kubectl top pods --all-namespaces
```

# activate the Horizontal Pod Autoscaler (hpa) for the first time
```bash
kubectl -n ${EKS_NS_BLUE} autoscale deployment ${EKS_APP} --cpu-percent=50 --min=3 --max=25
```

# in a dedicated terminal window, keep watching the k8s objects
```bash
watch "kubectl get nodes; echo; kubectl -n ${EKS_NS_BLUE} get deployments,hpa,pods -o wide"
```

# in a dedicated terminal window, use siege to put the app under heavy load
```bash
clb_dnsname=$(kubectl -n ${EKS_NS_BLUE} get service -l app=${EKS_APP} -o jsonpath='{.items[0].status.loadBalancer.ingress[0].hostname}')
siege -c 200 ${clb_dnsname}                                    # simulate 200 concurrent users
```

# NOTE this will cause the HPA to autoscale the pods to its max but many will remain in a "Pending" state
# Leave SIEGE running ... the Cluster Autoscaler will react

[Return To Main Menu](/README.md)
