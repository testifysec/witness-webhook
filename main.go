// Copyright 2024 Witness Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

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
	enableTLSEnvVar   = "WITNESS_WEBHOOK_ENABLE_TLS"
	tlsSkipVerifyVar  = "WITNESS_WEBHOOK_TLS_SKIP_VERIFY"
	tlsCertEnvVar     = "WITNESS_WEBHOOK_TLS_CERT"
	tlsKeyEnvVar      = "WITNESS_WEBHOOK_TLS_KEY"
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

	s, err := server.New(context.Background(), config)
	if err != nil {
		log.Fatalf("failed to start webhook server: %v\n", err)
	}

	listenAddr := os.Getenv(listenAddrEnvVar)
	if len(listenAddr) == 0 {
		listenAddr = defaultListenAddr
	}

	r := mux.NewRouter()
	r.PathPrefix("/debug").Handler(http.DefaultServeMux)
	r.PathPrefix("/webhook").Handler(http.StripPrefix("/webhook", s))
	r.Path("/ready").HandlerFunc(readyHandler)

	srv := &http.Server{
		Addr:    listenAddr,
		Handler: r,
	}

	tlsEnabled := strings.TrimSpace(strings.ToLower(os.Getenv(enableTLSEnvVar))) == "true"
	tlsSkipVerify := strings.TrimSpace(strings.ToLower(os.Getenv(tlsSkipVerifyVar))) == "true"
	if tlsEnabled && tlsSkipVerify {
		log.Println("enabling InsecureSkipVerify for TLS. DO NOT ENABLE IN PRODUCTION")
		srv.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	go func() {
		if tlsEnabled {
			log.Printf("listening with TLS on %v\n", listenAddr)
			if err := srv.ListenAndServeTLS(os.Getenv(tlsCertEnvVar), os.Getenv(tlsKeyEnvVar)); err != nil {
				log.Println(err)
			}
		} else {
			log.Printf("listening on %v\n", listenAddr)
			if err := srv.ListenAndServe(); err != nil {
				log.Println(err)
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	log.Println("caught interrupt, waiting for requests to finish...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("error shutting down server: %v", err)
	}

	log.Println("shutting down")
}

// for now this just writes 200 back to show the server is up and listening
func readyHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}
