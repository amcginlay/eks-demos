# Deploy From ECR To Kubernetes

### `kubectl` manifest generation (a side note)

Kubernetes objects, such as deployments and services, require YAML manifests in order to be created.
Familiarizing oneself with the structure of Kubernetes manifests can be a significant barrier for beginners.
`kubectl` supports a number of convenient factory-style commands, known as [generators](https://kubernetes.io/docs/reference/kubectl/conventions/#generators), which can build Kubernetes objects without the need to code their underlying manifests by hand.

Generators also supports a dry run facility, enabling them to serve as simple manifest builders.
This behaviour can be observed when executing the following non-destructive command.
NOTE the command also makes use of the `neat` plug-in for `kubectl` (installed via [krew](https://github.com/kubernetes-sigs/krew)) which helps reduce manifests to their essential elements.
```bash
kubectl create deployment dummy-deployment --image dummy --dry-run=client -o yaml | kubectl neat
```

### Deploy our application at scale

Create a directory in which to store your Kubernetes manifests.
```bash
mkdir -p ~/environment/echo-frontend/templates/
```

Create a namespace named `demos` which will host our objects.
```bash
cat << EOF | tee ~/environment/echo-frontend/templates/demos-namespace.yaml | \
             kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: demos
EOF
```

Create a manifest and deployment for the first incarantion of your app.
Version **1.0** of your app is packaged into a deployment suffixed with the color **blue**.
Observe how your manifest employs a `{{ .Values }}` templating syntax for these settings which `sed` resolves immediately prior to passing to `kubectl apply`.
The reason behind this choice of syntax will become evident as we progress through the demos.
```bash
cat << EOF | tee ~/environment/echo-frontend/templates/echo-frontend-deployment.yaml | \
             sed "s/{{ .Values.registry }}/${EKS_ECR_REGISTRY}/g" | \
             sed "s/{{ .Values.color }}/blue/g" | \
             sed "s/{{ .Values.version }}/1.0/g" | \
             kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: echo-frontend-{{ .Values.color }}
  namespace: demos
  labels:
    app: echo-frontend-{{ .Values.color }}
spec:
  replicas: 3
  revisionHistoryLimit: 0
  selector:
    matchLabels:
      app: echo-frontend-{{ .Values.color }}
  template:
    metadata:
      labels:
        app: echo-frontend-{{ .Values.color }}
    spec:
      containers:
      - name: echo-frontend
        image: {{ .Values.registry }}/echo-frontend:{{ .Values.version }}
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
first_pod=$(kubectl -n demos get pods -l app=echo-frontend-blue -o name | head -1)
kubectl -n demos exec -it ${first_pod} -- curl http://localhost:80
```

Do not delete this deployment. We will need it later.

[Return To Main Menu](/README.md)
