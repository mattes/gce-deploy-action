package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

var environ = os.Environ()

type GithubActionConfig struct {
	Config                           string
	GoogleApplicationCredentials     string
	googleApplicationCredentialsData string
}

func ReadGithubActionConfig() (*GithubActionConfig, error) {
	c := &GithubActionConfig{}

	c.Config = os.Getenv("INPUT_CONFIG")
	if c.Config == "" {
		c.Config = "deploy.yml"
	}

	// read Google Application Credentials if this is a path
	c.GoogleApplicationCredentials = os.Getenv("INPUT_CREDS")
	f, err := ioutil.ReadFile(c.GoogleApplicationCredentials)
	if err == nil {
		c.googleApplicationCredentialsData = string(f)
	} else {
		c.googleApplicationCredentialsData = c.GoogleApplicationCredentials
	}

	return c, nil
}

func ReadConfigFile(path string) (io.ReadCloser, error) {
	paths := []string{path}

	switch filepath.Ext(path) {
	case "yml":
		paths = append(paths, path[:len(path)-4]+".yaml")
	case "yaml":
		paths = append(paths, path[:len(path)-5]+".yml")
	}

	for _, p := range paths {
		f, err := os.Open(p)
		if err == nil {
			return f, nil
		}
	}

	return nil, fmt.Errorf("config: %v", path)
}

type Config struct {
	DeleteInstanceTemplatesAfter string `yaml:"delete_instance_templates_after"`
	deleteInstanceTemplatesAfter time.Duration
	Deploys                      []Deploy `yaml:"deploys"`
}

type Deploy struct {
	Name                             string `yaml:"name"`
	Project                          string `yaml:"project"`
	GoogleApplicationCredentials     string `yaml:"creds"`
	googleApplicationCredentialsData string
	Region                           string `yaml:"region"`
	InstanceGroup                    string `yaml:"instance_group"`
	InstanceTemplateBase             string `yaml:"instance_template_base"`
	InstanceTemplate                 string `yaml:"instance_template"`
	StartupScriptPath                string `yaml:"startup_script"`
	startupScript                    string
	ShutdownScriptPath               string `yaml:"shutdown_script"`
	shutdownScript                   string
	CloudInitPath                    string `yaml:"cloud_init"`
	cloudInit                        string
	Vars                             map[string]string `yaml:"vars"`
	Labels                           map[string]string `yaml:"labels"`
	Metadata                         map[string]string `yaml:"metadata"`
	Tags                             []string          `yaml:"tags"`
}

func ParseConfig(b io.Reader) (*Config, error) {
	c := &Config{}
	d := yaml.NewDecoder(b)
	d.SetStrict(true)
	if err := d.Decode(c); err != nil && err != io.EOF {
		return nil, fmt.Errorf("config: %v", err)
	}

	// if DeleteInstanceTemplatesAfter is not set to false
	if c.DeleteInstanceTemplatesAfter != "false" {
		// parse and set duration if set
		if c.DeleteInstanceTemplatesAfter != "" {
			duration, err := time.ParseDuration(c.DeleteInstanceTemplatesAfter)
			if err != nil {
				return nil, err
			}
			c.deleteInstanceTemplatesAfter = duration
		} else {
			// or set default
			c.deleteInstanceTemplatesAfter = 24 * time.Hour * 14 // 14 days
		}
	}

	// expand env variables
	for i := range c.Deploys {
		dy := &c.Deploys[i]

		dy.Name = expandShellRe(dy.Name, getEnv(nil))
		if dy.Name == "" {
			return nil, fmt.Errorf("deploy item #%v needs name", i+1)
		}

		dy.Project = expandShellRe(dy.Project, getEnv(nil))

		dy.GoogleApplicationCredentials = expandShellRe(dy.GoogleApplicationCredentials, getEnv(nil))

		f, err := ioutil.ReadFile(dy.GoogleApplicationCredentials)
		if err == nil {
			dy.googleApplicationCredentialsData = string(f)
		} else {
			dy.googleApplicationCredentialsData = dy.GoogleApplicationCredentials
		}

		dy.Region = expandShellRe(dy.Region, getEnv(nil))
		if dy.Region == "" {
			return nil, fmt.Errorf("deploy '%v' needs region", dy.Name)
		}

		dy.InstanceGroup = expandShellRe(dy.InstanceGroup, getEnv(nil))
		if dy.InstanceGroup == "" {
			return nil, fmt.Errorf("deploy '%v' needs instance_group", dy.Name)
		}

		dy.InstanceTemplateBase = expandShellRe(dy.InstanceTemplateBase, getEnv(nil))
		if dy.InstanceTemplateBase == "" {
			return nil, fmt.Errorf("deploy '%v' needs instance_template_base", dy.Name)
		}

		dy.InstanceTemplate = expandShellRe(dy.InstanceTemplate, getEnv(nil))
		if dy.InstanceTemplate == "" {
			return nil, fmt.Errorf("deploy '%v' needs instance_template", dy.Name)
		}

		dy.StartupScriptPath = expandShellRe(dy.StartupScriptPath, getEnv(nil))

		dy.ShutdownScriptPath = expandShellRe(dy.ShutdownScriptPath, getEnv(nil))

		dy.CloudInitPath = expandShellRe(dy.CloudInitPath, getEnv(nil))

		for k, v := range dy.Vars {
			dy.Vars[k] = expandShellRe(v, getEnv(nil))
		}

		for k, v := range dy.Labels {
			dy.Labels[k] = expandShellRe(v, getEnv(nil))
		}

		for k, v := range dy.Metadata {
			dy.Metadata[k] = expandShellRe(v, getEnv(nil))
		}

		for j := range dy.Tags {
			dy.Tags[j] = expandShellRe(dy.Tags[j], getEnv(nil))
		}
	}

	// read contents of scripts and expand env vars
	for i := range c.Deploys {
		dy := &c.Deploys[i]

		if dy.StartupScriptPath != "" {
			f, err := ioutil.ReadFile(dy.StartupScriptPath)
			if err != nil {
				return nil, fmt.Errorf("startup_script: %v", err)
			}
			dy.startupScript = expandMakeRe(string(f), getEnv(dy.Vars))
		}

		if dy.ShutdownScriptPath != "" {
			f, err := ioutil.ReadFile(dy.ShutdownScriptPath)
			if err != nil {
				return nil, fmt.Errorf("shutdown_script: %v", err)
			}
			dy.shutdownScript = expandMakeRe(string(f), getEnv(dy.Vars))
		}

		if dy.CloudInitPath != "" {
			f, err := ioutil.ReadFile(dy.CloudInitPath)
			if err != nil {
				return nil, fmt.Errorf("cloud_init: %v", err)
			}
			dy.cloudInit = expandMakeRe(string(f), getEnv(dy.Vars))
		}
	}

	return c, nil
}

func getEnv(locals map[string]string) map[string]string {
	m := make(map[string]string)

	for _, v := range environ {
		x := strings.SplitN(v, "=", 2)
		m[strings.ToLower(x[0])] = x[1]
	}

	for k, v := range locals {
		m[strings.ToLower(k)] = v
	}

	return m
}

var (
	shellVarRe    = regexp.MustCompile(`\\?\${?([a-zA-Z]([a-zA-Z0-9-_]+[a-zA-Z0-9]|[a-zA-Z0-9]*)(:\d(:\d)?)?)}?`)
	makefileVarRe = regexp.MustCompile(`\\?\$\([a-zA-Z0-9_-]+\)`)
)

func expandShellRe(str string, vars map[string]string) string {
	return shellVarRe.ReplaceAllStringFunc(str, func(x string) string {

		if strings.HasPrefix(x, `\$`) {
			return x
		}

		x = strings.Trim(x, "${}")

		if !strings.Contains(x, ":") {
			return vars[strings.ToLower(x)]
		}

		// parse ${string:position[:length]} and truncate string
		parts := strings.Split(x, ":")
		switch len(parts) {
		default:
			fallthrough
		case 1:
			return vars[strings.ToLower(parts[0])]

		case 2:
			v := vars[strings.ToLower(parts[0])]

			from, err := strconv.Atoi(parts[1])
			if err != nil {
				return v
			}
			return v[from:]

		case 3:
			v := vars[strings.ToLower(parts[0])]

			from, err := strconv.Atoi(parts[1])
			if err != nil {
				return v
			}

			to, err := strconv.Atoi(parts[2])
			if err != nil {
				return v
			}
			return v[from : from+to]
		}
	})
}

func expandMakeRe(str string, vars map[string]string) string {
	return makefileVarRe.ReplaceAllStringFunc(str, func(x string) string {

		if strings.HasPrefix(x, `\$`) {
			return x
		}

		x = strings.Trim(x, "$()")

		return vars[strings.ToLower(x)]
	})
}
