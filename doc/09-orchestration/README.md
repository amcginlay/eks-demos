# Orchestration

[Kubernetes](https://en.wikipedia.org/wiki/Kubernetes) can be described as an orchestration platform. To understand what that means we first need to understand some of the limiations of using a standalone containerization platform, such as [Docker](https://en.wikipedia.org/wiki/Docker_(software)).

[Spring Boot](https://en.wikipedia.org/wiki/Spring_Framework#Spring_Boot) is a popular Java framework for rapid application development which does much of the heavy lifting associated with modern software. So much so that it is possible to illustrate a key the benefit of orchestration without writing any code whatsoever. We are going to build a application which can be terminated by invoking its `/shutdown` endpoint.

Our first step is to use [Spring Initializr](https://start.spring.io/) - a web application which automates the construction of ready-to-roll Spring Boot application code. This code is downloaded as a zipfile and gives the developer a well defined starting position. Whilst it is common to use Spring Initializr from a web browser, it is equally easy to invoke it from the command line.

The required dependencies, which combine to provide the features needed, are **Web** and **Actuator**. Check the documentation to learn more about the [Actuator](https://docs.spring.io/spring-boot/docs/current/reference/htmlsingle/#production-ready) dependency.

From the Cloud9 terminal, invoke Spring Initializr to construct a [Gradle](https://en.wikipedia.org/wiki/Gradle) project with the required dependencies, then unzip the results.
```bash
mkdir ~/environment/eks-demos/src/orchestration
curl https://start.spring.io/starter.zip \
  -d type=gradle-project \
  -d language=java \
  -d platformVersion=2.4.4.RELEASE \
  -d packaging=jar \
  -d jvmVersion=11 \
  -d groupId=com.eks \
  -d artifactId=orchestration \
  -d name=orchestration \
  -d packageName=com.eks.orchestration \
  -d dependencies=web,actuator \
  -o ~/environment/eks-demos/src/orchestration/orchestration.zip
unzip ~/environment/eks-demos/src/orchestration/orchestration.zip -d ~/environment/eks-demos/src/orchestration/
``` 

For security reasons the Actuator's `/shutdown` endpoint is disabled by default. Re-enable it by updating the `application.properties` configuration file as follows.
```bash
cat > ~/environment/eks-demos/src/orchestration/src/main/resources/application.properties << EOF
management.endpoints.web.exposure.include=*
management.endpoint.shutdown.enabled=true
endpoints.shutdown.enabled=true
EOF
```

When we build and run the application the Cloud9 terminal will begin tailing [stdout](https://en.wikipedia.org/wiki/Standard_streams#Standard_output_(stdout)) and will not return a prompt. Running the application will take a couple of minutes on the first attempt. Look for a response like "**Started OrchestrationApplication**" to know that your app is running.
```bash
cd ~/environment/eks-demos/src/orchestration/
./gradlew build && java -jar ./build/libs/orchestration-0.0.1-SNAPSHOT.jar
```

From **another terminal session**, invoke the `/shutdown` endpoint
```bash
curl -X POST http://localhost:8080/actuator/shutdown
```

You will see the response "Shutting down, bye...". The application has now terminated and the prompt in the first terminal window is back. The `/shutdown` request was successful.

We need to containerize this app, so create a Dockerfile then build our container and run it.
```bash
cat > ~/environment/eks-demos/src/orchestration/Dockerfile << EOF
FROM adoptopenjdk/openjdk11:alpine-jre
ARG JAR_FILE=target/*.jar
COPY \${JAR_FILE} app.jar
ENTRYPOINT ["java","-jar","/app.jar"]
EOF

docker build --build-arg JAR_FILE=build/libs/\*.jar -t orchestration ~/environment/eks-demos/src/orchestration/
docker run --detach --rm -p 8081:8080 orchestration
```

Confirm the container instance is running, hit the `/shutdown` endpoint, then check if it was terminated.
```bash
docker ps                                     # container running?
curl -X POST localhost:8081/actuator/shutdown # send shutdown
docker ps                                     # container dead?
```

... TODO

[Return To Main Menu](/README.md)
