FROM adoptopenjdk/openjdk11:alpine-jre
COPY build/libs/*.jar app.jar
ENTRYPOINT ["java","-jar","/app.jar"]
