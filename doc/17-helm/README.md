# Helm - because packages need managing

This next section assumes that you you have completed the earlier section on **LoadBalancer services** and have a load balancer available.

Linux has [yum and apt](https://www.baeldung.com/linux/yum-and-apt).
Mac has [Homebrew](https://brew.sh/).
Windows has [Chocolatey](https://chocolatey.org/).
These are all package/dependency managers which help users consistently consume software in the manner inteded by its authors.

If the arrival of Kubernetes means [the operating system no longer matters](https://www.infoworld.com/article/3322120/sorry-linux-kubernetes-is-now-the-os-that-matters.html), then it too needs a package manager.
This is what [Helm](https://helm.sh/) is all about and, like Kubernetes, it is a graduated project of the [Cloud Native Computing Foundation (CNCF)](https://www.cncf.io/).
We can use Helm to both consume and produce software in a simple, repeatable and configurable way.

We start by showing how to consume popular open-source software published to a Helm repo.
A Helm repo consists of a collection of packages/releases which can be readily deployed to Kubernetes.
We start by adding (i.e. importing) a popular repo and displaying a menu from which we can now choose.
```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm search repo
```

The source for these packages/releases are known as charts and these particular ones can be inspected at [https://github.com/bitnami/charts/tree/master/bitnami](https://github.com/bitnami/charts/tree/master/bitnami).
The `bitnami/apache` release is present in this list so we can install it into its own namespace as follows.
```bash
kubectl create namespace apache
helm -n apache upgrade -i apache bitnami/apache # "upgrade -i" is interpreted as install or upgrade, as necessary
kubectl -n apache get all                       # view the objects created
```

The `bitnami/apache` release is a simple package/release comprising a `deployment` with an associated `service` of type LoadBalancer but it could be much larger.
It could include `configmaps`, `serviceaccounts` or any other YAML-defined object your cluster is capable of consuming.

If we misconfigure something down the line we can re-install Apache by applying the same command.
```bash
helm -n apache upgrade -i apache bitnami/apache
```

In our case, the previous command changed nothing but with Helm now taking responsibility for deploying our applications we can ask it what it has done.
```bash
helm list --all-namespaces                      # Helm operations are namespaced by default
helm -n apache history apache
```

But we are more interested in packaging our own applications so uninstall Apache and unwind what we have done.
```bash
helm -n apache uninstall apache
kubectl delete namespace apache
```

Helm tempts us to get us started with the `helm create` command which builds the skeleton of a new chart.
Whilst it is undeniably useful to observe the structure it builds, it is a little more comprehensive than we need right now.
Instead, we are going to build only the files we need and we will do that by hand.
So have a quick peek then discard it.
```bash
helm create dummy-app
tree -a dummy-app
rm -rf ./dummy-app
```

The package/release we want Helm to capture is the application we already have in our `EKS_APP_NS` namespace.
If we can achieve this, then our friends and customers can deploy our software on their own clusters in the exact manner we inteded.
We can begin by creating a local home for our new chart.
```bash
mkdir -p ~/environment/helm-charts/${EKS_APP}/templates
```

Start by building a minimal `Chart.yaml` file.
This is like a header sheet for our package/release and is mandatory for each Chart.
```bash
cat > ~/environment/helm-charts/${EKS_APP}/Chart.yaml << EOF 
apiVersion: v2
name: ${EKS_APP}
version: 1.0.0
EOF
```

Extract the manifests for the `deployment` and `service` objects of our application.
We use `kubectl neat` to keep our captured YAML manifests lean and we strip out the namespace for the purpose of generalization.
The resulta are acceptable but they can certainly be further refined.
For example it is not ideal to be hardcoding `clusterIP` addresses or `nodePort` values as these may already be in use on the target cluster which will result in a failed installation.
```bash
kubectl -n ${EKS_APP_NS} get deployment ${EKS_APP} -o yaml | \
  kubectl neat | \
  sed "/namespace: ${EKS_APP_NS}/d" \
  > ~/environment/helm-charts/${EKS_APP}/templates/deployment.yaml

kubectl -n ${EKS_APP_NS} get service ${EKS_APP} -o yaml | \
  kubectl neat | \
  sed "/namespace: ${EKS_APP_NS}/d" \
  > ~/environment/helm-charts/${EKS_APP}/templates/service.yaml
```

Helm offers a dry run option which allows us to "kick the tyres" and look for any potential errors.
```bash
helm -n ${EKS_APP_NS} upgrade -i --dry-run ${EKS_APP} ~/environment/helm-charts/${EKS_APP}
```

This dry run fails because we would be asking Helm to deploy over the top of an existing deployment which is not under its control.
Throwing caution to the wind, just delete the `EKS_APP_NS` namespace then recreate it and try the dry run again.
```bash
kubectl delete namespace ${EKS_APP_NS}
kubectl create namespace ${EKS_APP_NS}
helm -n ${EKS_APP_NS} upgrade -i --dry-run ${EKS_APP} ~/environment/helm-charts/${EKS_APP}
```

This time the dry run will produce no errors and we can just go for it.
```bash
helm -n ${EKS_APP_NS} upgrade -i ${EKS_APP} ~/environment/helm-charts/${EKS_APP}
```

[Return To Main Menu](/README.md)
