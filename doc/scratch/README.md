**App Mesh WIP**

eksctl create iamserviceaccount \
  --cluster ${EKS_CLUSTER_NAME} \
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

# TODO we should consider running the backend in a net new namespace, but only once we're sure it's working inside a shared one
# this might be critically important because as it stands ALL deployments need an associated virtualnode to avoid restart failures

# deploy primary and secondary sets of backend components (BLUE and GREEN)
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
sleep 10 && kubectl -n ${EKS_APP_NS} get deployments,services -o wide

# create the mesh component itself - this contains all the virtual elements which define communications within its scope.
# NOTE unlike the components it will later encapsulate, the mesh is not a namespaced k8s resource.
mkdir -p ~/environment/eks-demos/src/apps-mesh                                                       # <<<< TODO change name to mesh-apps or something which makes the overall file naming more logical
cat > ~/environment/eks-demos/src/apps-mesh/mesh.yaml << EOF
apiVersion: appmesh.k8s.aws/v1beta2
kind: Mesh
metadata:
  name: ${EKS_APP_NS}
spec:
  egressFilter:
    type: ALLOW_ALL       # without this setting pods will have no external connectivity - I think this would also prohibit private DNS resolution, so it's a big deal!
  namespaceSelector:
    matchLabels:
      mesh: ${EKS_APP_NS}
EOF
kubectl apply -f ~/environment/eks-demos/src/apps-mesh/mesh.yaml

# observe the change
kubectl -n ${EKS_APP_NS} get meshes                 # check from the k8s aspect
aws appmesh describe-mesh --mesh-name ${EKS_APP_NS} # check from the AWS aspect

# label the namespace for use with App Mesh then observe the change           ##### TODO maybe shift this down to just before when the deployments are ALL restarted
kubectl label namespaces ${EKS_APP_NS} mesh=${EKS_APP_NS} --overwrite
kubectl label namespaces ${EKS_APP_NS} appmesh.k8s.aws/sidecarInjectorWebhook=enabled --overwrite
kubectl get namespace ${EKS_APP_NS} -o yaml | kubectl neat

# create virtual nodes for the BLUE and GREEN backend
for color in blue green; do
cat > ~/environment/eks-demos/src/apps-mesh/vn-${EKS_APP_BE}-${color}.yaml << EOF
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
kubectl -n ${EKS_APP_NS} apply -f ~/environment/eks-demos/src/apps-mesh/vn-${EKS_APP_BE}-${color}.yaml
done

# observe the change
kubectl -n ${EKS_APP_NS} get virtualnodes                # check from the k8s aspect
aws appmesh list-virtual-nodes --mesh-name ${EKS_APP_NS} # check from the AWS aspect

# confirm that the backend pods currently have one container each (READY 1/1)
kubectl -n ${EKS_APP_NS} get pods

# to see the effect of labeling the namespace
# upon restart, each backend pod will be injected with two additional containers (sidecars)
# one for envoy, the other is the x-ray daemon which we enabled when we installed the App Mesh controller
for color in blue green; do
  kubectl -n ${EKS_APP_NS} rollout restart deployment ${EKS_APP_BE}-${color}
done

# observe as the pods are restarted
watch kubectl -n ${EKS_APP_NS} get pods                   # ctrl+c to quit loop

# the frontend virtualnode was defered until now because it depends upon the backend being fully meshified
# that involves introducing a virtualrouter and a virtualservice to marshal requests through to the backends

# a virtualrouter will distribute backend requests (blue and green) using weighted routes
# create a single virtualrouter which, for now, sends 100% of requests to blue
cat > ~/environment/eks-demos/src/apps-mesh/vr-${EKS_APP_BE}.yaml << EOF
apiVersion: appmesh.k8s.aws/v1beta2
kind: VirtualRouter
metadata:
  name: vr-${EKS_APP_BE}
  namespace: ${EKS_APP_NS}
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
kubectl apply -f ~/environment/eks-demos/src/apps-mesh/vr-${EKS_APP_BE}.yaml

# observe the change
kubectl -n ${EKS_APP_NS} get virtualrouters                                                                       # check from the k8s aspect
aws appmesh describe-route --mesh-name ${EKS_APP_NS} --virtual-router-name vr-${EKS_APP_BE} --route-name vrr-${EKS_APP_BE} # check from the AWS aspect

# it's important to ensure that a matching k8s service exists for each virtualservice
# it doesn't need to resolve to any pods, it just need to surface a cluster IP address
kubectl -n ${EKS_APP_NS} create service clusterip vs-${EKS_APP_BE} --tcp=80 -o yaml --dry-run=client | kubectl neat > ~/environment/eks-demos/src/apps-mesh/svc-vs-${EKS_APP_BE}.yaml
kubectl -n ${EKS_APP_NS} apply -f ~/environment/eks-demos/src/apps-mesh/svc-vs-${EKS_APP_BE}.yaml

# in the same way that k8s pods send requests to other pods via k8s services, 
# virtualnodes (which wrap k8s services) send requests to other virtualnodes via virtualservices
# we already have a virtualrouter, which knows how to locate the backend virtualnodes
# now we create a single virtualservice that forwards all its traffic to the virtualrouter
cat > ~/environment/eks-demos/src/apps-mesh/vs-${EKS_APP_BE}.yaml << EOF
apiVersion: appmesh.k8s.aws/v1beta2
kind: VirtualService
metadata:
  name: vs-${EKS_APP_BE}
  namespace: ${EKS_APP_NS}
spec:
  awsName: vs-${EKS_APP_BE}
  provider:
    virtualRouter:
      virtualRouterRef:
        name: vr-${EKS_APP_BE}
EOF
kubectl apply -f ~/environment/eks-demos/src/apps-mesh/vs-${EKS_APP_BE}.yaml

# observe the change
kubectl -n ${EKS_APP_NS} get virtualservices                                                           # check from the k8s aspect
aws appmesh describe-virtual-service --mesh-name ${EKS_APP_NS} --virtual-service-name vs-${EKS_APP_BE} # check from the AWS aspect

# finally we build a virtualnode for the frontend which is required before envoy can be injected there
# we could have applied this manifest before now, but since it's dependencies were not yet available
# it would be stuck in a pending state without the means to produce a corresponding AWS resource

# TODO we don't expect any inbound traffic so we should be able to drop the service discovery bit for frontend

cat > ~/environment/eks-demos/src/apps-mesh/vn-${EKS_APP_FE}.yaml << EOF
apiVersion: appmesh.k8s.aws/v1beta2
kind: VirtualNode
metadata:
  name: vn-${EKS_APP_FE}
  namespace: ${EKS_APP_NS}
spec:
  awsName: vn-${EKS_APP_FE}
  podSelector:
    matchLabels:
      app: ${EKS_APP_FE}
  listeners:
    - portMapping:
        port: 80
        protocol: http
  serviceDiscovery:
    dns:
      hostname: ${EKS_APP_FE}.${EKS_APP_NS}.svc.cluster.local                     # <<<<<<< probably not required. If not, then virtual nodes can be created together ^^^^^ !!!!!
  backends:
    - virtualService:
        virtualServiceRef:
          name: vs-${EKS_APP_BE}
EOF
kubectl apply -f ~/environment/eks-demos/src/apps-mesh/vn-${EKS_APP_FE}.yaml

# observe the change
kubectl -n ${EKS_APP_NS} get virtualnodes                # check from the k8s aspect
aws appmesh list-virtual-nodes --mesh-name ${EKS_APP_NS} # check from the AWS aspect

# the frontend will use an environment variable named BACKEND if present so set that variable now
kubectl -n ${EKS_APP_NS} set env deployment ${EKS_APP_FE} BACKEND=vs-${EKS_APP_BE}                # current Envoy limitation on use of FQDNs

# restart the frontend to see the new sidecars
kubectl -n ${EKS_APP_NS} rollout restart deployment ${EKS_APP_FE}
watch kubectl -n ${EKS_APP_NS} get pods                   # ctrl+c to quit loop

# In a **dedicated** terminal window observe the frontend routing 100% of traffic to the BLUE backend and 0% to green
clb_dnsname=$(kubectl -n ${EKS_APP_NS} get service ${EKS_APP_FE} -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${clb_dnsname}; sleep 0.25; done

# apply an updated manifest in which the weights are flipped to route 0% of traffic to the BLUE backend and 100% to green
# then observe the **dedicated** terminal window as the backend traffic switched from BLUE to GREEN
sed -i -e "s/weight: 100/weight: -1/g" -e "s/weight: 0/weight: 100/g" -e "s/weight: -1/weight: 0/g" ~/environment/eks-demos/src/apps-mesh/vr-${EKS_APP_BE}.yaml
kubectl -n ${EKS_APP_NS} apply -f ~/environment/eks-demos/src/apps-mesh/vr-${EKS_APP_BE}.yaml