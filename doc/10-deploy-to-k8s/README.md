# Deploy From ECR To Kubernetes

### `kubectl` manifest generation (a side note)

Kubernetes objects, such as deployments and services, require YAML manifests in order to be created.
Learning to familiarize oneself with the structure of Kubernetes manifests can be a significant barrier for new beginners.
`kubectl` supports a number of convenient `create` commands, known as [generators](https://kubernetes.io/docs/reference/kubectl/conventions/#generators), which can build Kubernetes objects without the need to code their underlying manifests by hand.

Generators also supports a dry run facility, enabling `kubectl create` commands to double-up as simple manifest builders.
This behaviour can be observed when executing the following non-destructive command.
NOTE the command also makes use of the `kubectl neat` add-on which reduces generated manifests down to their essential elements.
```bash
kubectl create deployment dummy-deployment --image dummy --dry-run=client -o yaml | kubectl neat
```

### Deploy our application at scale

Create a directory in which to store your Kubernetes manifests.
```bash
mkdir -p ~/environment/echo-frontend-1.0/manifests/
```

Create a namespace named `demos` which will host our objects.
```bash
cat << EOF | tee ~/environment/echo-frontend-1.0/manifests/demos-namespace.yaml | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: demos
EOF
```

Create a deployment for version 1 of your app.
```bash
cat << EOF | tee ~/environment/echo-frontend-1.0/manifests/echo-frontend-deployment.yaml | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: echo-frontend
  namespace: demos
  labels:
    app: echo-frontend
spec:
  replicas: 3
  revisionHistoryLimit: 0
  selector:
    matchLabels:
      app: echo-frontend
  template:
    metadata:
      labels:
        app: echo-frontend
    spec:
      containers:
      - name: echo-frontend
        image: ${EKS_ECR_REGISTRY}/echo-frontend:1.0
        imagePullPolicy: Always
        resources:
          requests:
            memory: 200Mi
            cpu: 200m
EOF
```

Inspect your first deployment.
```bash
sleep 10 && kubectl -n demos get deployments,pods -o wide
```

Exec into the first pod to perform curl test.
```bash
first_pod=$(kubectl -n demos get pods -l app=echo-frontend -o name | head -1)
kubectl -n demos exec -it ${first_pod} -- curl localhost:80
```

Do not delete this deployment. We will need it later.

[Return To Main Menu](/README.md)
