# K8s ClusterIP Services - because pods need to talk to each other

This section assumes that the `php-echo` app is deployed and scaled to 3 instances.

To test ClusterIP services we first need to gain "private" access to our deployments, just as we might do with a regular EC2 jumpbox (or bastion host).
To keep matters simple we can deploy [nginx](https://www.nginx.com) as a standalone pod which suits this purpose.
```bash
kubectl run jumpbox --image=nginx
sleep 5 && kubectl exec -it jumpbox -- curl localhost:80
```

Remote into nginx to demonstrate pod-to-pod communication ... which fails, because no such service exists yet ...
```bash
kubectl exec -it jumpbox -- curl ${EKS_APP_NAME}:80 # <---- FAILURE!
```

Introduce the service.
```bash
kubectl get services
kubectl expose deployment ${EKS_APP_NAME} --port=80 --type=ClusterIP
kubectl get services
```

Now pods can reach each other via services.
```bash
kubectl exec -it jumpbox -- /bin/bash -c "while true; do curl ${EKS_APP_NAME}:80; done"
# ctrl+c to quit loop
```

[Return To Main Menu](/README.md)
