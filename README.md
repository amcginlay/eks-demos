# eks-demos
A guide through the maze of [Elastic Kubernetes Service (EKS)](https://aws.amazon.com/eks)

## Setup
* [01. Create Cloud9 (EC2) Environment](doc/01-cloud9/README.md)
* [02. Set Variables](doc/02-set-variables/README.md)
* [03. Configure Client Tools](doc/03-client-tools/README.md)
* [04. Clone This Repo](doc/04-clone-repo/README.md)
* [05. Build EKS Cluster](doc/05-build-cluster/README.md)

## Workload Deployment
* [06. Build A Container Image](doc/06-build-container-image/README.md)
* [07. Push Container Image To ECR](doc/07-push-to-ecr/README.md)
* [08. Deploy From ECR To Kubernetes](doc/08-deploy-to-k8s/README.md)

## K8s Services
* [09. K8s ClusterIP Services](doc/09-clusterip-services/README.md) - because pods need to talk to each other
* [10. K8s NodePort Services](doc/10-nodeport-services/README.md) - because workloads outside the cluster need to talk to pods
* [11. K8s LoadBalancer Services](doc/11-loadbalancer-services/README.md) - because the world outside needs to talk to our cluster

## Extensions
* [Orchestration](doc/orchestration/README.md) - balancing desired against actual
