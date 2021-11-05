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

# deploy primary set of backend components (blue)
kubectl -n ${EKS_APP_NS} create deployment ${EKS_APP_BE}-blue --replicas 0 --image ${EKS_APP_BE_ECR_REPO}:${EKS_APP_BE_VERSION} # begin with zero replicas
kubectl -n ${EKS_APP_NS} set resources deployment ${EKS_APP_BE}-blue --requests=cpu=200m,memory=200Mi                           # right-size the pods
kubectl -n ${EKS_APP_NS} patch deployment ${EKS_APP_BE}-blue --patch="{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"${EKS_APP_BE}\",\"imagePullPolicy\":\"Always\"}]}}}}"
kubectl -n ${EKS_APP_NS} scale deployment ${EKS_APP_BE}-blue --replicas 2                                                       # start 2 instances
kubectl -n ${EKS_APP_NS} expose deployment ${EKS_APP_BE}-blue --port=80 --type=ClusterIP

# deploy secondary set backend components (green)
kubectl -n ${EKS_APP_NS} create deployment ${EKS_APP_BE}-green --replicas 0 --image ${EKS_APP_BE_ECR_REPO}:${EKS_APP_BE_VERSION_NEXT} # begin with zero replicas
kubectl -n ${EKS_APP_NS} set resources deployment ${EKS_APP_BE}-green --requests=cpu=200m,memory=200Mi                                # right-size the pods
kubectl -n ${EKS_APP_NS} patch deployment ${EKS_APP_BE}-green --patch="{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"${EKS_APP_BE}\",\"imagePullPolicy\":\"Always\"}]}}}}"
kubectl -n ${EKS_APP_NS} scale deployment ${EKS_APP_BE}-green --replicas 2                                                            # start 2 instances
kubectl -n ${EKS_APP_NS} expose deployment ${EKS_APP_BE}-green --port=80 --type=ClusterIP

# inspect out two sets of backend objects
sleep 10 && kubectl -n ${EKS_APP_NS} get deployments,services -o wide                                                                 # inspect objects

# TODO TEST FROM HERE

# create the mesh component itself - this contains all the virtual elements which define communications within its scope.
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
kubectl -n ${EKS_APP_NS} apply -f ~/environment/eks-demos/src/apps-mesh/mesh.yaml
sleep 2 && kubectl -n ${EKS_APP_NS} get meshes
aws appmesh describe-mesh --mesh-name ${EKS_APP_NS}  # check it produced an AWS resource

# activate the namespace for use with the mesh
kubectl label namespaces ${EKS_APP_NS} mesh=${EKS_APP_NS}
kubectl label namespaces ${EKS_APP_NS} appmesh.k8s.aws/sidecarInjectorWebhook=enabled

# create a virtual node for the backend
cat > ~/environment/eks-demos/src/apps-mesh/vn-${EKS_APP_BE}-blue.yaml << EOF
apiVersion: appmesh.k8s.aws/v1beta2
kind: VirtualNode
metadata:
  name: vn-${EKS_APP_BE}-blue
spec:
  awsName: vn-${EKS_APP_BE}-blue
  podSelector:
    matchLabels:
      app: ${EKS_APP_BE}-blue
  listeners:
    - portMapping:
        port: 80
        protocol: http
  serviceDiscovery:
    dns:
      hostname: ${EKS_APP_BE}-blue.${EKS_APP_NS}.svc.cluster.local
EOF
kubectl -n ${EKS_APP_NS} apply -f ~/environment/eks-demos/src/apps-mesh/vn-${EKS_APP_BE}-blue.yaml

# create a virtual router to control traffic shifting




kubectl -n ${EKS_APP_NS} rollout restart deployment ${EKS_APP_FE}
kubectl -n ${EKS_APP_NS} rollout restart deployment ${EKS_APP_BE}-blue
