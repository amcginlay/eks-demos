# Push Container Image To ECR

Create target ECR repo, deleting it first if needed.
```bash
aws ecr delete-repository --repository-name ${EKS_APP_FE} --force >/dev/null 2>&1
aws ecr create-repository \
  --repository-name ${EKS_APP_FE} \
  --region ${AWS_DEFAULT_REGION} \
  --image-scanning-configuration scanOnPush=true \
  --query 'repository.repositoryUri' \
  --output text
```

NOTE the generated repo name, which is output by the above commands, will match our exported variable named `EKS_APP_FE_ECR_REPO`.

Push the Docker images to the ECR repository.
```bash
aws ecr get-login-password --region ${AWS_DEFAULT_REGION} | docker login --username AWS --password-stdin ${EKS_ECR_REGISTRY}
docker tag ${EKS_APP_FE}:${EKS_APP_FE_VERSION} ${EKS_APP_FE_ECR_REPO}:${EKS_APP_FE_VERSION}
docker tag ${EKS_APP_FE}:${EKS_APP_FE_VERSION_NEXT} ${EKS_APP_FE_ECR_REPO}:${EKS_APP_FE_VERSION_NEXT}
docker images
docker push ${EKS_APP_FE_ECR_REPO}:${EKS_APP_FE_VERSION}
docker push ${EKS_APP_FE_ECR_REPO}:${EKS_APP_FE_VERSION_NEXT}
```

The EKS cluster can now locate these images by their version tags.
```bash
aws ecr list-images --repository-name ${EKS_APP_FE}
```

[Return To Main Menu](/README.md)
