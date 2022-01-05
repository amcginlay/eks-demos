# Helm - because packages need managing

If you have **not** completed the earlier section on Services (Load Distribution) then you may not have an appropriate service manifest and corresponding service object in place.
Your nginx "jumpbox" may also be missing.
If so, please return and complete the section named **"K8s ClusterIP Services"**.

Linux has [yum and apt](https://www.baeldung.com/linux/yum-and-apt).
Mac has [Homebrew](https://brew.sh/).
Windows has [Chocolatey](https://chocolatey.org/).
These are all package/dependency managers which help users consistently consume software in the manner inteded by their authors.

If the arrival of Kubernetes means [the operating system no longer matters](https://www.infoworld.com/article/3322120/sorry-linux-kubernetes-is-now-the-os-that-matters.html), then it too needs a package manager.
This is what [Helm](https://helm.sh/) is all about and, like Kubernetes, it is a graduated project of the [Cloud Native Computing Foundation (CNCF)](https://www.cncf.io/).
We can use Helm to both consume and publish software in a simple, repeatable and configurable way.

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
It could include `configmaps`, `serviceaccounts` or manifests for any other Kubernetes object your cluster is capable of consuming.

But we are more interested in packaging our own applications so uninstall Apache and unwind what we have done.
```bash
helm -n apache uninstall apache
kubectl delete namespace apache # be patient, this command may take few moments
```

Helm tempts us to get us started with the `helm create` command which builds the skeleton of a new chart.
Whilst it is undeniably useful to observe the structure it builds it is a little more comprehensive than we need right now.
Instead, we are going to build only the files we need and we will do that by hand.
So have a quick peek at what `helm create` produces then discard its results.
Observe the presence of a `Chart.yaml` file and a `templates` directory as these represent the basic requirements of a Helm chart.
```bash
helm create dummy-app
tree -a dummy-app
rm -rf ./dummy-app
```

The package/release you want to build is the `echo-frontend` app you have already deployed.
By achieving this, your friends and customers can deploy your software on their own clusters in the exact manner you intended.
You do this by creating a Helm chart from the manifests which comprise your app.
The `Chart.yaml` file is mandatory for each Chart and acts like a header sheet for our package/release.
```bash
cat > ~/environment/echo-frontend/Chart.yaml << EOF
apiVersion: v2
name: echo-frontend
version: 1.0.0
EOF
```

Throughout the previous sections, whilst deploying your app, you have been carefully preserving its manifests in a directory named `templates`.
With the `Chart.yaml` file in place it now should be clear that our intention was always to deploy apps using Helm.

Helm provides a dry run option which allows us to "kick the tyres" and look for any potential errors.
```bash
helm -n demos upgrade -i --dry-run echo-frontend-blue ~/environment/echo-frontend/
```

This dry run **fails** as the `{{ .Values }}` directives inside our manifests, specifically those without default settings, are not being translated as they were previously via `sed`.
The simplest way to assist `helm` in resolving these placeholders is to pass in the required values on the command line as follows.
```bash
helm -n demos upgrade -i --dry-run echo-frontend-blue ~/environment/echo-frontend/ \
  --set registry=${EKS_ECR_REGISTRY} \
  --set color=blue \
  --set version=1.0 \
  --set serviceType=LoadBalancer
```

The dry run **fails** again, this time because `echo-frontend-blue` already exists.
Helm refuses to deploy over the top of an existing deployment which it does not currently own.
Throwing caution to the wind, just empty the `demos` namespace then try the dry run again.
```bash
kubectl delete namespace demos # be patient, this command may take few moments
kubectl create namespace demos
helm -n demos upgrade -i --dry-run echo-frontend-blue ~/environment/echo-frontend/ \
  --set registry=${EKS_ECR_REGISTRY} \
  --set color=blue \
  --set version=1.0 \
  --set serviceType=LoadBalancer
```

This time the dry run will produce no errors and output the translated manifests.
Take a moment to observe the output before **removing** the `--dry-run` switch and re-installing the app.
```bash
helm -n demos upgrade -i echo-frontend-blue ~/environment/echo-frontend/ \
  --set registry=${EKS_ECR_REGISTRY} \
  --set color=blue \
  --set version=1.0 \
  --set serviceType=LoadBalancer
```

In a **dedicated** terminal window, remote into your "jumpbox" and begin sending requests to the service.
```bash
kubectl exec -it jumpbox -- /bin/bash -c "while true; do curl http://echo-frontend-blue.demos.svc.cluster.local:80; sleep 0.25; done"
# ctrl+c to quit loop
```

Leave this running for now.

`helm` now makes it easy now to upgrade the app to the version 2.0 image we created as follows.
```bash
helm -n demos upgrade -i echo-frontend-blue ~/environment/echo-frontend/ \
  --set registry=${EKS_ECR_REGISTRY} \
  --set color=blue \
  --set version=2.0 \
  --set serviceType=LoadBalancer
```

Hop over to the **dedicated** terminal window you left running and watch as the `curl` responses reveal the old pod replicas being rapidly superceded with new ones (check the `version` property).
This should only take few seconds and reveals something extremely valuable about running cloud native workloads on container orchestration platforms like Kubernetes.
Application updates can be applied in-place, quickly and usually with **zero downtime**.

Now imagine that you do not like the result of your rollout.
Helm has your back.
One simple command can roll back any deployment that fails to meet your expectations.
Keep an eye on the looped `curl` request as the following command is executed.
```bash
helm -n demos rollback echo-frontend-blue
```

If, at any point, we want Helm to reveal the path we took to get where we are, here are a few more commands to look at.
```bash
helm list --all-namespaces      # Helm operations are namespaced by default
helm -n demos status echo-frontend-blue
helm -n demos history echo-frontend-blue
```

Finally, if you want to publish your own repo, take a look at [this](https://medium.com/containerum/how-to-make-and-share-your-own-helm-package-50ae40f6c221) or [this](https://github.com/komljen/helm-charts) for more information on how to do so.

[Return To Main Menu](/README.md)
