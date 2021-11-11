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

# inspect the two sets of backend objects
sleep 10 && kubectl -n ${EKS_APP_NS} get deployments,services -o wide                                                                 # inspect objects

# create the mesh component itself - this contains all the virtual elements which define communications within its scope.
# NOTE unlike the components it will later encapsulate, the mesh is not a namespaced k8s resource.
mkdir -p ~/environment/eks-demos/src/apps-mesh
cat > ~/environment/eks-demos/src/apps-mesh/mesh.yaml << EOF
apiVersion: appmesh.k8s.aws/v1beta2
kind: Mesh
metadata:
  name: ${EKS_APP_NS}
spec:
  namespaceSelector:
    matchLabels:
      mesh: ${EKS_APP_NS}
EOF
kubectl apply -f ~/environment/eks-demos/src/apps-mesh/mesh.yaml
sleep 2 && kubectl -n ${EKS_APP_NS} get meshes      # check it from the k8s aspect
aws appmesh describe-mesh --mesh-name ${EKS_APP_NS} # check it from the AWS aspect

# label the namespace for use with App Mesh then observe the change
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
sleep 2 && kubectl -n ${EKS_APP_NS} get virtualnodes    # check the nodes from the k8s aspect
aws appmesh list-virtual-nodes --mesh-name ${EKS_APP_NS} # check the nodes from the AWS aspect

>>>>>>> TODO pick up from here


# create a virtual router to control traffic shifting




kubectl -n ${EKS_APP_NS} rollout restart deployment ${EKS_APP_FE}
kubectl -n ${EKS_APP_NS} rollout restart deployment ${EKS_APP_BE}-blue
