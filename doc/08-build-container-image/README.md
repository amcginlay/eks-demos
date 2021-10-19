# Build A Container Image

If you followed the previous steps the source code will have already been cloned onto your Cloud9 instance.

The target for our first image is a simple PHP app hosted as a single file which you can review here [eks-demos/src/php-echo/index.php](/src/php-echo/index.php).
Satisfy yourself that your can **run** this code inside Cloud9. To test, you can `curl http://localhost:8080/eks-demos/src/php-echo` from the Cloud9 terminal window. 
As you do so, observe that the recorded value of **ec2IP** is equivalent to **localhostIP** within this execution environment.

NOTE the use of [Instance Metadata](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) **(169.254.169.254)** within the source code is an indication that our app is tailor-made for deployment on EC2 instances.

Construct the Dockerfile for this application from a template you cloned earlier. If you've not seen `envsubst` before, see [here](https://stackoverflow.com/questions/14155596/how-to-substitute-shell-variables-in-complex-text-files)
```bash
envsubst < ~/environment/eks-demos/src/php-echo/Dockerfile.template > ~/environment/eks-demos/src/php-echo/Dockerfile
```

Inspect the resultant Dockerfile which initializes a container-scoped environment variable named **VERSION**.
```bash
cat ~/environment/eks-demos/src/php-echo/Dockerfile
```

Each Cloud9 instance has the Docker daemon installed. Build the Docker image from the Cloud9 terminal then run the newly containerized app.
```bash
docker build -t ${EKS_APP}:${EKS_APP_VERSION} ~/environment/eks-demos/src/${EKS_APP}/
docker images                                                                     # see what you produced
docker ps                                                                         # nothing running ...
container_id=$(docker run --detach --rm -p 8081:80 ${EKS_APP}:${EKS_APP_VERSION}) # request docker to instantiate a single container as a background process
docker ps                                                                         # ... now one container running
```

Invoke the webserver from inside the container.
```bash
docker exec -it ${container_id} curl localhost:80
```

Invoke the webserver from outside the container.
```bash
curl localhost:8081
```

If you wondered why the localhostIP now differs from the ec2IP ...
```bash
docker network inspect bridge | jq  .[0].IPAM.Config[0].Subnet
```

Have Docker build the next version of our simple app so we've got something extra to play with later on.
This might usually involve some real code changes.
In this case we're just incrementing the value of the `VERSION` environment variable inside the container image.

```bash
sed -i "s/ENV VERSION=${EKS_APP_VERSION}/ENV VERSION=${EKS_APP_VERSION_NEXT}/g" ./eks-demos/src/php-echo/Dockerfile
docker build -t ${EKS_APP}:${EKS_APP_VERSION_NEXT} ~/environment/eks-demos/src/${EKS_APP}/
```

We're done with Docker for now so stop the container (which will be terminated because we ran it with the --rm flag).
```bash
docker stop ${container_id}
```

[Return To Main Menu](/README.md)
