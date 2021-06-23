# Build A Container Image

If you followed the previous steps the source code will have already been cloned onto your Cloud9 instance.

The target for our first image is a simple PHP app hosted as a single file which you can review here [eks-demos/src/php-echo/index.php](/src/php-echo/index.php).
Satisfy yourself that your can **run** this code inside Cloud9. To test, you can `curl http://localhost:8080/eks-demos/src/php-echo` from the Cloud9 terminal window. 
As you do so, observe that the recorded value of **ec2IP** is equivalent to **localhostIP** within this execution environment.

NOTE the use of **169.254.169.254** within [eks-demos/src/php-echo/index.php](/src/php-echo/index.php) is an indication that our app is tailor-made for deployment on EC2 instances.

Construct the Dockerfile for this application locally from a template. The following command initializes the **VERSION** variable (initially set to 1.0.42) which we want to track.
```bash
envsubst < ~/environment/eks-demos/src/php-echo/Dockerfile.template > ~/environment/eks-demos/src/php-echo/Dockerfile
```

Each Cloud9 instance has the Docker daemon installed. Build the Docker image from the Cloud9 terminal then run the newly containerized app.
```bash
docker build -t ${EKS_APP_NAME} ~/environment/eks-demos/src/${EKS_APP_NAME}/
docker images                                                       # see what you produced
docker ps                                                           # nothing running ...
container_id=$(docker run --detach --rm -p 8081:80 ${EKS_APP_NAME}) # request docker to instantiate a single container as a background process
docker ps                                                           # ... now one container running
docker exec -it ${container_id} curl localhost:80                   # shell INTO that container and curl the INTERNAL port (80)
curl localhost:8081                                                 # show that the corresponding EXTERNAL port is mapped to a high-order port (8081) on the c9 instance
docker network inspect bridge | jq  .[0].IPAM.Config[0].Subnet      # see why the ec2IP is no longer equivalent to the localhostIP
docker stop ${container_id}                                         # stop the container (which will be terminated because we ran it with the --rm flag)
```

[Return To Main Menu](/README.md)
