package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	computeBeta "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

var (
	instanceTemplateDescription = "created by gce-deploy-action"
)

type ServiceAccountFile struct {
	Type string `json:"type"` // serviceAccountKey or userCredentialsKey

	ClientEmail  string `json:"client_email"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	TokenURL     string `json:"token_uri"`
	ProjectID    string `json:"project_id"`

	ClientSecret string `json:"client_secret"`
	ClientID     string `json:"client_id"`
	RefreshToken string `json:"refresh_token"`
}

func NewClientFromJSON(data string) (*http.Client, *ServiceAccountFile, error) {
	conf, err := google.JWTConfigFromJSON([]byte(data), compute.ComputeScope)
	if err != nil {
		return nil, nil, err
	}

	f := &ServiceAccountFile{}
	if err := json.Unmarshal([]byte(data), f); err != nil {
		return nil, nil, err
	}

	return conf.Client(oauth2.NoContext), f, nil
}

func CloneInstanceTemplate(c *compute.Service, d Deploy) (string, error) {
	s := compute.NewInstanceTemplatesService(c)

	// get base instance template
	instanceTemplateBase, err := s.Get(d.Project, d.InstanceTemplateBase).Do()
	if err != nil {
		return "", fmt.Errorf("get instance template base '%v/%v': %v", d.Project, d.InstanceTemplateBase, err)
	}

	// initialize new instance template
	instanceTemplate := instanceTemplateBase
	instanceTemplate.Name = d.InstanceTemplate
	instanceTemplate.Description = instanceTemplateDescription

	if instanceTemplate.Properties == nil {
		instanceTemplate.Properties = &compute.InstanceProperties{}
	}

	// add new tags
	if instanceTemplate.Properties.Tags == nil {
		instanceTemplate.Properties.Tags = &compute.Tags{}
		instanceTemplate.Properties.Tags.Items = make([]string, 0)
	}
	for _, v := range d.Tags {
		instanceTemplate.Properties.Tags.Items = append(instanceTemplate.Properties.Tags.Items, v)
	}

	// add new labels
	if instanceTemplate.Properties.Labels == nil {
		instanceTemplate.Properties.Labels = make(map[string]string)
	}
	for k, v := range d.Labels {
		instanceTemplate.Properties.Labels[k] = v
	}

	// add new metadata keys
	if instanceTemplate.Properties.Metadata == nil {
		instanceTemplate.Properties.Metadata = &compute.Metadata{}
		instanceTemplate.Properties.Metadata.Items = make([]*compute.MetadataItems, 0)
	}
	for k, v := range d.Metadata {
		instanceTemplate.Properties.Metadata.Items = append(instanceTemplate.Properties.Metadata.Items,
			newMetadataItem(k, v))
	}

	// startup script
	if d.StartupScriptPath != "" {
		instanceTemplate.Properties.Metadata.Items = append(instanceTemplate.Properties.Metadata.Items,
			newMetadataItem("startup-script", d.startupScript))
	}

	// shutdown script
	if d.ShutdownScriptPath != "" {
		instanceTemplate.Properties.Metadata.Items = append(instanceTemplate.Properties.Metadata.Items,
			newMetadataItem("shutdown-script", d.shutdownScript))
	}

	// cloud init
	if d.CloudInitPath != "" {
		instanceTemplate.Properties.Metadata.Items = append(instanceTemplate.Properties.Metadata.Items,
			newMetadataItem("user-data", d.cloudInit))
	}

	op, err := s.Insert(d.Project, instanceTemplate).Do()
	if err != nil {
		return "", fmt.Errorf("save instance template: %v", err)
	}

	// wait until ready
	retry := 0
	for {
		_, err := s.Get(d.Project, d.InstanceTemplate).Do()
		if err != nil && isNotReadyErr(err) {
			time.Sleep(2 * time.Second)
			retry++
			if retry > 10 {
				return "", fmt.Errorf("get saved instance template: too many retries")
			}
			continue

		} else if err != nil {
			return "", fmt.Errorf("get saved instance template: %v", err)
		} else if err == nil {
			return op.TargetLink, nil
		}
	}
}

func newMetadataItem(key string, value string) *compute.MetadataItems {
	return &compute.MetadataItems{
		Key:   key,
		Value: stringPtr(value),
	}
}

// https://cloud.google.com/compute/docs/instance-groups/rolling-out-updates-to-managed-instance-groups#starting_a_basic_rolling_update
func StartRollingUpdate(c *computeBeta.Service, d Deploy, instanceTemplateURL string) error {
	s := computeBeta.NewRegionInstanceGroupManagersService(c)

	ig, err := s.Get(d.Project, d.Region, d.InstanceGroup).Do()
	if err != nil {
		return fmt.Errorf("get instance group '%v/%v': %v", d.Project, d.InstanceGroup, err)
	}

	// TODO consider making the following check a configuration flag
	latestVersion := findLatestInstanceGroupManagerVersion(ig.Versions)
	if latestVersion != "" && !VersionLessThan(latestVersion, d.InstanceTemplate) {
		return fmt.Errorf("update instance group: instance template '%v' is too old, because '%v' is the latest instance template.", d.InstanceTemplate, latestVersion)
	}

	ig.InstanceTemplate = "" // make sure it's empty

	ig.Versions = []*computeBeta.InstanceGroupManagerVersion{
		{
			InstanceTemplate: instanceTemplateURL,
			Name:             d.InstanceTemplate,
		},
	}

	if ig.UpdatePolicy == nil {
		ig.UpdatePolicy = &computeBeta.InstanceGroupManagerUpdatePolicy{}
	}

	// force the following fields
	ig.UpdatePolicy.Type = d.UpdatePolicy.Type
	ig.UpdatePolicy.MinimalAction = d.UpdatePolicy.MinimalAction
	ig.UpdatePolicy.ReplacementMethod = d.UpdatePolicy.ReplacementMethod

	ig.UpdatePolicy.MinReadySec = int64(d.UpdatePolicy.minReadySec)
	ig.UpdatePolicy.ForceSendFields = []string{"MinReadySec"}

	if d.UpdatePolicy.maxSurgeInPercent {
		ig.UpdatePolicy.MaxSurge = &computeBeta.FixedOrPercent{Percent: int64(d.UpdatePolicy.maxSurge), ForceSendFields: []string{"Percent"}}
	} else {
		ig.UpdatePolicy.MaxSurge = &computeBeta.FixedOrPercent{Fixed: int64(d.UpdatePolicy.maxSurge), ForceSendFields: []string{"Fixed"}}
	}

	if d.UpdatePolicy.maxUnavailableInPercent {
		ig.UpdatePolicy.MaxUnavailable = &computeBeta.FixedOrPercent{Percent: int64(d.UpdatePolicy.maxUnavailable), ForceSendFields: []string{"Percent"}}
	} else {
		ig.UpdatePolicy.MaxUnavailable = &computeBeta.FixedOrPercent{Fixed: int64(d.UpdatePolicy.maxUnavailable), ForceSendFields: []string{"Fixed"}}
	}

	// wait until ready
	retry := 0
	for {
		_, err = s.Patch(d.Project, d.Region, d.InstanceGroup, ig).Do()
		if err != nil && isNotReadyErr(err) {
			time.Sleep(2 * time.Second)
			retry++
			if retry > 10 {
				return fmt.Errorf("update instance group: too many retries")
			}
			continue

		} else if err != nil {
			return fmt.Errorf("update instance group: %v", err)
		} else if err == nil {
			return nil
		}
	}
}

func CleanupInstanceTemplates(c *compute.Service, project string, after time.Duration) error {
	s := compute.NewInstanceTemplatesService(c)

	l, err := s.List(project).Do()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, item := range l.Items {

		// skip if this instance template was not created by us
		if !strings.Contains(item.Description, instanceTemplateDescription) {
			continue
		}

		// parse time and skip if the instance template is not old enough
		t, err := time.Parse(time.RFC3339, item.CreationTimestamp)
		if err != nil {
			return err
		}

		if !time.Now().UTC().After(t.UTC().Add(after)) {
			continue
		}

		// actually delete the instance template
		wg.Add(1)
		go func(instanceTemplate string) {
			defer wg.Done()
			if _, err := s.Delete(project, instanceTemplate).Do(); err != nil && !isInUseByAnotherResource(err) {
				LogWarning(err.Error(), nil)
			}
			Infof("Deleted old instance template '%v/%v'", project, instanceTemplate)
		}(item.Name)
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

func isInUseByAnotherResource(err error) bool {
	return isReasonErr(err, "resourceInUseByAnotherResource")
}

func stringPtr(in string) *string {
	return &in
}

func findLatestInstanceGroupManagerVersion(versions []*computeBeta.InstanceGroupManagerVersion) (name string) {
	latest := ""

	for _, v := range versions {
		if latest < v.Name {
			latest = v.Name
		}
	}

	return latest
}
