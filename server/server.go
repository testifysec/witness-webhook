package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/in-toto/go-witness/signer/kms"
	"github.com/testifysec/witness-webhook/config"
	"github.com/testifysec/witness-webhook/webhook"
	"github.com/testifysec/witness-webhook/webhook/github"
)

type Server struct {
	config config.Config
	r      *mux.Router
}

func New(config config.Config) (Server, error) {
	s := Server{
		config: config,
		r:      mux.NewRouter(),
	}

	kmsSignerProvider := kms.New(kms.WithRef("hashivault://testkey"), kms.WithHash("sha256"))
	signer, err := kmsSignerProvider.Signer(context.Background())
	if err != nil {
		log.Fatal("could not create signer\n", err)
	}

	githubHandler, err := github.New(
		github.WithSecretFile(os.Getenv("GITHUB_SECRET_FILE")),
		github.WithSigner(signer),
	)

	if err != nil {
		log.Fatalf("could not create webhook handler: %v\n", err)
	}

	httpHandler, err := createHttpHandler("github", githubHandler)
	if err != nil {
		log.Fatalf("could not create http handler: %v\n", err)
	}

	s.r.HandleFunc("/github", httpHandler).Methods(http.MethodPost)
	return s, nil
}

func (s Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.r.ServeHTTP(w, req)
}

func createHttpHandler(name string, h webhook.Handler) (http.HandlerFunc, error) {
	if h == nil {
		return nil, fmt.Errorf("webhook handler is required")
	}

	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("request for webhook %v received\n", name)
		env, err := h.HandleRequest(req)
		if err != nil {
			log.Printf("could not handle request for webhook %v: %v\n", name, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("%+v\n", env)
	}, nil
}
