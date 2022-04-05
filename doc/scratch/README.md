**App Mesh WIP**

If you have completed the earlier section on **LoadBalancer services** then you will already have a load balancer (CLB) in front of the `EKS_APP_FE` (i.e. `echo-frontend`) app.
If you do not have this, execute the following (2-3 mins).
```bash
kubectl -n ${EKS_APP_NS} expose deployment ${EKS_APP_FE} --port=80 --type=LoadBalancer
```

# createing the service account via eksctl ensure that IRSA is correctly configured (note the policy attachment)
eksctl create iamserviceaccount \
  --cluster ${C9_PROJECT} \
  --namespace kube-system \
  --name appmesh-controller \
  --attach-policy-arn arn:aws:iam::aws:policy/AWSCloudMapFullAccess,arn:aws:iam::aws:policy/AWSAppMeshFullAccess \
  --override-existing-serviceaccounts \
  --approve

helm repo add eks https://aws.github.io/eks-charts
helm upgrade -i appmesh-controller eks/appmesh-controller \
  --namespace kube-system \
  --set region=${AWS_DEFAULT_REGION} \
  --set serviceAccount.create=false \
  --set serviceAccount.name=appmesh-controller \
  --set tracing.enabled=true \
  --set tracing.provider=x-ray

kubectl -n kube-system get deployment appmesh-controller

# the aim is to use App Mesh to demonstrate a Blue Green deployment model within which traffic can be dynamically shifted between two different versions of fully-scaled backends
# create these backend images using tools you are already familiar with
# NOTE the generated repo name will match your exported variable named `EKS_APP_BE_ECR_REPO`.
envsubst < ~/environment/eks-demos/src/${EKS_APP_BE}/Dockerfile.template > ~/environment/eks-demos/src/${EKS_APP_BE}/Dockerfile
docker build -t ${EKS_APP_BE}:${EKS_APP_BE_VERSION} ~/environment/eks-demos/src/${EKS_APP_BE}/
sed -i "s/ENV VERSION=${EKS_APP_BE_VERSION}/ENV VERSION=${EKS_APP_BE_VERSION_NEXT}/g" ~/environment/eks-demos/src/${EKS_APP_BE}/Dockerfile
docker build -t ${EKS_APP_BE}:${EKS_APP_BE_VERSION_NEXT} ~/environment/eks-demos/src/${EKS_APP_BE}/
aws ecr delete-repository --repository-name ${EKS_APP_BE} --force >/dev/null 2>&1
aws ecr create-repository \
  --repository-name ${EKS_APP_BE} \
  --region ${AWS_DEFAULT_REGION} \
  --image-scanning-configuration scanOnPush=true \
  --query 'repository.repositoryUri' \
  --output text
aws ecr get-login-password --region ${AWS_DEFAULT_REGION} | docker login --username AWS --password-stdin ${EKS_ECR_REGISTRY}
docker tag ${EKS_APP_BE}:${EKS_APP_BE_VERSION} ${EKS_APP_BE_ECR_REPO}:${EKS_APP_BE_VERSION}
docker tag ${EKS_APP_BE}:${EKS_APP_BE_VERSION_NEXT} ${EKS_APP_BE_ECR_REPO}:${EKS_APP_BE_VERSION_NEXT}
docker images
docker push ${EKS_APP_BE_ECR_REPO}:${EKS_APP_BE_VERSION}
docker push ${EKS_APP_BE_ECR_REPO}:${EKS_APP_BE_VERSION_NEXT}
aws ecr list-images --repository-name ${EKS_APP_BE}

# deploy two sets of independently versioned backend components (BLUE and GREEN)
declare -A image_tags=()
image_tags[blue]=${EKS_APP_BE_VERSION}
image_tags[green]=${EKS_APP_BE_VERSION_NEXT}
for color in blue green; do
  kubectl -n ${EKS_APP_NS} create deployment ${EKS_APP_BE}-${color} --replicas 0 --image ${EKS_APP_BE_ECR_REPO}:${image_tags[${color}]} # begin with zero replicas
  kubectl -n ${EKS_APP_NS} set resources deployment ${EKS_APP_BE}-${color} --requests=cpu=200m,memory=200Mi                           # right-size the pods
  kubectl -n ${EKS_APP_NS} patch deployment ${EKS_APP_BE}-${color} --patch="{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"${EKS_APP_BE}\",\"imagePullPolicy\":\"Always\"}]}}}}"
  kubectl -n ${EKS_APP_NS} scale deployment ${EKS_APP_BE}-${color} --replicas 2                                                       # start 2 instances
  kubectl -n ${EKS_APP_NS} expose deployment ${EKS_APP_BE}-${color} --port=80 --type=ClusterIP
done

# inspect the two sets of backend objects.
sleep 10 && kubectl -n ${EKS_APP_NS} get deployments,services,pods -o wide

# create the mesh component itself - this contains all the virtual elements which define communications within its scope.
# NOTE unlike the components it will later encapsulate, the mesh is not a namespaced k8s resource.
mkdir -p ~/environment/eks-demos/src/mesh-apps/
cat > ~/environment/eks-demos/src/mesh-apps/mesh.yaml << EOF
apiVersion: appmesh.k8s.aws/v1beta2
kind: Mesh
metadata:
  name: ${EKS_APP_NS}
spec:
  egressFilter:
    type: ALLOW_ALL                 # permit pods inside the mesh to communicate externally
  namespaceSelector:
    matchLabels:
      mesh: ${EKS_APP_NS}           # associate this mesh with appropriately labeled namespaces
EOF
kubectl apply -f ~/environment/eks-demos/src/mesh-apps/mesh.yaml

# observe the change
kubectl -n ${EKS_APP_NS} get meshes                 # check from the k8s aspect
aws appmesh describe-mesh --mesh-name ${EKS_APP_NS} # check from the AWS aspect

# label the namespace for use with App Mesh then observe the change
# this must be done before any virtual components can be created inside the namespace
kubectl label namespaces ${EKS_APP_NS} mesh=${EKS_APP_NS} --overwrite
kubectl label namespaces ${EKS_APP_NS} appmesh.k8s.aws/sidecarInjectorWebhook=enabled --overwrite
kubectl get namespace ${EKS_APP_NS} -o yaml | kubectl neat

# create virtual nodes for the BLUE and GREEN backend
# NOTE the backend nodes require `listeners` and `serviceDiscovery` sections as the underlying service expects inbound traffic 
for color in blue green; do
cat > ~/environment/eks-demos/src/mesh-apps/vn-${EKS_APP_BE}-${color}.yaml << EOF
apiVersion: appmesh.k8s.aws/v1beta2
kind: VirtualNode
metadata:
  name: vn-${EKS_APP_BE}-${color}
spec:
  awsName: vn-${EKS_APP_BE}-${color}
  podSelector:
    matchLabels:
      app: ${EKS_APP_BE}-${color}
  listeners:
    - portMapping:
        port: 80
        protocol: http
  serviceDiscovery:
    dns:
      hostname: ${EKS_APP_BE}-${color}.${EKS_APP_NS}.svc.cluster.local
EOF
kubectl -n ${EKS_APP_NS} apply -f ~/environment/eks-demos/src/mesh-apps/vn-${EKS_APP_BE}-${color}.yaml
done

# observe the change
kubectl -n ${EKS_APP_NS} get virtualnodes                # check from the k8s aspect
aws appmesh list-virtual-nodes --mesh-name ${EKS_APP_NS} # check from the AWS aspect

# confirm that the backend pods currently have one container each (READY 1/1)
kubectl -n ${EKS_APP_NS} get pods

# with the backend virtual nodes in place you can now deploy a virtual router which will distribute backend requests (BLUE and GREEN) using weighted routes
# create a single virtualrouter which, for now, sends 100% of requests to BLUE
# NOTE a virtual router, which sits in front of a set of routes to virtual nodes, is only required when traffic shifting is desired.
cat > ~/environment/eks-demos/src/mesh-apps/vr-${EKS_APP_BE}.yaml << EOF
apiVersion: appmesh.k8s.aws/v1beta2
kind: VirtualRouter
metadata:
  name: vr-${EKS_APP_BE}
spec:
  awsName: vr-${EKS_APP_BE}
  listeners:
    - portMapping:
        port: 80
        protocol: http
  routes:
    - name: vrr-${EKS_APP_BE}
      httpRoute:
        match:
          prefix: /
        action:
          weightedTargets:
          - virtualNodeRef:
              name: vn-${EKS_APP_BE}-blue
            weight: 100
          - virtualNodeRef:
              name: vn-${EKS_APP_BE}-green
            weight: 0
EOF
kubectl -n ${EKS_APP_NS} apply -f ~/environment/eks-demos/src/mesh-apps/vr-${EKS_APP_BE}.yaml

# observe the change
kubectl -n ${EKS_APP_NS} get virtualrouters                                                                                # check from the k8s aspect
aws appmesh describe-route --mesh-name ${EKS_APP_NS} --virtual-router-name vr-${EKS_APP_BE} --route-name vrr-${EKS_APP_BE} # check from the AWS aspect

# with the virtual router in place you can now deploy a virtual service and the matching underlying ClusterIP service which it will provide a version independent target for backend traffic.
# the ClusterIP service should never resolve to any pods in the traditional sense, it just surfaces a DNS name and IP address which the mesh can reference internally to anchor its redirections.
# in the same way that k8s pods send requests to other pods via k8s services, virtualnodes send requests to other virtualnodes via virtualservices
# you already have a virtualrouter, which knows how to locate the blue and green backend virtualnodes
# now you can create a single virtualservice that forwards all its traffic to the virtualrouter
cat > ~/environment/eks-demos/src/mesh-apps/vs-${EKS_APP_BE}.yaml << EOF
apiVersion: appmesh.k8s.aws/v1beta2
kind: VirtualService
metadata:
  name: ${EKS_APP_BE}
spec:
  awsName: ${EKS_APP_BE}.${EKS_APP_NS}.svc.cluster.local
  provider:
    virtualRouter:
      virtualRouterRef:
        name: vr-${EKS_APP_BE}
---
EOF
kubectl -n ${EKS_APP_NS} create service clusterip ${EKS_APP_BE} --tcp=80 -o yaml --dry-run=client | kubectl neat >> ~/environment/eks-demos/src/mesh-apps/vs-${EKS_APP_BE}.yaml
kubectl -n ${EKS_APP_NS} apply -f ~/environment/eks-demos/src/mesh-apps/vs-${EKS_APP_BE}.yaml

# observe the change
kubectl -n ${EKS_APP_NS} get virtualservices                                                        # check from the k8s aspect
aws appmesh describe-virtual-service --mesh-name ${EKS_APP_NS} --virtual-service-name ${EKS_APP_BE} # check from the AWS aspect

# the final piece is the virtual node component which represents the frontend deployment.
# you could have applied this manifest before now, but since it's dependencies were not yet available
# it would be held in a partially-complete state without the means to produce a corresponding AWS resource

cat > ~/environment/eks-demos/src/mesh-apps/vn-${EKS_APP_FE}.yaml << EOF
apiVersion: appmesh.k8s.aws/v1beta2
kind: VirtualNode
metadata:
  name: vn-${EKS_APP_FE}
spec:
  awsName: vn-${EKS_APP_FE}
  podSelector:
    matchLabels:
      app: ${EKS_APP_FE}
  backends:
    - virtualService:
        virtualServiceRef:
          name: ${EKS_APP_BE}
EOF
kubectl -n ${EKS_APP_NS} apply -f ~/environment/eks-demos/src/mesh-apps/vn-${EKS_APP_FE}.yaml

# observe the change
kubectl -n ${EKS_APP_NS} get virtualnodes                # check from the k8s aspect
aws appmesh list-virtual-nodes --mesh-name ${EKS_APP_NS} # check from the AWS aspect

# upon restart, each backend pod will be injected with two additional containers (sidecars)
# one is envoy, the other is the x-ray daemon which you enabled when you installed the App Mesh controller
for color in blue green; do
  kubectl -n ${EKS_APP_NS} rollout restart deployment ${EKS_APP_BE}-${color}
done

# if present, the frontend code will look for an environment variable named BACKEND and `curl` that endpoint
# setting this now will cause the frontend deployment to also be restarted and for the backend to become utilized
kubectl -n ${EKS_APP_NS} set env deployment ${EKS_APP_FE} BACKEND=http://${EKS_APP_BE}.${EKS_APP_NS}.svc.cluster.local:80

# observe as the pods are restarted. the new sidecar containers will be injected this time.
watch kubectl -n ${EKS_APP_NS} get pods                   # ctrl+c to quit loop

# In a **dedicated** terminal window observe the frontend routing 100% of traffic to the BLUE backend and 0% to green
clb_dnsname=$(kubectl -n ${EKS_APP_NS} get service ${EKS_APP_FE} -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${clb_dnsname}; sleep 0.25; done

# apply an updated manifest for the virtual router in which the weights are flipped to route 0% of traffic to the BLUE backend and 100% to green
# observe in the **dedicated** terminal window as the backend traffic is switched from BLUE to GREEN (~5-10 secs)
# this was achieved by reconfiguring the App Mesh configuration, which re-publishes that configuration down to all the appropriate envoy proxies, dynamically re-routing the requests flowing out through those containers
# your application remains completely unaware that any of this reconfiguration is taking place
sed -i -e "s/weight: 100/weight: -1/g" -e "s/weight: 0/weight: 100/g" -e "s/weight: -1/weight: 0/g" ~/environment/eks-demos/src/mesh-apps/vr-${EKS_APP_BE}.yaml
kubectl -n ${EKS_APP_NS} apply -f ~/environment/eks-demos/src/mesh-apps/vr-${EKS_APP_BE}.yaml

# the virtual components you applied all have AWS counterparts which can be observed from the App Mesh console at https://console.aws.amazon.com/appmesh/meshes.
# spend a moment to familiarize yourself with the service mesh but do not make any changes.
# when you installed the App Mesh controller you essentially made a commitment to configure your service mesh through the exclusive use of k8s manifests (think, Infrastructure as Code) so you must resist the temptation to make stateful modifications via other tools (e.g. AWS Console/CLI).
# this approach will protect you against configuration drift and make your deployments portable between EKS clusters.

# you may recall you enabled x-ray when you installed the App Mesh controller so your application will also now be emiting trace diagnostics to the X-Ray service.
# head over to the x-ray service map at https://us-west-2.console.aws.amazon.com/xray/home to observe as the traffic begins to shift from BLUE to GREEN
# the default service map window is 5 minutes but you can reduce this to 1 minute.
# as you refresh the service map you will observe the volume of recorded traffic drop from BLUE until the service icon disappears altogether