package validator

import (
	"errors"
	"fmt"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

func ValidatePipelineYaml(content []byte) error {
	loader := jsonschema.SchemeURLLoader{
		"file":  jsonschema.FileLoader{},
		"http":  newHTTPURLLoader(false),
		"https": newHTTPURLLoader(false),
	}

	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(loader)

	var data map[string]interface{}
	err := yaml.Unmarshal(content, &data)
	if err != nil {
		return err
	}

	url := "https://bitbucket.org/atlassianlabs/intellij-bitbucket-references-plugin/raw/master/src/main/resources/schemas/bitbucket-pipelines.schema.json"
	schema, err := compiler.Compile(url)
	if err != nil {
		return err
	}

	return schema.Validate(data)
}

func PrintError(err error) {
	var vErr *jsonschema.ValidationError
	if errors.As(err, &vErr) {
		fmt.Printf("%v\n", vErr)
	} else {
		fmt.Println(err)
	}
}
