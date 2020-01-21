package main

import (
	"net/http"

	computeBeta "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/compute/v1"
)

func Run(githubActionConfig *GithubActionConfig, config *Config, deploy Deploy) {

	Infof("%v: Starting deploy", deploy.Name)

	// create google client with application credentials from deploy config or
	// github action config
	var googleClient *http.Client
	if deploy.googleApplicationCredentialsData != "" {
		client, f, err := NewClientFromJSON(deploy.googleApplicationCredentialsData)
		if err != nil {
			Fatalf("Invalid deploys.*.google_application_credentials: %v", err)
		}
		googleClient = client

		if deploy.Project == "" {
			deploy.Project = f.ProjectID
		}

	} else {
		client, f, err := NewClientFromJSON(githubActionConfig.googleApplicationCredentialsData)
		if err != nil {
			Fatalf("Invalid github_action.google_application_credentials: %v", err)
		}
		googleClient = client

		if deploy.Project == "" {
			deploy.Project = f.ProjectID
		}
	}

	// create compute service client
	computeService, err := compute.New(googleClient)
	if err != nil {
		Fatalf("%v", err)
	}

	// create compute beta service client
	computeBetaService, err := computeBeta.New(googleClient)
	if err != nil {
		Fatalf("%v", err)
	}

	// clone instance template and update instance group
	instanceTemplateURL, err := CloneInstanceTemplate(computeService, deploy)
	if err != nil {
		LogError(err.Error(), map[string]string{"name": deploy.Name})
		return
	}

	Infof("%v: Created new instance template: %v", deploy.Name, instanceTemplateURL)

	// start rolling update via instance group manager
	if err := StartRollingUpdate(computeBetaService, deploy, instanceTemplateURL); err != nil {
		LogError(err.Error(), map[string]string{"name": deploy.Name})
		return
	}

	Infof("%v: Started rolling update with new instance template", deploy.Name)

	if config.DeleteInstanceTemplatesAfter > 0 {
		if err := CleanupInstanceTemplates(computeService, deploy.Project, config.DeleteInstanceTemplatesAfter); err != nil {
			LogWarning(err.Error(), map[string]string{"project": deploy.Project})
		}
	}
}
