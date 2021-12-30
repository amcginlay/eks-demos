# eks-demos
A selection of demos guiding you through the maze of [Elastic Kubernetes Service (EKS)](https://aws.amazon.com/eks)

## Setup
* [01. Prepare Your CloudShell Environment](doc/01-cloudshell/README.md)
* [02. Configure IAM Role](doc/02-iam-role/README.md)
* [03. Create Cloud9 (EC2) Environment](doc/03-cloud9/README.md)
* [04. Set Variables](doc/04-set-variables/README.md)
* [05. Configure Client Tools](doc/05-client-tools/README.md)
* [06. Build EKS Cluster](doc/06-build-cluster/README.md)
* [07. Configure Local Machine Access](doc/07-local-access/README.md)

## Workload Deployment
* [08. Build A Container Image](doc/08-build-container-image/README.md)
* [09. Push Container Image To ECR](doc/09-push-to-ecr/README.md)
* [10. Deploy From ECR To Kubernetes](doc/10-deploy-to-k8s/README.md)
* [11. Upgrade A Deployment](doc/11-upgrade-deployment/README.md)

## Load Distribution
* [12. K8s ClusterIP Services](doc/12-clusterip-services/README.md) - because pods need to talk to each other
* [13. K8s NodePort Services](doc/13-nodeport-services/README.md) - because workloads outside the cluster need to talk to pods
* [14. K8s LoadBalancer Services](doc/14-loadbalancer-services/README.md) - because the world needs to talk to our cluster
* [15. AWS Load Balancer Controller](doc/15-aws-load-balancer-controller/README.md) - because one load balancer per service is wasteful

## Autoscaling
* [16. Cluster Autoscaler](doc/16-ca/README.md) - because no one likes a pending pod
* [17. Horizonal Pod Autoscaler](doc/17-hpa/README.md) - because demand for pods can grow

## Release Management
* [18. Helm](doc/18-helm/README.md) - because packages need managing

## Service Mesh
* (WIP) [19. AWS App Mesh](doc/19-appmesh/README.md) - because managing microservices at scale is hard
