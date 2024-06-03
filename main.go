package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/testifysec/witness-webhook/config"
	"github.com/testifysec/witness-webhook/server"

	_ "net/http/pprof"

	_ "github.com/in-toto/go-witness"
)

const (
	defaultConfigPath = "/webhook-config.yaml"
	configPathEnvVar  = "WITNESS_WEBHOOK_CONFIG_PATH"
)

func main() {
	/*
		configPath := os.Getenv(configPathEnvVar)
		if len(configPath) == 0 {
			configPath = defaultConfigPath
		}

		config, err := loadConfig(configPath)
		if err != nil {
			log.Fatalf("could not load config: %w", err)
		}
	*/
	r := mux.NewRouter()
	r.PathPrefix("/debug/").Handler(http.DefaultServeMux)
	s, err := server.New(config.Config{})
	if err != nil {
		log.Fatalf("filed to start webhook server: %v\n", err)
	}

	r.PathPrefix("/").Handler(s)
	log.Println("listening...")
	log.Fatal(http.ListenAndServe("0.0.0.0:3000", r))
}
