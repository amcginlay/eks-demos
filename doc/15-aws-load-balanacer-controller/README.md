# AWS Load Balancer Controller - because one load balancer per service is wasteful

The previous section introduced the Kubernetes LoadBalancer service.
The EKS implementation of this creates one [Classic Load Balancer](https://aws.amazon.com/elasticloadbalancing/classic-load-balancer/) per service.
Whilst this provides a working solution it is not best suited for modern deployments built on upon VPC infrastructure and is not as configurable as we would like.
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

With the help of the Application Load Balancer, the AWS Load Balancer Controller is able to support path-based routing which means a single load balancer resource is able to route traffic to multiple deployments simultaneously.
Deploy the next iteration of our app alongside the current one so we can see this feature in practice.
```bash
kubectl -n ${EKS_APP_NS} create deployment ${EKS_APP_GREEN} --replicas 0 --image ${EKS_APP_ECR_REPO}:${EKS_APP_VERSION_NEXT} # begin with zero replicas
kubectl -n ${EKS_APP_NS} set resources deployment ${EKS_APP_GREEN} --requests=cpu=200m,memory=200Mi                          # right-size the pods
kubectl -n ${EKS_APP_NS} patch deployment ${EKS_APP_GREEN} --patch="{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"${EKS_APP}\",\"imagePullPolicy\":\"Always\"}]}}}}"
kubectl -n ${EKS_APP_NS} scale deployment ${EKS_APP_GREEN} --replicas 3
kubectl -n ${EKS_APP_NS} expose deployment ${EKS_APP_GREEN} --port=80 --type=NodePort
```

Test the updated deployment for reachability
```bash
worker_nodes=($(kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type=="InternalIP")].address}'))
node_port=$(kubectl -n ${EKS_APP_NS} get service -l app=${EKS_APP_GREEN} -o jsonpath='{.items[0].spec.ports[0].nodePort}')
kubectl exec -it jumpbox -- /bin/bash -c "curl ${worker_nodes[0]}:${node_port}"
```

As noted in the previous section, Kubernetes services of type **LoadBalancer** are derived from services of type **NodePort**.
The AWS Load Balancer Controller depends upon NodePort services to build its routing rules.
Hence, either service type can be referenced as targets within these rules.
Use of NodePort services will however, in this context, require fewer AWS resources and be cost optimal.

Create an Application Load Balancer object with separate paths to the two underlying NodePort services, identified as `EKS_APP_BLUE` and `EKS_APP_GREEN`.
```bash
kubectl -n ${EKS_APP_NS} create ingress ${EKS_APP_GREEN} \
  --annotation kubernetes.io/ingress.class=alb \
  --annotation alb.ingress.kubernetes.io/scheme=internet-facing \
  --annotation alb.ingress.kubernetes.io/group.name=${EKS_APP} \
  --annotation alb.ingress.kubernetes.io/group.order=1 \
  --rule="/alt-path=${EKS_APP_GREEN}:80"

kubectl -n ${EKS_APP_NS} create ingress ${EKS_APP_BLUE} \
  --annotation kubernetes.io/ingress.class=alb \
  --annotation alb.ingress.kubernetes.io/scheme=internet-facing \
  --annotation alb.ingress.kubernetes.io/group.name=${EKS_APP} \
  --annotation alb.ingress.kubernetes.io/group.order=2 \
  --rule="/*=${EKS_APP_BLUE}:80"
```

External port 80 requests are now load balanced across the two underlying deployments/services. Grab the load balancer DNS name (from either ingress object) and put the following `curl` command in a loop until the AWS resource is resolved (2-3 mins). If you receive any errors, just wait a little longer.
```bash
alb_dnsname=$(kubectl -n ${EKS_APP_NS} get ingress ${EKS_APP_BLUE} -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${alb_dnsname}; sleep 0.25; done
# ctrl+c to quit loop
```

One of the benefits of the Application Load Balancer is that it can support path-based routing within its rules to allow a single load balancer to support mappings between paths and target groups.
You may have noticed that we supplied two rules in our call to `create ingress` above.
Send a curl request to `alt-path` to see how, in this case, two different endpoint paths can resolve to the same NodePort service.
```bash
curl http://${alb_dnsname}/alt-path/
```

[Return To Main Menu](/README.md)
