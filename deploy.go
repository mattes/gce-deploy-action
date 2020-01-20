package main

import (
	"net/http"

	computeBeta "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/compute/v1"
)

func Run(githubActionConfig *GithubActionConfig, config *Config, deploy Deploy) {

	Infof("%v: Starting deploy", deploy.Name)

	var googleClient *http.Client
	if deploy.googleApplicationCredentialsData != "" {
		client, err := NewClientFromJSON(deploy.googleApplicationCredentialsData)
		if err != nil {
			Fatalf("Invalid deploys.*.google_application_credentials: %v", err)
		}
		googleClient = client

	} else {
		client, err := NewClientFromJSON(githubActionConfig.googleApplicationCredentialsData)
		if err != nil {
			Fatalf("Invalid github_action.google_application_credentials: %v", err)
		}
		googleClient = client
	}

	computeService, err := compute.New(googleClient)
	if err != nil {
		Fatalf("%v", err)
	}

	computeBetaService, err := computeBeta.New(googleClient)
	if err != nil {
		Fatalf("%v", err)
	}

	_ = computeService
	_ = computeBetaService

	Infof("%v: Finished deploy", deploy.Name)
}
