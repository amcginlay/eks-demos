# eks-demos
A guide through the maze of [Elastic Kubernetes Service (EKS)](https://aws.amazon.com/eks)

## Setup
* [01. Prepare Your CloudShell Environment](doc/01-cloudshell/README.md)
* [02. Configure IAM Role](doc/02-iam-role/README.md)
* [03. Create Cloud9 (EC2) Environment](doc/03-cloud9/README.md)
* [04. Set Variables](doc/04-set-variables/README.md)
* [05. Configure Client Tools](doc/05-client-tools/README.md)
* [06. Clone This Repo](doc/06-clone-repo/README.md)
* [07. Build EKS Cluster](doc/07-build-cluster/README.md)

## Workload Deployment
* [08. Build A Container Image](doc/08-build-container-image/README.md)
* [09. Push Container Image To ECR](doc/09-push-to-ecr/README.md)
* [10. Deploy From ECR To Kubernetes](doc/10-deploy-to-k8s/README.md)

## K8s Services
* [11. K8s ClusterIP Services](doc/11-clusterip-services/README.md) - because pods need to talk to each other
* [12. K8s NodePort Services](doc/12-nodeport-services/README.md) - because workloads outside the cluster need to talk to pods
* [13. K8s LoadBalancer Services](doc/13-loadbalancer-services/README.md) - because the world needs to talk to our cluster
* TODO [14. AWS Load Balancer Controller](doc/14-aws-loadbalancer-controller/README.md) - because one load balancer per service is wasteful

## Autoscaling
* TODO [15. Horizonal Pod Autoscaler](doc/15-hpa/README.md) - because demand for pods can grow
* TODO [16. Cluster Autoscaler](doc/16-ca/README.md) - because no one likes a pending pod

## Extensions
* [Orchestration](doc/orchestration/README.md) - balancing desired against actual
