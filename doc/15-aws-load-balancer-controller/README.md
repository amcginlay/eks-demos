# AWS Load Balancer Controller - because one load balancer per service is wasteful

The previous section introduced the Kubernetes LoadBalancer service.
The EKS implementation of this creates one AWS Classic Load Balancer (CLB) per service.
Whilst this provides a working solution it is not best suited for modern deployments built upon VPC infrastructure and is not as configurable as you may expect.
It would be preferable to support multiple deployments from a single load balancer but this is a requirement which the CLB cannot satisfy.
For this reason it is recommended to use the [AWS Load Balancer Controller](https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html).
This controller supports the use of [AWS Application Load Balancers (ALB)](https://aws.amazon.com/elasticloadbalancing/application-load-balancer/) and [Network Load Balancers (NLB)](https://aws.amazon.com/elasticloadbalancing/network-load-balancer/) which are the preferred modern solutions.

The AWS Load Balancer Controller does not come installed as standard on EKS clusters so you need to follow the documented installation instructions which are presented in short form below.
These instructions install the deployment using `helm` - a package manager for Kubernetes that will be covered in a later section.
This section assumes that the OIDC provider of your cluster has been registered for use with IAM.
This is the case as you previously set `withOIDC: true` in the cluster config YAML file, but check out [this link](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html) if you need help re-applying the setting.

Install the AWS Load Balancer Controller as follows.
```bash
aws iam delete-policy --policy-arn arn:aws:iam::${AWS_ACCOUNT_ID}:policy/AWSLoadBalancerControllerIAMPolicy >/dev/null 2>&1
aws iam create-policy \
  --policy-name AWSLoadBalancerControllerIAMPolicy \
  --policy-document \
  file://<(curl --silent iam_policy.json https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.3.1/docs/install/iam_policy.json)

eksctl create iamserviceaccount \
  --cluster=${EKS_CLUSTER_NAME} \
  --namespace kube-system \
  --name=aws-load-balancer-controller \
  --attach-policy-arn=arn:aws:iam::${AWS_ACCOUNT_ID}:policy/AWSLoadBalancerControllerIAMPolicy \
  --override-existing-serviceaccounts \
  --approve

helm repo add eks https://aws.github.io/eks-charts

helm -n kube-system install aws-load-balancer-controller eks/aws-load-balancer-controller \
  --set clusterName=${EKS_CLUSTER_NAME} \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller
```

Verify that the AWS Load Balancer Controller is installed.
```bash
kubectl -n kube-system get deployment aws-load-balancer-controller
```

You are now about to configure the ALB to support multiple front ends simulataneously.
To support this you will now create a separate deployment and service.
You will use version **2.0** of your app, packaged into a deployment suffixed with the color **green** alongside an associated service of type **LoadBalancer**.
Recall that services of type **LoadBalancer** are derived from the **NodePort** service type, which is a minimum requirement for targets in ALB routing rules.
```bash
cat ~/environment/echo-frontend/templates/echo-frontend-deployment.yaml \
    <(echo ---) \
    ~/environment/echo-frontend/templates/echo-frontend-service.yaml | \
    sed "s/{{ .*.registry }}/${EKS_ECR_REGISTRY}/g" | \
    sed "s/{{ .*.color }}/green/g" | \
    sed "s/{{ .*.replicas }}/3/g" | \
    sed "s/{{ .*.version }}/2.0/g" | \
    sed "s/{{ .*.backend }}/none/g" | \
    sed "s/{{ .*.serviceType }}/NodePort/g" | \
    tee /dev/tty | \
    kubectl -n demos apply -f -
```

Check that both `blue` and `green` services exist and are of a compatible TYPE (either **LoadBalancer** or **NodePort**).
If necessary, revisit the appropriate sections to create/upgrade any services before moving on.
```
kubectl -n demos get services
```

Now you can download the manifest for an [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) object.
```bash
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-frontend/templates/echo-frontend-ingress.yaml \
  -O ~/environment/echo-frontend/templates/echo-frontend-ingress.yaml
```

Open `~/environment/echo-frontend/templates/echo-frontend-ingress.yaml` in Cloud9 IDE to review the code.
You will observe that your ALB is configured to route traffic to both services; version 1.0 via the **/blue/** path and version 2.0 via the **/green/** path.
You will also note that it contains no templated settings, therefore, no `sed` replacements are required this time. 

Deploy your ingress object to deploy and configure an ALB.
```bash
kubectl -n demos apply -f ~/environment/echo-frontend/templates/echo-frontend-ingress.yaml
```

Inspect your first ingress object and confirm that an ADDRESS (i.e. DNS name) is displayed.
```bash
sleep 20 && kubectl -n demos get ingress
```

Grab the DNS name for your ALB and put the following `curl` command in a loop until the AWS resource is resolved (2-3 mins).
If you receive any errors, just wait a little longer.
```bash
alb_dnsname=$(kubectl -n demos get ingress echo-frontend -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${alb_dnsname}/blue/; sleep 0.25; done
# ctrl+c to quit loop
```

Once you get a response, try sending requests on your **blue/green** paths to observe how a single ALB can simultaneously support traffic routed to multiple services/deployments.
```bash
curl http://${alb_dnsname}/blue  # version 1.0
curl http://${alb_dnsname}/green # version 2.0
```

In a production environment you would likely favour ALBs over CLBs but, for now, CLBs will suffice so it is recommended that you unwind the ALB and **green** resources as follows.
```bash
rm ~/environment/echo-frontend/templates/echo-frontend-ingress.yaml
kubectl -n demos delete ingress echo-frontend                       # this discards the ALB so be patient here
kubectl -n demos delete service echo-frontend-green
kubectl -n demos delete deployment echo-frontend-green
```

[Return To Main Menu](/README.md)
