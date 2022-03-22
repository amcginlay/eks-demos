# Helm - because packages need managing

Before you continue, please ensure you have completed the following sections
- **"K8s ClusterIP Services"**.
- **"Prepare Upgraded Image"**

Linux has [yum and apt](https://www.baeldung.com/linux/yum-and-apt).
Mac has [Homebrew](https://brew.sh/).
Windows has [Chocolatey](https://chocolatey.org/).
These are all package/dependency managers which help users consistently consume software in the manner inteded by their authors.

If the arrival of Kubernetes means [the operating system no longer matters](https://www.infoworld.com/article/3322120/sorry-linux-kubernetes-is-now-the-os-that-matters.html), then it too needs a package manager.
This is what [Helm](https://helm.sh/) is all about and, like Kubernetes, it is a graduated project of the [Cloud Native Computing Foundation (CNCF)](https://www.cncf.io/).
We can use Helm to both consume and publish software in a simple, repeatable and configurable way.

We start by showing how to consume popular open-source software published to a Helm repo.
A Helm repo consists of a collection of packages/releases which can be readily deployed to Kubernetes.
Add (i.e. import) a popular repo which provides us a menu of packages/releases from which you can choose.
```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm search repo
```

The source for these packages/releases are known as charts and these particular ones can be inspected at [https://github.com/bitnami/charts/tree/master/bitnami](https://github.com/bitnami/charts/tree/master/bitnami).
The `bitnami/apache` release is present in this list so you can install it into its own namespace as follows.
```bash
kubectl create namespace apache
helm -n apache upgrade -i apache bitnami/apache # "upgrade -i" is interpreted as install or upgrade, as necessary
kubectl -n apache get all                       # view the objects created
```

The `bitnami/apache` release is a simple package/release comprising a `deployment` with an associated `service` of type LoadBalancer but it could be much larger.
It could include `configmaps`, `serviceaccounts` or manifests for any other Kubernetes object your cluster is capable of consuming.

In this demo, you are more interested in packaging your own applications so uninstall Apache and unwind what you have done.
```bash
helm -n apache uninstall apache
kubectl delete namespace apache # be patient, this command may take few moments
```

Helm tempts us to get us started with the `helm create` command which builds the skeleton of a new chart.
Whilst it is undeniably useful to observe the structure it builds it is a little more comprehensive than you need right now.
Instead, you are going to build only the files you need and this will be done by hand.
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
The `Chart.yaml` file is mandatory for each Chart and acts like a header sheet for your package/release.
```bash
cat > ~/environment/echo-frontend/Chart.yaml << EOF
apiVersion: v2
name: echo-frontend
version: 1.0.0
EOF
```

Throughout the previous sections, whilst deploying your app, you have been carefully preserving its manifests in a directory named `templates`.
With the `Chart.yaml` file in place it now should be clear that the intention in these demos was always to deploy apps using Helm.

Helm provides a dry run option which allows us to "kick the tyres" and look for any potential errors.
```bash
helm -n demos upgrade -i --dry-run echo-frontend-blue ~/environment/echo-frontend/ \
  --create-namespace
```

This dry run **fails** as the `{{ .Values }}` directives inside your manifests, specifically those without default settings, are not being translated as they were previously via `sed`.
The simplest way to assist `helm` in resolving these placeholders is to pass in the required values on the command line as follows.
```bash
helm -n demos upgrade -i --dry-run echo-frontend-blue ~/environment/echo-frontend/ \
  --create-namespace \
  --set registry=${EKS_ECR_REGISTRY} \
  --set color=blue \
  --set version=1.0 \
  --set serviceType=ClusterIP
```

The dry run **fails** again, this time because `echo-frontend-blue` already exists.
Helm refuses to deploy over the top of an existing deployment which it does not currently own.
Empty the `demos` namespace then try the dry run again.
This time, note that we ask helm to re-create the namespace for us.
```bash
kubectl delete namespace demos # be patient, this command may take few moments
helm -n demos upgrade -i --dry-run echo-frontend-blue ~/environment/echo-frontend/ \
  --create-namespace \
  --set registry=${EKS_ECR_REGISTRY} \
  --set color=blue \
  --set version=1.0 \
  --set serviceType=ClusterIP
```

This time the dry run will produce no errors and output the translated manifests, just as the `tee` command did for you previously.
Take a moment to observe the output before **removing** the `--dry-run` switch and re-installing the app.
```bash
helm -n demos upgrade -i echo-frontend-blue ~/environment/echo-frontend/ \
  --create-namespace \
  --set registry=${EKS_ECR_REGISTRY} \
  --set color=blue \
  --set version=1.0 \
  --set serviceType=ClusterIP
```

In a **dedicated** terminal window, remote into your "jumpbox" and begin sending requests to the service.
```bash
kubectl exec -it jumpbox -- /bin/bash -c "while true; do curl http://echo-frontend-blue.demos.svc.cluster.local:80; sleep 0.25; done"
# ctrl+c to quit loop
```

Leave the looped command running for now.

`helm` now makes it easy now to upgrade the app to the version 2.0 image you created as follows.
```bash
helm -n demos upgrade -i echo-frontend-blue ~/environment/echo-frontend/ \
  --create-namespace \
  --set registry=${EKS_ECR_REGISTRY} \
  --set color=blue \
  --set version=2.0 \
  --set serviceType=ClusterIP
```

Hop over to the **dedicated** terminal window you left running and watch as the `curl` responses reveal the old pod replicas being rapidly superseded with new ones (check the `version` property).
This should only take few seconds and reveals something extremely valuable about running cloud native workloads on container orchestration platforms like Kubernetes.
Application updates can be applied in-place, quickly and usually with **zero downtime**.

Now imagine that you do not like the result of your rollout.
Helm has your back.
One simple command can roll back any deployment that fails to meet your expectations.
Keep an eye on the looped `curl` request as the following command is executed.
```bash
helm -n demos rollback echo-frontend-blue
```

If, at any point, you want Helm to reveal the path taken to get to where you are, here are a few more commands to look at.
```bash
helm list --all-namespaces      # Helm operations are namespaced by default
helm -n demos status echo-frontend-blue
helm -n demos history echo-frontend-blue
```

As we move on from this chapter we do, in fact, wish to be running the latest version (v2.0) of the frontend app.
We can do this by "rolling back the rollback", so once more ...
```bash
helm -n demos rollback echo-frontend-blue
```

... before finally checking that the looped `curl` request is reporing `"version":"2.0"`.

Finally, if you want to publish your own repo, take a look at [this](https://medium.com/containerum/how-to-make-and-share-your-own-helm-package-50ae40f6c221) or [this](https://github.com/komljen/helm-charts) for more information on how to do so.

[Return To Main Menu](/README.md)
