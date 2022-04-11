# X-Ray - because observability at scale is hard

This section assumes that you have completed the previous section named **"AWS App Mesh"**.
This section is **not** designed to be a deep dive on X-Ray as you can get that info from many alternate sources.
Instead your goal is to see how App Mesh makes it easy to enable distributed tracing in your newly "meshed" architecture.

## Enable X-Ray via App Mesh

Re-apply the Helm release for the **App Mesh Controller**, this time providing appropriate settings for the pair of `tracing` parameters.
```bash
helm -n kube-system upgrade -i appmesh-controller eks/appmesh-controller \
  --set region=${AWS_DEFAULT_REGION} \
  --set serviceAccount.create=false \
  --set serviceAccount.name=appmesh-controller \
  --set tracing.enabled=true \
  --set tracing.provider=x-ray
```

## Prepare your watchers

**If not already in place**, in a **dedicated** terminal window run a looped command against the **frontend**.
This generates the traffic necessary to cause a service trace to appear in the X-Ray console
```bash
kubectl exec -it jumpbox -- /bin/bash -c "while true; do curl http://echo-frontend-blue.demos.svc.cluster.local:80; sleep 0.25; done"
```

From **another dedicated** terminal window, observe what happens now to the pods.
```bash
# ctrl+c to quit
watch kubectl -n demos get pods
```

## Restart the deployments

You may recall from the previous section that, with the App Mesh Controller in place and correctly configured, a restart caused an `envoy` container to appear in each of the `demo` namespace pods.
With X-Ray now enabled, a restart will cause each pod in the `demos` namespace to also acquire an X-Ray container, which is the equivalent of the X-Ray daemon that you may have encountered previously on EC2.

```bash
kubectl -n demos rollout restart deployment \
  echo-backend-blue \
  echo-backend-green \
  echo-frontend-blue
```

## View the traces

Head over to [https://us-west-2.console.aws.amazon.com/xray/home](https://us-west-2.console.aws.amazon.com/xray/home) to see your service graph.
Note there was no need to augment your codebase to achieve this.

Next: [Main Menu](/README.md)
