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

var backend = getEnv("BACKEND", "none")

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
	if backend != "none" {
		backendJson := shellExec("curl", "--silent", backend)
		var backendMap map[string]interface{}
		json.Unmarshal([]byte(backendJson), &backendMap)
		backend = fmt.Sprint(backendMap["version"])
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
