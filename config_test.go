package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	// write tmp file to be used as startup/shutdown script
	tmpFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	tmpFile.WriteString("Foo: $BAR ${{BAR}} ${{SCRIPTVARKEY}}")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	config := `
delete_instance_templates_after: 14h
deploys:
- name: name-${{BAR}}
  project: project-${{BAR}}
  creds: google-application-credentials-${{BAR}}
  region: region-${{BAR}}
  instance_group: instance-group-${{BAR}}
  instance_template_base: instance-template-base-${{BAR}}
  instance_template: instance-template-${{BAR}}
  startup_script: ` + tmpFile.Name() + `
  shutdown_script: ` + tmpFile.Name() + `
  cloud_init: ` + tmpFile.Name() + `
  vars:
    scriptvarkey: scriptvarvalue-${{BAR}} 
  labels:
    labelkey: labelvalue-${{BAR}}
  metadata:
    metadatakey: metadatavalue-${{BAR}}
  tags:
    - tagvalue-${{BAR}}
  update_policy:
    type: type-${{BAR}}
    minimal_action: minimal-action-${{BAR}}
    replacement_method: replacement-method-${{BAR}}
    min_ready_sec: ${{MIN_READY_SEC}}
    max_surge: ${{MAX_SURGE}}
    max_unavailable: ${{MAX_UNAVAILABLE}}
`

	environ = append(environ, "BAR=FOO")
	environ = append(environ, "MIN_READY_SEC=2")
	environ = append(environ, "MAX_SURGE=15%")
	environ = append(environ, "MAX_UNAVAILABLE=14")
	c, err := ParseConfig(strings.NewReader(config))
	require.NoError(t, err)

	assert.Equal(t, "14h", c.DeleteInstanceTemplatesAfter)
	assert.Equal(t, 14*time.Hour, c.deleteInstanceTemplatesAfter)

	require.Len(t, c.Deploys, 1)
	require.Len(t, c.Deploys[0].Vars, 1)
	require.Len(t, c.Deploys[0].Labels, 1)
	require.Len(t, c.Deploys[0].Metadata, 1)
	require.Len(t, c.Deploys[0].Tags, 1)

	assert.Equal(t, "name-FOO", c.Deploys[0].Name)
	assert.Equal(t, "project-FOO", c.Deploys[0].Project)
	assert.Equal(t, "google-application-credentials-FOO", c.Deploys[0].GoogleApplicationCredentials)
	assert.Equal(t, "region-FOO", c.Deploys[0].Region)
	assert.Equal(t, "instance-group-FOO", c.Deploys[0].InstanceGroup)
	assert.Equal(t, "instance-template-base-FOO", c.Deploys[0].InstanceTemplateBase)
	assert.Equal(t, "instance-template-FOO", c.Deploys[0].InstanceTemplate)
	assert.Equal(t, tmpFile.Name(), c.Deploys[0].StartupScriptPath)
	assert.Equal(t, "Foo: $BAR FOO scriptvarvalue-FOO", c.Deploys[0].startupScript)
	assert.Equal(t, tmpFile.Name(), c.Deploys[0].ShutdownScriptPath)
	assert.Equal(t, "Foo: $BAR FOO scriptvarvalue-FOO", c.Deploys[0].shutdownScript)
	assert.Equal(t, tmpFile.Name(), c.Deploys[0].CloudInitPath)
	assert.Equal(t, "Foo: $BAR FOO scriptvarvalue-FOO", c.Deploys[0].cloudInit)
	assert.Equal(t, "scriptvarvalue-FOO", c.Deploys[0].Vars["scriptvarkey"])
	assert.Equal(t, "labelvalue-FOO", c.Deploys[0].Labels["labelkey"])
	assert.Equal(t, "metadatavalue-FOO", c.Deploys[0].Metadata["metadatakey"])
	assert.Equal(t, "tagvalue-FOO", c.Deploys[0].Tags[0])

	assert.Equal(t, "type-FOO", c.Deploys[0].UpdatePolicy.Type)
	assert.Equal(t, "minimal-action-FOO", c.Deploys[0].UpdatePolicy.MinimalAction)
	assert.Equal(t, "replacement-method-FOO", c.Deploys[0].UpdatePolicy.ReplacementMethod)
	assert.Equal(t, "2", c.Deploys[0].UpdatePolicy.MinReadySec)
	assert.Equal(t, 2, c.Deploys[0].UpdatePolicy.minReadySec)
	assert.Equal(t, "15%", c.Deploys[0].UpdatePolicy.MaxSurge)
	assert.Equal(t, 15, c.Deploys[0].UpdatePolicy.maxSurge)
	assert.Equal(t, true, c.Deploys[0].UpdatePolicy.maxSurgeInPercent)
	assert.Equal(t, "14", c.Deploys[0].UpdatePolicy.MaxUnavailable)
	assert.Equal(t, 14, c.Deploys[0].UpdatePolicy.maxUnavailable)
	assert.Equal(t, false, c.Deploys[0].UpdatePolicy.maxUnavailableInPercent)
}

func TestParseConfigWithCommonConfig(t *testing.T) {
	// write tmp file to be used as startup/shutdown script
	tmpFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	tmpFile.WriteString("common-file")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	config := `
common: 
  project: commonproject
  region: commonregion
  startup_script: ` + tmpFile.Name() + `
  shutdown_script: ` + tmpFile.Name() + `
  cloud_init: ` + tmpFile.Name() + `
  vars:
    var1: commonvar1
    var2: commonvar2
  labels:
    label1: commonlabel1
    label2: commonlabel2
  metadata:
    metadata1: commonmetadata1
    metadata2: commonmetadata2
  tags:
    - commontag1
    - commontag2
  update_policy:
    type: commontype
    minimal_action: common-minimal-action
    replacement_method: common-replacement-method
    min_ready_sec: 10
    max_surge: 11
    max_unavailable: 12

deploys:
  - name: test
    instance_group: x
    instance_template_base: y
    instance_template: z
    vars:
      var1: var1
    labels:
      label1: label1
    metadata:
      metadata1: metadata1
    tags:
      - tag1
`

	c, err := ParseConfig(strings.NewReader(config))
	require.NoError(t, err)

	assert.Equal(t, "test", c.Deploys[0].Name)
	assert.Equal(t, "commonproject", c.Deploys[0].Project)
	assert.Equal(t, "commonregion", c.Deploys[0].Region)
	assert.Equal(t, tmpFile.Name(), c.Deploys[0].StartupScriptPath)
	assert.Equal(t, tmpFile.Name(), c.Deploys[0].ShutdownScriptPath)
	assert.Equal(t, tmpFile.Name(), c.Deploys[0].CloudInitPath)

	require.Len(t, c.Deploys[0].Vars, 2)
	assert.Equal(t, "var1", c.Deploys[0].Vars["var1"])
	assert.Equal(t, "commonvar2", c.Deploys[0].Vars["var2"])

	require.Len(t, c.Deploys[0].Labels, 2)
	assert.Equal(t, "label1", c.Deploys[0].Labels["label1"])
	assert.Equal(t, "commonlabel2", c.Deploys[0].Labels["label2"])

	require.Len(t, c.Deploys[0].Metadata, 2)
	assert.Equal(t, "metadata1", c.Deploys[0].Metadata["metadata1"])
	assert.Equal(t, "commonmetadata2", c.Deploys[0].Metadata["metadata2"])

	require.Len(t, c.Deploys[0].Tags, 3)
	assert.Equal(t, "tag1", c.Deploys[0].Tags[0])
	assert.Equal(t, "commontag1", c.Deploys[0].Tags[1])
	assert.Equal(t, "commontag2", c.Deploys[0].Tags[2])

	assert.Equal(t, "commontype", c.Deploys[0].UpdatePolicy.Type)
	assert.Equal(t, "common-replacement-method", c.Deploys[0].UpdatePolicy.ReplacementMethod)
	assert.Equal(t, "common-minimal-action", c.Deploys[0].UpdatePolicy.MinimalAction)
	assert.Equal(t, "10", c.Deploys[0].UpdatePolicy.MinReadySec)
	assert.Equal(t, "11", c.Deploys[0].UpdatePolicy.MaxSurge)
	assert.Equal(t, "12", c.Deploys[0].UpdatePolicy.MaxUnavailable)
}

func TestNilMaps(t *testing.T) {
	config := `
common: 
  vars:
    var1: x
  labels:
    label1: x
  metadata:
    metadata1: x
  tags:
    - tag1

deploys:
  - name: test
    region: w
    instance_group: x
    instance_template_base: y
    instance_template: z
`

	_, err := ParseConfig(strings.NewReader(config))
	require.NoError(t, err)
}

func TestExpandVars(t *testing.T) {
	in := `f fo foo $f $fo $foo ${f} ${fo} ${foo} \${{f}} \${{fo}} \${{foo}} ${{f}} ${{fo}} ${{foo}} ${{ f }} ${{ fo }} ${{ foo }} a${{f}}b a${{fo}}b a${{foo}}b a\${{f}}b a\${{fo}}b a\${{foo}}b`

	vars := map[string]string{
		"f":   "b",
		"fo":  "ba",
		"foo": "bar",
	}

	out := expandVars(in, vars)
	assert.Equal(t, `f fo foo $f $fo $foo ${f} ${fo} ${foo} \${{f}} \${{fo}} \${{foo}} b ba bar b ba bar abb abab abarb a\${{f}}b a\${{fo}}b a\${{foo}}b`, out)
}

func TestVariableCaseInsensitive(t *testing.T) {
	in := `${{foo}} ${{FOO}}`

	vars := map[string]string{
		"foo": "bar",
	}

	out := expandVars(in, vars)
	assert.Equal(t, `bar bar`, out)
}

func TestVariableTruncate(t *testing.T) {
	in := `${{foo:0}} ${{foo:1}} ${{foo:7}} ${{foo:7:3}} ${{foo:1:1}}`

	vars := map[string]string{
		"foo": "abcABC123ABCabc",
	}

	out := expandVars(in, vars)
	assert.Equal(t, `abcABC123ABCabc bcABC123ABCabc 23ABCabc 23A b`, out)
}
