package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/testifysec/witness-webhook/config"
	"github.com/testifysec/witness-webhook/server"

	_ "net/http/pprof"

	_ "github.com/in-toto/go-witness"
)

const (
	defaultConfigPath = "/webhook-config.yaml"
	configPathEnvVar  = "WITNESS_WEBHOOK_CONFIG_PATH"
	defaultListenAddr = ":8085"
	listenAddrEnvVar  = "WITNESS_WEBHOOK_LISTEN_ADDR"
)

func main() {
	configPath := os.Getenv(configPathEnvVar)
	if len(configPath) == 0 {
		configPath = defaultConfigPath
	}

	config, err := config.New(configPath)
	if err != nil {
		log.Fatalf("could not load config: %v\n", err)
	}

	log.Println(config)

	s, err := server.New(context.Background(), config)
	if err != nil {
		log.Fatalf("filed to start webhook server: %v\n", err)
	}

	listenAddr := os.Getenv(listenAddrEnvVar)
	if len(listenAddr) == 0 {
		listenAddr = defaultListenAddr
	}

	r := mux.NewRouter()
	r.PathPrefix("/debug/").Handler(http.DefaultServeMux)
	r.PathPrefix("/").Handler(s)
	log.Println("listening...")
	log.Fatal(http.ListenAndServe(listenAddr, r))
}
