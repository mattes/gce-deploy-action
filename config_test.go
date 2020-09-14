package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	// write tmp file to be used as startup/shutdown script
	tmpFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	tmpFile.WriteString("Foo: $BAR $(BAR) $(SCRIPTVARKEY)")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	config := `
deploys:
- name: name-$BAR-${BAR}
  project: project-$BAR-${BAR}
  google_application_credentials: google-application-credentials-$BAR-${BAR}
  region: region-$BAR-${BAR}
  instance_group: instance-group-$BAR-${BAR}
  instance_template_base: instance-template-base-$BAR-${BAR}
  instance_template: instance-template-$BAR-${BAR}
  startup_script: ` + tmpFile.Name() + `
  shutdown_script: ` + tmpFile.Name() + `
  cloud_init: ` + tmpFile.Name() + `
  vars:
    scriptvarkey: scriptvarvalue-$BAR-${BAR} 
  labels:
    labelkey: labelvalue-$BAR-${BAR}
  metadata:
    metadatakey: metadatavalue-$BAR-${BAR}
  tags:
    - tagvalue-$BAR-${BAR}
`

	environ = append(environ, "BAR=FOO")
	c, err := ParseConfig(strings.NewReader(config))
	require.NoError(t, err)

	require.Len(t, c.Deploys, 1)
	require.Len(t, c.Deploys[0].Vars, 1)
	require.Len(t, c.Deploys[0].Labels, 1)
	require.Len(t, c.Deploys[0].Metadata, 1)
	require.Len(t, c.Deploys[0].Tags, 1)

	assert.Equal(t, "name-FOO-FOO", c.Deploys[0].Name)
	assert.Equal(t, "project-FOO-FOO", c.Deploys[0].Project)
	assert.Equal(t, "google-application-credentials-FOO-FOO", c.Deploys[0].GoogleApplicationCredentials)
	assert.Equal(t, "region-FOO-FOO", c.Deploys[0].Region)
	assert.Equal(t, "instance-group-FOO-FOO", c.Deploys[0].InstanceGroup)
	assert.Equal(t, "instance-template-base-FOO-FOO", c.Deploys[0].InstanceTemplateBase)
	assert.Equal(t, "instance-template-FOO-FOO", c.Deploys[0].InstanceTemplate)
	assert.Equal(t, tmpFile.Name(), c.Deploys[0].StartupScriptPath)
	assert.Equal(t, "Foo: $BAR FOO scriptvarvalue-FOO-FOO", c.Deploys[0].startupScript)
	assert.Equal(t, tmpFile.Name(), c.Deploys[0].ShutdownScriptPath)
	assert.Equal(t, "Foo: $BAR FOO scriptvarvalue-FOO-FOO", c.Deploys[0].shutdownScript)
	assert.Equal(t, tmpFile.Name(), c.Deploys[0].CloudInitPath)
	assert.Equal(t, "Foo: $BAR FOO scriptvarvalue-FOO-FOO", c.Deploys[0].cloudInit)
	assert.Equal(t, "scriptvarvalue-FOO-FOO", c.Deploys[0].Vars["scriptvarkey"])
	assert.Equal(t, "labelvalue-FOO-FOO", c.Deploys[0].Labels["labelkey"])
	assert.Equal(t, "metadatavalue-FOO-FOO", c.Deploys[0].Metadata["metadatakey"])
	assert.Equal(t, "tagvalue-FOO-FOO", c.Deploys[0].Tags[0])
}

func TestExpandShellRe(t *testing.T) {
	in := `$foo $FOO ${foo} a${foo}b \$foo \${foo} a\${foo}b $foo-$foo $foo-${foo} $fo $f`

	vars := map[string]string{
		"f":   "b",
		"fo":  "ba",
		"foo": "bar",
	}

	out := expandShellRe(in, vars)
	assert.Equal(t, `bar bar bar abarb \$foo \${foo} a\${foo}b bar-bar bar-bar ba b`, out)
}

func TestShellReTruncate(t *testing.T) {
	in := `${foo:0} ${foo:1} ${foo:7} ${foo:7:3} ${foo:1:1}`

	vars := map[string]string{
		"foo": "abcABC123ABCabc",
	}

	out := expandShellRe(in, vars)
	assert.Equal(t, `abcABC123ABCabc bcABC123ABCabc 23ABCabc 23A b`, out)
}

func TestExpandMakeRe(t *testing.T) {
	in := `$(foo) $(FOO) a$(foo)b \$(foo) a\$(foo)b $(foo)-$(foo) $(fo) $(f)`

	vars := map[string]string{
		"f":   "b",
		"fo":  "ba",
		"foo": "bar",
	}

	out := expandMakeRe(in, vars)
	assert.Equal(t, `bar bar abarb \$(foo) a\$(foo)b bar-bar ba b`, out)
}
