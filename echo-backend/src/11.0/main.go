package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

const version = "11.0"

var failureRate = getEnvInt("FAILURE_RATE", 0)

func getEnvInt(key string, fallback int) int {
	result, _ := strconv.Atoi(getEnv(key, strconv.Itoa(fallback)))
	return result
}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		rnd := rand.Intn(100) + 1
		if rnd <= failureRate {
			http.Error(w, "Simulated Error", http.StatusInternalServerError)
			return
		}

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
