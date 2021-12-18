# AWS Load Balancer Controller - because one load balancer per service is wasteful

The previous section introduced the Kubernetes LoadBalancer service.
The EKS implementation of this creates one AWS Classic Load Balancer (CLB) per service.
Whilst this provides a working solution it is not best suited for modern deployments built upon VPC infrastructure and is not as configurable as you may expect.
It would be preferable to support multiple deployments from a single load balancer but this is a requirement which the CLB cannot satisfy.
For this reason we recommend using the [AWS Load Balancer Controller](https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html).
This controller supports the use of [AWS Application Load Balancers (ALB)](https://aws.amazon.com/elasticloadbalancing/application-load-balancer/) and [Network Load Balancers (NLB)](https://aws.amazon.com/elasticloadbalancing/network-load-balancer/) which are the preferred modern solutions.

The AWS Load Balancer Controller does not come installed as standard on EKS clusters so we need to follow the documented installation instructions which are presented in short form below.
These instructions install the deployment using `helm` - a package manager for Kubernetes that we will cover in a later section.
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
  --namespace kube-system \
  --cluster=${EKS_CLUSTER_NAME} \
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

Start by re-implementing what we had in the previous section - a single load balancer forwarding all traffic to one deployment via its service.
This time we will be creating an Application Load Balancer (ALB).
```bash
cat << EOF | tee ~/environment/echo-frontend-1.0/manifests/echo-frontend-ingress.yaml | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: echo-frontend
  namespace: demos
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/group.name: shared
    alb.ingress.kubernetes.io/group.order: "200"
spec:
  rules:
  - http:
      paths:
      - backend:
          service:
            name: echo-frontend
            port:
              number: 80
        path: /
        pathType: Prefix
EOF
```

Grab the DNS name for your ALB and put the following `curl` command in a loop until the AWS resource is resolved (2-3 mins).
If you receive any errors, just wait a little longer.
```bash
sleep 20 && alb_dnsname=$(kubectl -n demos get ingress echo-frontend -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${alb_dnsname}; sleep 0.25; done
# ctrl+c to quit loop
```

The AWS Load Balancer Controller depends upon NodePort services for its routing rules.
We currently have one compatible service but if we're going to test multiple routes through the ALB we will need an alternate deployment and service.
This deployment will have an accompanying NodePort service which will become a new target for the ALB.

Create a directory in which to store your manifests for your alternative app.
```bash
mkdir -p ~/environment/echo-alt/manifests/
```

Create a namespace named `demos-alt` which will host our objects.
```bash
cat << EOF | tee ~/environment/echo-alt/manifests/demos-alt-namespace.yaml | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: demos-alt
EOF
```

Create a deployment and NodePort service for your alternative app.
```bash
cat << EOF | tee ~/environment/echo-alt/manifests/echo-alt-deployment.yaml | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: echo-alt
  namespace: demos-alt
  labels:
    app: echo-alt
spec:
  replicas: 1
  revisionHistoryLimit: 0
  selector:
    matchLabels:
      app: echo-alt
  template:
    metadata:
      labels:
        app: echo-alt
    spec:
      containers:
      - name: echo-alt
        image: gcr.io/google_containers/echoserver:1.10
        imagePullPolicy: Always
        resources:
          requests:
            memory: 200Mi
            cpu: 200m
EOF
```

```bash
cat << EOF | tee ~/environment/echo-alt/manifests/echo-alt-service.yaml | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: echo-alt
  namespace: demos-alt
  labels:
    app: echo-alt
spec:
  type: NodePort
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: echo-alt
EOF
```

Take a look at what was produced
```bash
kubectl -n demos-alt get deployments,pods,services -o wide
```

Test this new service for internal reachability.
The output is notably different to our previously deployed application.
```bash
worker_nodes=($(kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type=="InternalIP")].address}'))
node_port=$(kubectl -n demos-alt get service -l app=echo-alt -o jsonpath='{.items[0].spec.ports[0].nodePort}')
kubectl exec -it jumpbox -- curl ${worker_nodes[0]}:${node_port}
```

Now extend the ALB definition by creating a second ingress resource alongside your new deployment.
The `group-name` matches our first ingress, so it will be associated with the same ALB as before, but the `group-order` is lower so this path will be evaluated for a pattern match first.
```bash
cat << EOF | tee ~/environment/echo-alt/manifests/echo-alt-ingress.yaml | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: echo-alt
  namespace: demos-alt
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/group.name: shared
    alb.ingress.kubernetes.io/group.order: "100"
spec:
  rules:
  - http:
      paths:
      - backend:
          service:
            name: echo-alt
            port:
              number: 80
        path: /echo-alt/
        pathType: Prefix
EOF
```

Send separate curl requests to observe how a single ALB can forward traffic to multiple deployments in different namespaces.
```bash
curl http://${alb_dnsname}          # our original app
curl http://${alb_dnsname}/echo-alt # our alternative app
```

We only require one load balancer but we currently have two.
In a production environment we would likely favour the ALB over the CLB but for demo purposes the CLB will suffice so we recommend that you unwind all the resources generated in this demo as follows.
```bash
kubectl -n demos-alt delete ingress echo-alt
kubectl -n demos delete ingress echo-frontend # this discards the ALB so be patient here
helm -n kube-system uninstall aws-load-balancer-controller
kubectl delete namespace demos-alt
```

[Return To Main Menu](/README.md)
