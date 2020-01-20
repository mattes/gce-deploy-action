package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	computeBeta "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

func NewClientFromJSON(data string) (*http.Client, error) {
	conf, err := google.JWTConfigFromJSON([]byte(data), compute.ComputeScope)
	if err != nil {
		return nil, err
	}

	return conf.Client(oauth2.NoContext), nil
}

func CloneInstanceTemplate(c *compute.Service, projectId, oldName, newName string, cloudconfig, shutdownScript []byte) (string, error) {
	s := compute.NewInstanceTemplatesService(c)

	// get template with existing name
	oldTemplate, err := s.Get(projectId, oldName).Do()
	if err != nil {
		return "", err
	}

	// remove existing user-data from old template if any
	metadataItems := make([]*compute.MetadataItems, 0)
	if oldTemplate.Properties != nil && oldTemplate.Properties.Metadata != nil {
		for _, i := range oldTemplate.Properties.Metadata.Items {
			if i.Key != "user-data" && i.Key != "shutdown-script" {
				metadataItems = append(metadataItems, i)
			}
		}
	}

	// get the new user data ready and add to metadata
	userDataItem := &compute.MetadataItems{
		Key:   "user-data",
		Value: stringPtr(string(cloudconfig)),
	}
	metadataItems = append(metadataItems, userDataItem)

	// create shutdown script and add to metadata
	// The main process inside the container will receive SIGTERM, and after a grace period, SIGKILL.
	// https://cloud.google.com/compute/docs/shutdownscript#limitations
	// On-demand instances: 90 seconds after you stop or delete an instance
	// Preemptible instances: 30 seconds after instance preemption begins
	shutdownScriptItem := &compute.MetadataItems{
		Key:   "shutdown-script",
		Value: stringPtr(string(shutdownScript)),
	}
	metadataItems = append(metadataItems, shutdownScriptItem)

	newTemplate := oldTemplate
	newTemplate.Name = newName
	newTemplate.Properties.Metadata.Items = metadataItems

	op, err := s.Insert(projectId, newTemplate).Do()
	if err != nil {
		return "", err
	}

	// wait until ready
	for {
		_, err := s.Get(projectId, newName).Do()
		if err != nil && isNotReadyErr(err) {
			time.Sleep(2 * time.Second)
			continue
		} else if err != nil {
			return "", err
		} else if err == nil {
			return op.TargetLink, nil
		}
	}
}

// https://cloud.google.com/compute/docs/instance-groups/rolling-out-updates-to-managed-instance-groups#starting_a_basic_rolling_update
func StartRollingUpdate(c *computeBeta.Service, projectId, region, instanceGroup, instanceTemplateName, instanceTemplateURL string) error {
	s := computeBeta.NewRegionInstanceGroupManagersService(c)

	err := retryIfResourceNotReady(func() error {
		ig, err := s.Get(projectId, region, instanceGroup).Do()
		if err != nil {
			return err
		}

		ig.InstanceTemplate = ""
		ig.Versions = []*computeBeta.InstanceGroupManagerVersion{
			{
				InstanceTemplate: instanceTemplateURL,
				Name:             instanceTemplateName,
			},
		}

		ig.UpdatePolicy = &computeBeta.InstanceGroupManagerUpdatePolicy{
			InstanceRedistributionType: "PROACTIVE",
			MaxSurge:                   &computeBeta.FixedOrPercent{Fixed: 3},
			MaxUnavailable:             &computeBeta.FixedOrPercent{Fixed: 0},
			MinimalAction:              "REPLACE",
			Type:                       "PROACTIVE",
			MinReadySec:                10,
		}

		_, err = s.Update(projectId, region, instanceGroup, ig).Do()
		return err
	})
	if err != nil {
		return fmt.Errorf("StartRollingUpdate: %v", err)
	}

	return nil
}

func CleanupInstanceTemplates(c *compute.Service, projectId string) error {
	s := compute.NewInstanceTemplatesService(c)

	l, err := s.List(projectId).Do()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, item := range l.Items {
		t, err := time.Parse(time.RFC3339, item.CreationTimestamp)
		if err != nil {
			return err
		}

		if time.Now().UTC().Add(-30 * 24 * time.Hour).After(t) {
			wg.Add(1)
			go func(instanceTemplate string) {

				// TODO delete docker image, too

				s.Delete(projectId, instanceTemplate).Do()
				wg.Done()
			}(item.Name)
		}
	}

	wg.Wait()
	return nil
}

func isReasonErr(err error, reason string) bool {
	if e, ok := err.(*googleapi.Error); ok {
		for _, x := range e.Errors {
			if x.Reason == reason {
				return true
			}
		}
	}
	return false
}

func isAlreadyExistErr(err error) bool {
	return isReasonErr(err, "alreadyExists")
}

func isNotReadyErr(err error) bool {
	return isReasonErr(err, "resourceNotReady")
}

func stringPtr(in string) *string {
	return &in
}

func retryIfResourceNotReady(fn func() error) error {
	for i := 0; i < 20; i++ {

		err := fn()
		if err == nil {
			return nil
		}

		if isNotReadyErr(err) {
			time.Sleep(2 * time.Second)
			continue
		}

		return err
	}

	return fmt.Errorf("too many retries")
}
