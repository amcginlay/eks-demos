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

Create an Application Load Balancer object to take the place of the LoadBalancer service.
Note this new resource depends directly upon the underlying NodePort service which is why we left it running.
```bash
kubectl -n ${EKS_NS_BLUE} create ingress ${EKS_APP_NAME} \
  --annotation kubernetes.io/ingress.class=alb \
  --annotation alb.ingress.kubernetes.io/scheme=internet-facing \
  --rule="/=${EKS_APP_NAME}:80" \
  --rule="alt-path/=${EKS_APP_NAME}:80"
```

External port 80 requests are now load balanced across the underlying NodePort service. Grab the load balancer DNS name and put the following curl command in a loop as the AWS resource will not be immediately resolved (2-3 mins). If you receive any curl errors, just wait a little longer.
```bash
alb_dnsname=$(kubectl -n ${EKS_NS_BLUE} get ingress ${EKS_APP_NAME} -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${lb_dnsname}; sleep 1; done
```

[Return To Main Menu](/README.md)
