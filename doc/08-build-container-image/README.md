# Build A Container Image

If you followed the previous steps the source code will have already been cloned onto your Cloud9 instance.

The target for our first image is a simple PHP app hosted as a single file which you can review here [eks-demos/src/echo-frontend/index.php](/src/echo-frontend/index.php).
Satisfy yourself that your can **run** this code inside Cloud9. To test, you can `curl http://localhost:8080/eks-demos/src/echo-frontend` from the Cloud9 terminal window. 
As you do so, observe that the recorded value of **ec2IP** is equivalent to **localhostIP** within this execution environment.

NOTE the use of [Instance Metadata](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) **(169.254.169.254)** within the source code is an indication that our app is tailor-made for deployment on EC2 instances.

Each Cloud9 instance has the [Docker](https://en.wikipedia.org/wiki/Docker_(software)) daemon installed with a set of images pre-loaded. Remove them as they are not required.
```bash
for i in $(docker ps -q); do docker kill $i; done
docker system prune --all --force
```

Construct the Dockerfile for this application from a template you cloned earlier. If you've not seen `envsubst` before, see [here](https://stackoverflow.com/questions/14155596/how-to-substitute-shell-variables-in-complex-text-files)
```bash
envsubst < ~/environment/eks-demos/src/${EKS_APP_FE}/Dockerfile.template > ~/environment/eks-demos/src/${EKS_APP_FE}/Dockerfile
```

Inspect the resultant Dockerfile which initializes a container-scoped environment variable named **VERSION**.
```bash
cat ~/environment/eks-demos/src/e${EKS_APP_FE}/Dockerfile
```

Each Cloud9 instance has the Docker daemon installed. Build the Docker image from the Cloud9 terminal then run the newly containerized app.
```bash
docker build -t ${EKS_APP_FE}:${EKS_APP_FE_VERSION} ~/environment/eks-demos/src/${EKS_APP_FE}/
docker images                                                                             # see what you produced
docker ps                                                                                 # nothing running ...
container_id=$(docker run --detach --rm -p 8081:80 ${EKS_APP_FE}:${EKS_APP_FE_VERSION}) # request docker to instantiate a single container as a background process
docker ps                                                                                 # ... now one container running
```

Invoke the webserver from inside the container.
```bash
docker exec -it ${container_id} curl localhost:80
```

Invoke the webserver from outside the container.
```bash
curl localhost:8081
```

We are done with running images in Docker for now so stop the container (which will be terminated because we ran it with the --rm flag).
```bash
docker stop ${container_id}
```

If you wondered why the localhostIP now differs from the ec2IP ...
```bash
docker network inspect bridge | jq  .[0].IPAM.Config[0].Subnet
```

Before we move on, instruct Docker to build the **next** version of our simple app and a pair of back-end releases so we've got something extra to play with later on.
This might usually involve some real code changes.
In this case we're just incrementing the value of the `VERSION` environment variable inside the Dockerfile before rebuilding the container images.
```bash
sed -i "s/ENV VERSION=${EKS_APP_FE_VERSION}/ENV VERSION=${EKS_APP_FE_VERSION_NEXT}/g" ~/environment/eks-demos/src/${EKS_APP_FE}/Dockerfile
docker build -t ${EKS_APP_FE}:${EKS_APP_FE_VERSION_NEXT} ~/environment/eks-demos/src/${EKS_APP_FE}/

envsubst < ~/environment/eks-demos/src/${EKS_APP_BE}/Dockerfile.template > ~/environment/eks-demos/src/${EKS_APP_BE}/Dockerfile
docker build -t ${EKS_APP_BE}:${EKS_APP_BE_VERSION} ~/environment/eks-demos/src/${EKS_APP_BE}/
sed -i "s/ENV VERSION=${EKS_APP_BE_VERSION}/ENV VERSION=${EKS_APP_BE_VERSION_NEXT}/g" ~/environment/eks-demos/src/${EKS_APP_BE}/Dockerfile
docker build -t ${EKS_APP_BE}:${EKS_APP_BE_VERSION_NEXT} ~/environment/eks-demos/src/${EKS_APP_BE}/

docker images
```

[Return To Main Menu](/README.md)
