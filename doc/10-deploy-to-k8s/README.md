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

### Deploy your application at scale

Create a directory in which to store your Kubernetes manifests.
```bash
mkdir -p ~/environment/echo-frontend/templates/
```

Create a namespace named `demos` which will host your objects.
```bash
kubectl create namespace demos
```

Before you ask your cluster to deploy the first incarnation of your deployment object, download its manifest to your Cloud9 environment.
```bash
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-frontend/templates/echo-frontend-deployment.yaml \
  -O ~/environment/echo-frontend/templates/echo-frontend-deployment.yaml
```

Open `~/environment/echo-frontend/templates/echo-frontend-deployment.yaml` in Cloud9 IDE to review the code.
Observe that this is a templated version of your manifest which employs a `{{ .Values }}` templating syntax.

You will preserve this baseline version on disk and, for now, use [`sed`](https://en.wikipedia.org/wiki/Sed) to dynamically resolve the `{{ .Values }}` settings.
[tee](https://en.wikipedia.org/wiki/Tee_(command)) dumps this to the terminal so you get a chance to observe the result which get forwarded to `kubectl apply`.
The reason behind this approach and choice of syntax will become evident as you progress through the demos.
Initially, you want `version` **1.0** of your app to be packaged into a deployment suffixed with the `color` **blue**.
```bash
cat ~/environment/echo-frontend/templates/echo-frontend-deployment.yaml | \
    sed "s/{{ .*.registry }}/${EKS_ECR_REGISTRY}/g" | \
    sed "s/{{ .*.color }}/blue/g" | \
    sed "s/{{ .*.replicas }}/3/g" | \
    sed "s/{{ .*.version }}/1.0/g" | \
    sed "s/{{ .*.backend }}/none/g" | \
    tee /dev/tty | \
    kubectl -n demos apply -f -
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

Do not delete this deployment. You will need it later.

[Return To Main Menu](/README.md)
