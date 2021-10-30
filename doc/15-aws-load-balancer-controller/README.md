# AWS Load Balancer Controller - because one load balancer per service is wasteful

The previous section introduced the Kubernetes LoadBalancer service.
The EKS implementation of this creates one [Classic Load Balancer](https://aws.amazon.com/elasticloadbalancing/classic-load-balancer/) per service.
Whilst this provides a working solution it is not best suited for modern deployments built upon VPC infrastructure and is not as configurable as we would like.
For example, it would be preferable to support multiple deployments from a single load balancer.
For this reason we recommend using the [AWS Load Balancer Controller](https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html).
This controller supports the use of [Application Load Balancers](https://aws.amazon.com/elasticloadbalancing/application-load-balancer/) and [Network Load Balancers](https://aws.amazon.com/elasticloadbalancing/network-load-balancer/) which are the preferred modern solutions.

The AWS Load Balancer Controller does not come installed as standard on EKS clusters so we need to follow the documented installation instructions which are presented in short form below.
These instructions install the deployment using `helm` - a package manager for Kubernetes which we have not yet encountered but will do so in a later section.
```bash
aws iam create-policy \
  --policy-name AWSLoadBalancerControllerIAMPolicy \
  --policy-document \
  file://<(curl --silent iam_policy.json https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.2.0/docs/install/iam_policy.json)

eksctl utils associate-iam-oidc-provider \
  --cluster ${EKS_CLUSTER_NAME} \
  --approve
  
eksctl create iamserviceaccount \
  --namespace=kube-system \
  --cluster=${EKS_CLUSTER_NAME} \
  --name=aws-load-balancer-controller \
  --attach-policy-arn=arn:aws:iam::${AWS_ACCOUNT_ID}:policy/AWSLoadBalancerControllerIAMPolicy \
  --override-existing-serviceaccounts \
  --approve

helm repo add eks https://aws.github.io/eks-charts

helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  --namespace kube-system \
  --set clusterName=${EKS_CLUSTER_NAME} \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller
```

Verify that the controller is installed.
```bash
kubectl -n kube-system get deployment aws-load-balancer-controller
```

Start by re-implementing what we had in the previous section - a single load balancer forwarding all traffic to one deployment via its service.
This time will be creating an Application Load Balancer (ALB).
```bash
kubectl -n ${EKS_APP_NS} create ingress ${EKS_APP} \
  --annotation kubernetes.io/ingress.class=alb \
  --annotation alb.ingress.kubernetes.io/scheme=internet-facing \
  --annotation alb.ingress.kubernetes.io/group.name=shared \
  --annotation alb.ingress.kubernetes.io/group.order=200 \
  --rule="/*=${EKS_APP}:80"
```

Grab the ALB DNS name and put the following `curl` command in a loop until the AWS resource is resolved (2-3 mins).
If you receive any errors, just wait a little longer.
```bash
alb_dnsname=$(kubectl -n ${EKS_APP_NS} get ingress ${EKS_APP} -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${alb_dnsname}; sleep 0.25; done
# ctrl+c to quit loop
```

As noted in the previous section, Kubernetes services of type **LoadBalancer** are derived from services of type **NodePort**.
The AWS Load Balancer Controller depends upon NodePort services in its routing rules.
Hence, either service type can be referenced as targets within these rules.
Use of NodePort services would however, in this context, require fewer AWS resources and be cost optimal.

If we're going to test multiple routes we need an alternate deployment.
Deploy the GCP echo server into an alternate namespace.
This deployment has an accompanying NodePort service which will become a new target for the ALB.
```bash
EKS_APP_ALT=alt-echo
EKS_APP_ALT_NS=${EKS_APP_ALT}
kubectl create namespace ${EKS_APP_ALT_NS}
kubectl -n ${EKS_APP_ALT_NS} create deployment ${EKS_APP_ALT} --replicas 0 --image gcr.io/google_containers/echoserver:1.10 # begin with zero replicas
kubectl -n ${EKS_APP_ALT_NS} set resources deployment ${EKS_APP_ALT} --requests=cpu=200m,memory=200Mi                       # right-size the pods
kubectl -n ${EKS_APP_ALT_NS} patch deployment ${EKS_APP_ALT} --patch="{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"${EKS_APP_ALT}\",\"imagePullPolicy\":\"Always\"}]}}}}"
kubectl -n ${EKS_APP_ALT_NS} scale deployment ${EKS_APP_ALT} --replicas 1
kubectl -n ${EKS_APP_ALT_NS} expose deployment ${EKS_APP_ALT} --port=80 --target-port=8080 --type=NodePort                  # echoserver uses port 8080 internally
sleep 10 && kubectl -n ${EKS_APP_ALT_NS} get deployments,pods,services -o wide
```

Test this new service for internal reachability, checking for the incremented `version` attribute to confirm we have the correct container image.
```bash
worker_nodes=($(kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type=="InternalIP")].address}'))
node_port=$(kubectl -n ${EKS_APP_ALT_NS} get service -l app=${EKS_APP_ALT} -o jsonpath='{.items[0].spec.ports[0].nodePort}')
kubectl exec -it jumpbox -- /bin/bash -c "curl ${worker_nodes[0]}:${node_port}"
```

Now extend the ALB definition by creating a second ingress resource alongside our new deployment.
The `group-name` matches our first ingress, so it will be associated with the same ALB as before, but the `group-order` is lower so this path will be evaluated for a pattern match first.
```bash
kubectl -n ${EKS_APP_ALT_NS} create ingress ${EKS_APP_ALT} \
  --annotation kubernetes.io/ingress.class=alb \
  --annotation alb.ingress.kubernetes.io/scheme=internet-facing \
  --annotation alb.ingress.kubernetes.io/group.name=shared \
  --annotation alb.ingress.kubernetes.io/group.order=100 \
  --rule="/${EKS_APP_ALT}/*=${EKS_APP_ALT}:80"
```

Send separate curl requests to observe how a single ALB can forward traffic to multiple deployments in different namespaces.
```bash
curl http://${alb_dnsname}
curl http://${alb_dnsname}/${EKS_APP_ALT}
```

[Return To Main Menu](/README.md)
