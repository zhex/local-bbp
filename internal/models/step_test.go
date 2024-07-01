package models

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestStepScript(t *testing.T) {
	yamlData := `
name: Build and Deploy
image: my-image
script:
  - echo "Hello World"
  - pipe: atlassian/docker-build:1.1.0
    name: Docker Build
    variables:
      IMAGE_NAME: myimage
      TAG: latest
`
	var step Step
	_ = yaml.Unmarshal([]byte(yamlData), &step)

	assert.Equal(t, "Build and Deploy", step.Name)
	assert.Equal(t, 2, len(step.Script))

	script1 := step.Script[0]
	assert.Equal(t, ScriptTypeCmd, script1.Type())
	assert.Equal(t, "echo \"Hello World\"", script1.(*CmdScript).Cmd)

	script2 := step.Script[1]
	assert.Equal(t, ScriptTypePipe, script2.Type())
	pipe := script2.(*Pipe)
	assert.Equal(t, "atlassian/docker-build:1.1.0", pipe.Pipe)
	assert.Equal(t, 2, len(pipe.Variables))
}
