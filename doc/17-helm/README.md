# Helm - because packages need managing

This next section assumes that you you have completed the earlier section on **LoadBalancer services** and have a load balancer available.

Linux has [yum and apt](https://www.baeldung.com/linux/yum-and-apt).
Mac has [Homebrew](https://brew.sh/).
Windows has [Chocolatey](https://chocolatey.org/).
These are all package/dependency managers which help users consistently consume software in the manner inteded by their authors.

If the arrival of Kubernetes means [the operating system no longer matters](https://www.infoworld.com/article/3322120/sorry-linux-kubernetes-is-now-the-os-that-matters.html), then it too needs a package manager.
This is what [Helm](https://helm.sh/) is all about and, like Kubernetes, it is a graduated project of the [Cloud Native Computing Foundation (CNCF)](https://www.cncf.io/).
We can use Helm to both consume and produce software in a simple, repeatable and configurable way.

We start by showing how to consume popular open-source software published to a Helm repo.
A Helm repo consists of a collection of packages/releases which can be readily deployed to Kubernetes.
Add (i.e. import) a popular repo which provides us a menu of packages/releases from which we can choose.
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
We can prepare by creating a local home for our new chart.
```bash
mkdir -p ~/environment/helm-charts/${EKS_APP_FE}/templates
```

Start by building a minimal `Chart.yaml` file.
This is like a header sheet for our package/release and is mandatory for each Chart.
```bash
cat > ~/environment/helm-charts/${EKS_APP_FE}/Chart.yaml << EOF 
apiVersion: v2
name: ${EKS_APP_FE}
version: 1.0.0
EOF
```

Extract the manifests for the `deployment` and `service` objects of our application.
We use `kubectl neat` to keep our captured YAML manifests lean and we strip out the namespace for the purpose of generalization.
The results are acceptable but they can certainly be further refined.
For example it is not ideal to be hardcoding `clusterIP` addresses or `nodePort` values as these may already be in use on the target cluster which will result in a failed installation.
```bash
kubectl -n ${EKS_APP_NS} get deployment ${EKS_APP_FE} -o yaml | \
  kubectl neat | \
  sed "/namespace: ${EKS_APP_NS}/d" \
  > ~/environment/helm-charts/${EKS_APP_FE}/templates/deployment.yaml

kubectl -n ${EKS_APP_NS} get service ${EKS_APP_FE} -o yaml | \
  kubectl neat | \
  sed "/namespace: ${EKS_APP_NS}/d" \
  > ~/environment/helm-charts/${EKS_APP_FE}/templates/service.yaml
```

Helm offers a dry run option which allows us to "kick the tyres" and look for any potential errors.
```bash
helm -n ${EKS_APP_NS} upgrade -i --dry-run ${EKS_APP_FE} ~/environment/helm-charts/${EKS_APP_FE}
```

This dry run **fails** because we would be asking Helm to deploy over the top of an existing deployment which is not under its control.
Throwing caution to the wind, just delete the `EKS_APP_NS` namespace which, in turn, obliterates our entire application.
We can then recreate the empty namespace and try the dry run again.
```bash
kubectl delete namespace ${EKS_APP_NS} # this command will take few moments as it needs to dispose of the CLB
kubectl create namespace ${EKS_APP_NS}
helm -n ${EKS_APP_NS} upgrade -i --dry-run ${EKS_APP_FE} ~/environment/helm-charts/${EKS_APP_FE}
```

This time the dry run will produce no errors and we can just go for it.
```bash
helm -n ${EKS_APP_NS} upgrade -i ${EKS_APP_FE} ~/environment/helm-charts/${EKS_APP_FE}
```

Remember our application incorporates a CLB.
In a **dedicated** terminal window, grab the CLB DNS name and put the following `curl` command in a loop as the AWS resource will not be immediately resolved (2-3 mins).
Leave this looped request running and we will return to view it again later.
```bash
clb_dnsname=$(kubectl -n ${EKS_APP_NS} get service ${EKS_APP_FE} -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${clb_dnsname}; sleep 0.25; done
```

We have seen how Helm can help us deploy in a repeatable way but we also stated that it was configurable.
The manifests, which are stored in a directory named `templates`, include hard-coded references to things like container images, so deploying new versions would require updates to those source files.
The clue is in the word "template", suggesting the manifests could contain placeholders that are dynamically re-written at the point of use.
Helm supports templating with the use of `{{ define }}` directives inside our manifests.

Modify the `deployment` manifest to make it a version agnostic template and perform a dry run to test this.
```bash
sed -i "s/${EKS_APP_FE_VERSION}/{{ .Values.version }}/g" ~/environment/helm-charts/${EKS_APP_FE}/templates/deployment.yaml
helm -n ${EKS_APP_NS} upgrade -i --dry-run ${EKS_APP_FE} ~/environment/helm-charts/${EKS_APP_FE}
```

As expected, this dry run **fails**.
Helm is unable to resolve any value for the new `{{ .Values.version }}` directive inside our `deployment` manifest.
The simplest way to resolve this is to the set the missing variable on the command line.
Perform another dry run to test this.
```bash
helm -n ${EKS_APP_NS} upgrade -i --dry-run ${EKS_APP_FE} ~/environment/helm-charts/${EKS_APP_FE} --set version=${EKS_APP_FE_VERSION_NEXT}
```

The dry run output **works** and reveals the templated replacement which looks as intended.
Run this once more, this time **for real**. 
```bash
helm -n ${EKS_APP_NS} upgrade -i ${EKS_APP_FE} ~/environment/helm-charts/${EKS_APP_FE} --set version=${EKS_APP_FE_VERSION_NEXT}
```

Now hop over to the **dedicated** terminal window and watch as the `curl` responses reveal the old pod replicas being rapidly superceded with new ones (check the `version` property).
This should only take few seconds and reveals something extremely valuable about running cloud native workloads on container orchestration platforms like Kubernetes.
Application updates with **zero downtime**.

If you do not like the result of your rollout, Helm has your back.
One simple command can roll back any deployment that fails to meet your expectations.
Keep an eye on the looped `curl` request as the following command is executed.
```bash
helm -n ${EKS_APP_NS} rollback ${EKS_APP_FE}
```

If, at any point, we want Helm to reveal where we currently are and the path we took to get there, here are a few more commands to look at.
```bash
helm list --all-namespaces                  # Helm operations are namespaced by default
helm -n ${EKS_APP_NS} status ${EKS_APP_FE}
helm -n ${EKS_APP_NS} history ${EKS_APP_FE}
```

Finally, if you want to publish your own repo, take a look at [this](https://medium.com/containerum/how-to-make-and-share-your-own-helm-package-50ae40f6c221) or [this](https://github.com/komljen/helm-charts) for more information on how to do so.

[Return To Main Menu](/README.md)
