# Upgrade your deployment

Rapid, iterative code changes are commonplace in cloud native software deployments and Kubernetes copes well with these demands.
You will now make a small change to your code and redeploy your app using `kubectl`.

Version 2.0 of your app provides support for the use of a **backend** app which will be introduced in a later chapter.

Run the following snippet in the terminal to create the new source code for your app.
```bash
mkdir -p ~/environment/echo-frontend/src/2.0/
cat > ~/environment/echo-frontend/src/2.0/main.go << EOF
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

Open `~/environment/echo-frontend/src/2.0/main.go` in Cloud9 IDE to review the updated code.

Copy and re-use the version 1.0 Dockerfile.
```bash
cp ~/environment/echo-frontend/src/1.0/Dockerfile ~/environment/echo-frontend/src/2.0/
```

Use Docker to build and run your new container image.
```bash
docker build -t echo-frontend:2.0 ~/environment/echo-frontend/src/2.0/
container_id=$(docker run --detach --rm -p 8082:80 echo-frontend:2.0)
```

Give it a quick test then stop the running container.
```bash
curl localhost:8082
docker stop ${container_id}
```

Observe the new `backend` attribute ("n/a" by default) and the value for the `version` attribute which is set to 2.0.

Tag and push the Docker image to the ECR repository.
```bash
docker tag echo-frontend:2.0 ${EKS_ECR_REGISTRY}/echo-frontend:2.0
aws ecr get-login-password | docker login --username AWS --password-stdin ${EKS_ECR_REGISTRY}
docker push ${EKS_ECR_REGISTRY}/echo-frontend:2.0
```

Review the version 1.0 and version 2.0 images, now side by side in ECR.
```bash
aws ecr list-images --repository-name echo-frontend
```

Re-apply the deployment manifest, adjusting only for the new version, to update your app **in-place**.
```bash
cat ~/environment/echo-frontend/templates/echo-frontend-deployment.yaml | \
    sed "s/{{ .Values.registry }}/${EKS_ECR_REGISTRY}/g" | \
    sed "s/{{ .Values.color }}/blue/g" | \
    sed "s/{{ .Values.version }}/2.0/g" | \
    kubectl apply -f -
```

Inspect your updated deployment.
Observe the version change from 1.0 to 2.0 under the "IMAGES" heading.
```bash
sleep 10 && kubectl -n demos get deployments,pods -o wide
```

Exec into the first pod to perform curl test.
Satisfy yourself that your app has been upgraded.
```bash
first_pod=$(kubectl -n demos get pods -l app=echo-frontend-blue -o name | head -1)
kubectl -n demos exec -it ${first_pod} -- curl localhost:80
```

For now, roll back your deployment to version 1.0.
```bash
cat ~/environment/echo-frontend/templates/echo-frontend-deployment.yaml | \
    sed "s/{{ .Values.registry }}/${EKS_ECR_REGISTRY}/g" | \
    sed "s/{{ .Values.color }}/blue/g" | \
    sed "s/{{ .Values.version }}/1.0/g" | \
    kubectl apply -f -
```

The version 2.0 image remains in ECR for later use.

[Return To Main Menu](/README.md)