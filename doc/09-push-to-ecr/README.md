# Push Container Image To ECR

Create target ECR repo, deleting it first if needed.
```bash
aws ecr delete-repository --repository-name echo-frontend --force >/dev/null 2>&1
aws ecr create-repository \
  --repository-name echo-frontend \
  --image-scanning-configuration scanOnPush=true \
  --query 'repository.repositoryUri' \
  --output text
```

Tag your image in preparation for uploading to ECR and review the contents of the local Docker image cache.
```bash
docker tag echo-frontend:1.0 ${EKS_ECR_REGISTRY}/echo-frontend:1.0
docker images
```

Authenticate the Docker CLI with ECR and push the image to the ECR repository.
```bash
aws ecr get-login-password | docker login --username AWS --password-stdin ${EKS_ECR_REGISTRY}
docker push ${EKS_ECR_REGISTRY}/echo-frontend:1.0
```

The EKS cluster will now be able to locate this image in ECR by its version tag.
Review the ECR repository for your app.
```bash
aws ecr list-images --repository-name echo-frontend
```

Next: [Main Menu](/README.md) | [Deploy From ECR To Kubernetes](../10-deploy-to-k8s/README.md)
