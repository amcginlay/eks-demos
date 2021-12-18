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

Push the Docker images to the ECR repository.
```bash
aws ecr get-login-password | docker login --username AWS --password-stdin ${EKS_ECR_REGISTRY}
docker tag echo-frontend:1.0 ${EKS_ECR_REGISTRY}/echo-frontend:1.0
docker tag echo-frontend:2.0 ${EKS_ECR_REGISTRY}/echo-frontend:2.0
docker push ${EKS_ECR_REGISTRY}/echo-frontend:1.0
docker push ${EKS_ECR_REGISTRY}/echo-frontend:2.0
```

The EKS cluster can now locate these images by their version tags.
```bash
aws ecr list-images --repository-name echo-frontend
```

[Return To Main Menu](/README.md)
