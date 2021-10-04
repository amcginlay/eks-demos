# AWS Load Balancer Controller - because one load balancer per service is wasteful

The previous section introduced the Kubernetes LoadBalancer service.
The EKS implementation of this creates one [Classic Load Balancer](https://aws.amazon.com/elasticloadbalancing/classic-load-balancer/) per service.
Whilst this provides a working solution it is not best suited for modern deployments built on upon VPC infrastructure and is not a configurable as we would like.
For this reason we recommend using the [AWS Load Balancer Controller](https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html).
This controller supports the use of Application and Network Load Balancers which are the preferred modern solutions.
It does not come installed as standard on EKS clusters so we need to follow the documented installation instructions which are presented in short form below.
We have not yet encountered the `helm` CLI tool but will do so in a later section.

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

[Return To Main Menu](/README.md)
