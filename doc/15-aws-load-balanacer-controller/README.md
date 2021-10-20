# AWS Load Balancer Controller - because one load balancer per service is wasteful

The previous section introduced the Kubernetes LoadBalancer service.
The EKS implementation of this creates one [Classic Load Balancer](https://aws.amazon.com/elasticloadbalancing/classic-load-balancer/) per service.
Whilst this provides a working solution it is not best suited for modern deployments built on upon VPC infrastructure and is not as configurable as we would like.
For exmaaple, it would be preferable to support multiple deployments from a single load balancer.
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
kubectl -n ${EKS_NS_BLUE} create ingress ${EKS_APP} \
  --annotation kubernetes.io/ingress.class=alb \
  --annotation alb.ingress.kubernetes.io/scheme=internet-facing \
  --annotation alb.ingress.kubernetes.io/group.name=${EKS_APP} \
  --annotation alb.ingress.kubernetes.io/group.order=200 \
  --rule="/*=${EKS_APP}:80"
```

Grab the load balancer DNS name and put the following `curl` command in a loop until the AWS resource is resolved (2-3 mins).
If you receive any errors, just wait a little longer.
```bash
alb_dnsname=$(kubectl -n ${EKS_NS_BLUE} get ingress ${EKS_APP} -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${alb_dnsname}; sleep 0.25; done
# ctrl+c to quit loop
```

As noted in the previous section, Kubernetes services of type **LoadBalancer** are derived from services of type **NodePort**.
The AWS Load Balancer Controller depends upon NodePort services in its routing rules.
Hence, either service type can be referenced as targets within these rules.
Use of NodePort services would however, in this context, require fewer AWS resources and be cost optimal.

If we're going to test multiple routes we need an alternate deployment.
Deploy the **next** iteration of our app into an alternate namespace.
This deployment has an accompanying NodePort service which will become a new target for the ALB.
```bash
kubectl create namespace ${EKS_NS_GREEN}
kubectl -n ${EKS_NS_GREEN} create deployment ${EKS_APP} --replicas 0 --image ${EKS_APP_ECR_REPO}:${EKS_APP_VERSION_NEXT} # begin with zero replicas
kubectl -n ${EKS_NS_GREEN} set resources deployment ${EKS_APP} --requests=cpu=200m,memory=200Mi                          # right-size the pods
kubectl -n ${EKS_NS_GREEN} patch deployment ${EKS_APP} --patch="{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"${EKS_APP}\",\"imagePullPolicy\":\"Always\"}]}}}}"
kubectl -n ${EKS_NS_GREEN} scale deployment ${EKS_APP} --replicas 3
kubectl -n ${EKS_NS_GREEN} expose deployment ${EKS_APP} --port=80 --type=NodePort
sleep 10 && kubectl -n ${EKS_NS_GREEN} get deployments,pods,services -o wide
```

Test this new service for internal reachability, checking for the incremented `version` attribute to confirm we have the correct container image.
```bash
worker_nodes=($(kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type=="InternalIP")].address}'))
node_port=$(kubectl -n ${EKS_NS_GREEN} get service -l app=${EKS_APP} -o jsonpath='{.items[0].spec.ports[0].nodePort}')
kubectl exec -it jumpbox -- /bin/bash -c "curl ${worker_nodes[0]}:${node_port}"
```

Now extend the load balancer definition by creating a second ingress resource alongside our new deployment.
The `group-name` matches our first ingress, so it will be associated with the same load balancer as before, but the `group-order` is lower so this path will be evaluated for a pattern match first.
```bash
kubectl -n ${EKS_NS_GREEN} create ingress ${EKS_APP} \
  --annotation kubernetes.io/ingress.class=alb \
  --annotation alb.ingress.kubernetes.io/scheme=internet-facing \
  --annotation alb.ingress.kubernetes.io/group.name=${EKS_APP} \
  --annotation alb.ingress.kubernetes.io/group.order=100 \
  --rule="/alt-path/*=${EKS_APP}:80"
```

Send a curl request to `alt-path` to see how a single ALB can support traffic to multiple deployments.
```bash
curl http://${alb_dnsname}/alt-path
```

[Return To Main Menu](/README.md)
