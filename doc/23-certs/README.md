# Certificates Management - because you trust no one

This section assumes that you have completed the previous section named **"AWS App Mesh"**.
The assumptions listed in that section also apply here.

## Install system components

Install cert-manager
```bash
helm repo add jetstack https://charts.jetstack.io && helm repo update
# helm search repo jetstack/cert-manager --versions
helm -n cert-manager upgrade -i cert-manager jetstack/cert-manager \
  --create-namespace \
  --version v1.7.2 \
  --set installCRDs=true
watch kubectl -n cert-manager get pods # ctrl+c the break
```

Install Addon for cert-manager that issues certificates using AWS Private CA Issuer
```bash
helm repo add awspca https://cert-manager.github.io/aws-privateca-issuer && helm repo update
# helm search repo awspca/aws-privateca-issuer --versions
helm -n aws-privateca-issuer upgrade -i aws-privateca-issuer awspca/aws-privateca-issuer \
  --create-namespace \
  --version 1.2.1
watch kubectl -n aws-privateca-issuer get pods # ctrl+c the break 
```

## Hammer the external load balancer

In a **dedicated** terminal window run a looped command against the **frontend** NLB.
```bash
nlb_dnsname=$(kubectl -n demos get service gw-echo-frontend -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${nlb_dnsname}; sleep 0.25; done
# ctrl+c to quit loop
```

## Create a private certificate authority

```bash
CA_ARN=$(
  aws acm-pca create-certificate-authority \
    --certificate-authority-type ROOT \
    --query "CertificateAuthorityArn" \
    --output text \
    --certificate-authority-configuration file://<(cat <<EOF
      {
        "KeyAlgorithm": "RSA_2048",
        "SigningAlgorithm": "SHA256WITHRSA",
        "Subject": {
          "Organization": "demos"
        }
      }
EOF
    )
)
```

## Install Root CA (via console)

Locate the Private certificate authority named **demos** at `https://us-west-2.console.aws.amazon.com/acm-pca`, select it and click `Action`-> `Install CA certificate`, accept the defaults then hit `Confirm and install`.

## Link the AWS PCA issuer to your private CA

```bash
mkdir -p ~/environment/cert-manager/
cat <<EOF | tee ~/environment/cert-manager/pca-issuer.yaml | kubectl -n demos apply -f -
apiVersion: awspca.cert-manager.io/v1beta1
kind: AWSPCAIssuer
metadata:
  name: pca-issuer
spec:
  arn: ${CA_ARN}
  region: ${AWS_DEFAULT_REGION}
EOF
sleep 5 && kubectl -n demos get AWSPCAIssuer
```

## Ask cert-manager to generate a certificate and its associated secret

Deploy a cert-manager `Certificate` for the App Mesh service endpoint (dst) then review cert and secret.
```bash
cat <<EOF | tee ~/environment/cert-manager/pca-cert.yaml | kubectl -n demos apply -f -
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: pca-cert
spec:
  commonName: '*.demos.svc.cluster.local'
  dnsNames:
    - '*.demos.svc.cluster.local'
  usages:
    - server auth
    - client auth
  duration: 2160h0m0s   # 90 days
  renewBefore: 360h0m0s # 15 days
  secretName: pca-secret
  issuerRef:
    group: awspca.cert-manager.io
    kind: AWSPCAIssuer
    name: pca-issuer
EOF

kubectl -n demos get certificates -o wide
kubectl -n demos get secrets
```

Take a closer look at a certificate
```bash
kubectl -n demos get secret pca-secret -o 'go-template={{index .data "tls.crt"}}' | base64 --decode | openssl x509 -noout -text
```

## Encrypt inter-pod traffic via App Mesh configuration

Patch the deployments with the following App Mesh specific annotations, telling **envoy** where to place its secrets mount point.
The exception to this is `gw-echo-frontend` where the mount point is set explicitly.

Note that changing the mount points will cause the pods to restart so use your dedicated "watch" terminal to WAIT until the pods have been successfully replaced.

```bash
for target_deploy in echo-backend-blue echo-backend-green echo-frontend-blue; do
  kubectl -n demos patch deployment ${target_deploy} \
    -p '{"spec":{"template":{"metadata":{"annotations":{"appmesh.k8s.aws/secretMounts": "pca-secret:/etc/keys"}}}}}'
done

# the gateway needs a manual patch, but this achieves the same thing
kubectl -n demos patch deployment gw-echo-frontend --patch "$(cat <<EOF
{
    "spec": {
        "template": {
            "spec": {
                "containers": [
                    {
                        "name": "envoy",
                        "volumeMounts": [
                            {
                                "mountPath": "/etc/keys",
                                "name": "pca-secret",
                                "readOnly": true
                            }
                        ]
                    }
                ],
                "volumes": [
                    {
                        "name": "pca-secret",
                        "secret": {
                            "secretName": "pca-secret"
                        }
                    }
                ]
            }
        }
    }
}
EOF
)"
```

Verify that the secret containing the keys was correctly mounted and is accessible from within `envoy`.
```bash
for target_deploy in echo-backend-blue echo-backend-green echo-frontend-blue gw-echo-frontend; do
  echo "tls.crt in ${target_deploy}:"
  kubectl -n demos exec -it deploy/${target_deploy} -c envoy -- cat /etc/keys/tls.crt
done

Tell App Mesh to start using the cert across all the workloads (will not cause pod restarts)
```bash
for target_vn in vn-echo-backend-blue vn-echo-backend-green vn-echo-frontend-blue; do
  kubectl -n demos patch virtualnode ${target_vn} --type='json' \
    -p='[{"op": "add", "path": "/spec/listeners/0/tls", "value": {"mode": "STRICT","certificate": {"file": {"certificateChain": "/etc/keys/tls.crt", "privateKey": "/etc/keys/tls.key"} } } }]'
done
```

<!-- Enforce certificate validation (like your browser does for public CAs)
```bash
for target_vn in vn-echo-backend-blue vn-echo-backend-green vn-echo-frontend-blue; do
  kubectl -n ${target_ns} patch virtualnode ${target_vn} --type='json' \
    -p='[{"op": "remove", "path":  "/spec/backendDefaults", "value": {"clientPolicy": {"tls": {"enforce": true, "validation": {"trust": {"file": {"certificateChain": "/etc/keys/ca.crt"}}}}}} }]'
done
``` -->

Return to the terminal window running a looped command against the **frontend** NLB.
It should be business as usual.

To further validate that all is well, we can exec into the frontend and curl the backend as before.
As this command is issued **inside** your frontend which is **encapsulated** by the mesh it works as before whilst looking just like plain old HTTP.
```bash
kubectl -n demos exec -it deploy/echo-frontend-blue -c echo-frontend -- curl http://vs-echo-backend.demos.svc.cluster.local:80
```

Similar commands initiated **outside** the mesh will now fail in various ways.
```bash
kubectl exec -it jumpbox -- curl http://vs-echo-backend.demos.svc.cluster.local:80
kubectl exec -it jumpbox -- curl http://echo-frontend-blue.demos.svc.cluster.local:80
```

Verify with `envoy` that SSL metrics are being recorded.
```bash
for target_deploy in echo-backend-blue echo-backend-green echo-frontend-blue gw-echo-frontend; do
  echo "SSL handshake stats for ${target_deploy}:"
  kubectl -n ${target_ns} exec -it deploy/${target_deploy} -c envoy -- curl -s localhost:9901/stats | grep ssl.handshake
done
```

If you see non-zero responses for `ssl.handshake` it's because traffic between the **frontend** and **backend** components is now encrypted.

## TODO list (FUBAR!)

- figure out why **enforced certificate validation** fails for these **frontend/backend** pods. I've seen it working for plain old nginx?!?!
- figure out why acm-pca certs cannot be applied to `aws-load-balancer-ssl-cert` ELB annotations.

<!--

But traffic between the **NLB** and the **frontend** which passes through your VirtualGateway's deployment `gw-echo-frontend` remains in plaintext.
Let's address that now.

# TODO might need to use ACM rather than PCA

```bash
# use the following command to identity the certficate ARN
kubectl -n aws-privateca-issuer logs deploy/aws-privateca-issuer | grep arn:aws:acm-pca | tail -1 | jq .msg --raw-output
#e.g. CERT_ARN=aws:acm-pca:us-west-2:390758498079:certificate-authority/e325aacf-e92c-4bd4-a375-744d97f03474/certificate/c0ade5bcad98402b1b0d66662074fed7
#e.g. CERT_ARN=arn:aws:acm:us-west-2:390758498079:certificate/56e9cb4d-79ba-49f5-9e8d-86988a6dbb61
CERT_ARN=<cert_arn_from_logs>

kubectl -n demos patch service gw-echo-frontend --patch "$(cat <<EOF
{
    "metadata": {
        "annotations": {
            "service.beta.kubernetes.io/aws-load-balancer-ssl-cert": "${CERT_ARN}",
            "service.beta.kubernetes.io/aws-load-balancer-backend-protocol": "ssl"
        }
    }
}
EOF
)"
```

-->

## Rollback

Revert the changes from this chapter as follows.
```bash
# remove the mesh
helm -n demos uninstall mesh

# remove the namespace labels
kubectl label namespace demos mesh-
kubectl label namespace demos appmesh.k8s.aws/sidecarInjectorWebhook-

# bounce the backends (they will lose thier envoy sidecars)
kubectl -n demos rollout restart deployment \
  echo-backend-blue \
  echo-backend-green

# set the backend env var in echo-frontend-blue to re-point at the original echo-backend-blue.
helm -n demos upgrade -i echo-frontend-blue ~/environment/echo-frontend/ \
  --create-namespace \
  --set registry=${EKS_ECR_REGISTRY} \
  --set color=blue \
  --set version=2.0 \
  --set backend=http://echo-backend-blue.demos.svc.cluster.local:80 \
  --set serviceType=ClusterIP

# The NLB no longer exists so revert to the jumpbox for ingress
kubectl exec -it jumpbox -- /bin/bash -c "while true; do curl http://echo-frontend-blue.demos.svc.cluster.local:80; sleep 0.25; done"
```

[Return To Main Menu](/README.md)
