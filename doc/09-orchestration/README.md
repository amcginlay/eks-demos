# Orchestration

[Kubernetes](https://en.wikipedia.org/wiki/Kubernetes) can be described as an orchestration platform. To understand what that means we first need to understand some of the limitations of using a standalone containerization platform such as [Docker](https://en.wikipedia.org/wiki/Docker_(software)).

[Spring Boot](https://en.wikipedia.org/wiki/Spring_Framework#Spring_Boot) is a popular Java framework for rapid application development which does much of the heavy lifting associated with modern software. So much so that it is possible to illustrate a key the benefit of orchestration without writing any code whatsoever. We are going to build a application which can be terminated by invoking its `/shutdown` endpoint.

Our first step is to use [Spring Initializr](https://start.spring.io/) - a web application which automates the construction of ready-to-roll Spring Boot application code. This code is downloaded as a zipfile and gives the developer a well defined starting position. Whilst it is common to use Spring Initializr from a web browser, it is equally easy to invoke it from the command line.

The required dependencies, which combine to provide the features needed, are **Web** and **Actuator**. Check the documentation to learn more about the [Actuator](https://docs.spring.io/spring-boot/docs/current/reference/htmlsingle/#production-ready) dependency.

If you followed the previous steps a directory for the source code will have already been cloned onto your Cloud9 instance. Confirm that the `~/environment/eks-demos/src/boot-orch` directory exists and contains a single file named `Dockerfile`. From the Cloud9 terminal, invoke Spring Initializr to construct a [Gradle](https://en.wikipedia.org/wiki/Gradle) project with the required dependencies, then unzip the results which will appear alongside the `Dockerfile`.
```bash
curl https://start.spring.io/starter.zip \
  -d type=gradle-project \
  -d language=java \
  -d platformVersion=2.4.4.RELEASE \
  -d packaging=jar \
  -d jvmVersion=11 \
  -d groupId=com.eks \
  -d artifactId=boot-orch \
  -d name=boot-orch \
  -d packageName=com.eks.boot-orch \
  -d dependencies=web,actuator \
  -o ~/environment/eks-demos/src/boot-orch/app.zip
unzip ~/environment/eks-demos/src/boot-orch/app.zip -d ~/environment/eks-demos/src/boot-orch/
```

For security reasons the Actuator's `/shutdown` endpoint is disabled by default. Re-enable it by updating the `application.properties` configuration file as follows.
```bash
cat > ~/environment/eks-demos/src/boot-orch/src/main/resources/application.properties << EOF
management.endpoints.web.exposure.include=*
management.endpoint.shutdown.enabled=true
endpoints.shutdown.enabled=true
EOF
```

When we build and run the application, the Cloud9 terminal will begin tailing [stdout](https://en.wikipedia.org/wiki/Standard_streams#Standard_output_(stdout)) and will not return a prompt. Running the application will take a couple of minutes on the first attempt. Look for a response like "**Started OrchestrationApplication**" to know that your app is running.
```bash
cd ~/environment/eks-demos/src/boot-orch/
./gradlew build && java -jar ./build/libs/boot-orch-0.0.1-SNAPSHOT.jar
```

From **another terminal session**, invoke the `/shutdown` endpoint
```bash
curl -X POST http://localhost:8080/actuator/shutdown
```

You will see the response "Shutting down, bye...". The application has now terminated and the prompt in the first terminal window is back. The `/shutdown` request was successful.

With the Dockerfile in place we can containerize and run the app as follows.
```bash
docker build -t boot-orch ~/environment/eks-demos/src/boot-orch/
docker run --detach --rm -p 8081:8080 boot-orch
```

This time the app is running detached inside Docker so the command prompt remains available. Confirm the container instance is running, then hit the `/shutdown` endpoint and confirm it was terminated. This proves that Docker alone cannot protect us from process termination.
```bash
docker ps                                     # container running?
curl -X POST localhost:8081/actuator/shutdown
docker ps                                     # container dead?
```

Create target ECR repo, deleting it first if needed, then push the Docker image to ECR repository
```bash
aws ecr delete-repository --repository-name boot-orch --force >/dev/null 2>&1
boot_orch_repo=$(aws ecr create-repository \
  --repository-name boot-orch \
  --region ${AWS_DEFAULT_REGION} \
  --image-scanning-configuration scanOnPush=true \
  --query 'repository.repositoryUri' \
  --output text \
)
aws ecr get-login-password --region ${AWS_DEFAULT_REGION} | docker login --username AWS --password-stdin ${boot_orch_repo}
docker tag boot-orch:latest ${boot_orch_repo}:1.0.0
docker images
docker push ${boot_orch_repo}:1.0.0
```

The EKS cluster can now locate this image by its tag

[Return To Main Menu](/README.md)
