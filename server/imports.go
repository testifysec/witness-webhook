package server

// These are imported so their init functions run and registry them with their factories
import (
	// webhook handlers
	_ "github.com/testifysec/witness-webhook/webhook/github"
)
