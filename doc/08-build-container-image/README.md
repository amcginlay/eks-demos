# Build A Container Image

The target for our first image is a simple web app written in [Go](https://go.dev/).
Go compiles to standalone binaries which are well suited to producing smaller container images.

Run the following snippet in the terminal to create the source code for your app.
```bash
mkdir -p ~/environment/echo-frontend-1.0/
cat > ~/environment/echo-frontend-1.0/main.go << EOF
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "math"
    "net/http"
    "os"
    "os/exec"
    "strings"
    "time"
)

const version = "1.0"

var x = 0.0
func doWork() {
    x = 0.0001
    for i := 0; i <= 1000000; i++ {
        x += math.Sqrt(x)
    }
}

func getEnv(key string, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}

func shellExec(prg string, args ...string) string {
    cmd := exec.Command(prg, args...)
    stdout, _ := cmd.Output()
    return string(stdout)
}

func getResponse() string {
    doWork()
    ec2Ip := shellExec("curl", "--silent", "http://169.254.169.254/latest/meta-data/local-ipv4")
    hostname := strings.TrimSuffix(shellExec("hostname"), "\n")
    time := time.Now().Format("15:04:05")
    resMap := map[string]string{"ec2IP": ec2Ip, "hostname": hostname, "time": time, "version": version}
    resJson, _ := json.Marshal(resMap)
    return string(resJson)
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        resp := getResponse()
        log.Printf("%s", resp)
        fmt.Fprintf(w, "%s\n", resp)
    })
    port := getEnv("PORT", "8080")
    log.Printf("Server listening on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, mux))
}
EOF
```

Open `~/environment/echo-frontend-1.0/main.go` in Cloud9 IDE to review the code.

You can launch your app from the terminal session using the following.
```bash
go run ~/environment/echo-frontend-1.0/main.go
```

Your webserver app will tie up this first Cloud9 terminal session until its process is stopped.
Leave the webserver running and select `Window -> New Terminal` to make a second terminal session available.

In the second terminal session, use the curl command to send an HTTP GET request to the webserver as follows.
```bash
curl http://localhost:8080
```

As you do so, observe that the recorded value of **hostname** is synonymous with the value of **ec2IP** in this execution context.

NOTE the use of [Instance Metadata](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) **(169.254.169.254)** within the source code is an indication that your app is tailor-made for deployment on EC2 instances.

Return to the first terminal session and use `ctrl+c` to quit the app and recover your command prompt.

Each Cloud9 instance has the [Docker](https://en.wikipedia.org/wiki/Docker_(software)) daemon installed with a set of images pre-loaded. Remove them as they are not required.
```bash
for i in $(docker ps -q); do docker kill $i; done
docker system prune --all --force
```

Run the following snippet in the terminal to create the [`Dockerfile`](https://docs.docker.com/engine/reference/builder/) for your app.
```bash
cat > ~/environment/echo-frontend-1.0/Dockerfile << EOF
FROM golang:1.12.0-alpine3.9
ENV PORT=80
WORKDIR /app/
COPY main.go .
RUN go build -o main && \\
    apk add curl bind-tools
ENTRYPOINT [ "./main" ]
EOF
```

Open `~/environment/echo-frontend-1.0/Dockerfile` in Cloud9 IDE to review the code.

Each Cloud9 instance has the Docker daemon installed. Build the Docker image from the Cloud9 terminal then run the newly containerized app.
```bash
docker build -t echo-frontend:1.0 ~/environment/echo-frontend-1.0/    # build the container image
docker images                                                         # see what you produced
docker ps                                                             # nothing running ...
container_id=$(docker run --detach --rm -p 8081:80 echo-frontend:1.0) # ask docker to instantiate a single container as a background process
docker ps                                                             # ... now one container running
```

Invoke the webserver from inside the container.
```bash
docker exec -it ${container_id} curl localhost:80
```

Invoke the webserver from outside the container.
```bash
curl localhost:8081
```

The response for the two previous `curl` requests are identical because it is the same operation, only the perspective is different.
Observe that the recorded values of **hostname** and **ec2IP** have now diverged.
This is because your app is now containerized and running inside its own namespace.

We are done with running images in Docker for now so stop the container (which will be terminated because we ran it with the `--rm` flag).
```bash
docker stop ${container_id}
```

Before we move on, instruct Docker to build version 2.0 of your app so we've got something extra to play with later on.
Version 2.0 is extended to support the use of a **backend** app.
This feature is disabled by default.

Run the following snippet in the terminal to create the version 2.0 source code for your app.
```bash
mkdir -p ~/environment/echo-frontend-2.0/
cat > ~/environment/echo-frontend-2.0/main.go << EOF
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "math"
    "net/http"
    "os"
    "os/exec"
    "strings"
    "time"
)

const version = "2.0"

var x = 0.0
func doWork() {
    x = 0.0001
    for i := 0; i <= 1000000; i++ {
        x += math.Sqrt(x)
    }
}

func getEnv(key string, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}

func shellExec(prg string, args ...string) string {
    cmd := exec.Command(prg, args...)
    stdout, _ := cmd.Output()
    return string(stdout)
}

func getResponse() string {
    doWork()
    backend := getEnv("BACKEND", "n/a")
    if backend != "n/a" {
        backend = shellExec("curl", "--silent", backend)
    }
    ec2Ip := shellExec("curl", "--silent", "http://169.254.169.254/latest/meta-data/local-ipv4")
    hostname := strings.TrimSuffix(shellExec("hostname"), "\n")
    time := time.Now().Format("15:04:05")
    resMap := map[string]string{"backend": backend, "ec2IP": ec2Ip, "hostname": hostname, "time": time, "version": version}
    resJson, _ := json.Marshal(resMap)
    return string(resJson)
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        resp := getResponse()
        log.Printf("%s", resp)
        fmt.Fprintf(w, "%s\n", resp)
    })
    port := getEnv("PORT", "8080")
    log.Printf("Server listening on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, mux))
}
EOF
```

Copy and re-use the version 1.0 Dockerfile.
```bash
cp ~/environment/echo-frontend-1.0/Dockerfile ~/environment/echo-frontend-2.0/
```

Build version 2.0 and give it a quick test.
```bash
docker build -t echo-frontend:2.0 ~/environment/echo-frontend-2.0/
container_id=$(docker run --detach --rm -p 8082:80 echo-frontend:2.0)
curl localhost:8082
docker stop ${container_id}
```

[Return To Main Menu](/README.md)
