package main

import (
    "log"
    "os"
    "os/exec"
    "fmt"
    "time"
    "net/http"
    "strings"
)

const version = "1.0.42"

func shellExec(prg string, args ...string) string {
    cmd := exec.Command(prg, args...)
    stdout, err := cmd.Output()

    if err != nil {
        panic(err)
    }

    return string(stdout)
}

func getResponse() string {
    time := time.Now().Format("15:04:05")
    instanceId := shellExec("curl", "--silent", "http://169.254.169.254/latest/meta-data/instance-id")
    ec2Ip := shellExec("curl", "--silent", "http://169.254.169.254/latest/meta-data/local-ipv4")
    localHostIp := strings.TrimSuffix(shellExec("hostname", "-i"), "\n")
    res := fmt.Sprintf(`{ "1_time": "%s", "2_version": "%s", "ec2Instance": "%s", "ec2IP": "%s", "localhostIP": "%s", "backend": "n/a" }`, time, version, instanceId, ec2Ip, localHostIp)
    return res
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        resp := getResponse()
        log.Printf("%s", resp)
        fmt.Fprintf(w, "%s\n", resp)
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Server listening on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, mux))
}
