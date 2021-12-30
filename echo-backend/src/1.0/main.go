package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
)

const version = "1.0"

func getEnv(key string, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        resMap := map[string]string{"version": version}
        resJson, _ := json.Marshal(resMap)
        resp := string(resJson)        
        log.Printf("%s", resp)
        fmt.Fprintf(w, "%s\n", resp)
    })
    port := getEnv("PORT", "8080")
    log.Printf("Server listening on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, mux))
}
