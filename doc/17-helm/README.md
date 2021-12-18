# Helm - because packages need managing

If you have completed the earlier section on **LoadBalancer services** then you will already have a load balancer (CLB) in front of your `echo-frontend` app.
If you do not have this, execute the following (2-3 mins).
```bash
kubectl -n demos expose deployment echo-frontend --port=80 --type=LoadBalancer
```

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
kubectl delete namespace apache
```

Helm tempts us to get us started with the `helm create` command which builds the skeleton of a new chart.
Whilst it is undeniably useful to observe the structure it builds it is a little more comprehensive than we need right now.
Instead, we are going to build only the files we need and we will do that by hand.
So have a quick peek at `helm create` then discard its results.
```bash
helm create dummy-app
tree -a dummy-app
rm -rf ./dummy-app
```

The package/release you want Helm to capture is the `echo-frontend` app you have already deployed.
If you can achieve this, then your friends and customers can deploy your software on their own clusters in the exact manner you intended.
You do this by creating a chart for your deployment.
The `Chart.yaml` file is mandatory for each Chart and acts like a header sheet for our package/release.
```bash
mkdir -p ~/environment/helm-charts/echo-frontend/templates
cat > ~/environment/helm-charts/echo-frontend/Chart.yaml << EOF 
apiVersion: v2
name: echo-frontend
version: 1.0.0
EOF
```

Throughout the previous sections, whilst deploying your app, you have been carefully preserving the app manifests on the Cloud9 file system.
Now it will become clear why we did this.
Copy selected contents of the `echo-frontend-1.0/manifests/` directory into the `helm-charts/echo-frontend/templates/` directory.
```bash
cp ~/environment/echo-frontend-1.0/manifests/demos-namespace.yaml \
   ~/environment/echo-frontend-1.0/manifests/echo-frontend-deployment.yaml \
   ~/environment/echo-frontend-1.0/manifests/echo-frontend-service.yaml \
   ~/environment/helm-charts/echo-frontend/templates/
```

Helm offers a dry run option which allows us to "kick the tyres" and look for any potential errors.
```bash
helm upgrade -i --dry-run echo-frontend ~/environment/helm-charts/echo-frontend/
```

This dry run **fails** because we would be asking Helm to deploy over the top of an existing deployment which is not under its control.
Throwing caution to the wind, just delete the `echo-frontend` namespace which will obliterates your entire application.
We can then try the dry run again.
```bash
kubectl delete namespace demos # this command will take few moments as it needs to dispose of the CLB
helm upgrade -i --dry-run echo-frontend ~/environment/helm-charts/echo-frontend/
```

This time the dry run will produce no errors and we can just go for it.
```bash
helm upgrade -i echo-frontend ~/environment/helm-charts/echo-frontend/
```

Remember our application incorporates a CLB.
In a **dedicated** terminal window, grab the CLB DNS name and put the following `curl` command in a loop as the AWS resource will not be immediately resolved (2-3 mins).
Leave this looped request running and we will return to view it again later.
```bash
clb_dnsname=$(kubectl -n demos get service echo-frontend -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${clb_dnsname}; sleep 0.25; done
```

We have seen how Helm can help us deploy in a repeatable way but we also stated that it was configurable.
The manifests, which are stored in a directory named `templates`, include hard-coded references to things like container images, so deploying new versions would require updates to those source files.
The clue is in the word "template", suggesting the manifests could contain placeholders that are dynamically resolved at the point of use.
Helm supports templating with the use of `{{ define }}` directives inside our manifests.

Modify the `deployment` manifest to make it a version agnostic template and perform a dry run to test this.
```bash
sed -i "s/echo-frontend:1.0/echo-frontend:{{ .Values.version }}/g" ~/environment/helm-charts/echo-frontend/templates/echo-frontend-deployment.yaml
helm upgrade -i --dry-run echo-frontend ~/environment/helm-charts/echo-frontend/
```

As expected, this dry run **fails**.
Helm is unable to resolve any value for the new `{{ .Values.version }}` directive inside our `deployment` manifest.
The simplest way to resolve this is to the set the missing variable on the command line.
Perform another dry run to test this.
```bash
helm upgrade -i --dry-run echo-frontend ~/environment/helm-charts/echo-frontend/ --set version=2.0
```

The dry run output **works** and reveals the templated replacement which looks as intended.
Run this once more, this time **for real**. 
```bash
helm upgrade -i echo-frontend ~/environment/helm-charts/echo-frontend/ --set version=2.0
```

Now hop over to the **dedicated** terminal window you left running and watch as the `curl` responses reveal the old pod replicas being rapidly superceded with new ones (check the `version` property).
This should only take few seconds and reveals something extremely valuable about running cloud native workloads on container orchestration platforms like Kubernetes.
Application updates can be applied quickly and with **zero downtime**.

If you do not like the result of your rollout, Helm has your back.
One simple command can roll back any deployment that fails to meet your expectations.
Keep an eye on the looped `curl` request as the following command is executed.
```bash
helm rollback echo-frontend
```

If, at any point, we want Helm to reveal where we currently are and the path we took to get there, here are a few more commands to look at.
```bash
helm list --all-namespaces                  # Helm operations are namespaced by default
helm status echo-frontend
helm history echo-frontend
```

Finally, if you want to publish your own repo, take a look at [this](https://medium.com/containerum/how-to-make-and-share-your-own-helm-package-50ae40f6c221) or [this](https://github.com/komljen/helm-charts) for more information on how to do so.

[Return To Main Menu](/README.md)
