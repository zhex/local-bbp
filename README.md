# Local-BBP

![GitHub License](https://img.shields.io/github/license/zhex/local-bbp)
![GitHub Release](https://img.shields.io/github/v/release/zhex/local-bbp)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/zhex/local-bbp)


Local-BBP is an open-source CLI tool written in Go, designed to simulate Bitbucket pipelines on your local machine. This tool allows developers to test and debug their pipeline configurations faster and more efficiently before pushing to Bitbucket.


## Why Local-BBP?

As a developer, I often find that the first few pipeline runs for a new project fail due to configuration issues. This can be incredibly time-consuming and frustrating, as each failure requires a push to Bitbucket, a wait for the pipeline to run, and then reviewing the results to identify and fix issues. With Local-BBP, I can run and debug these pipelines locally, saving time and reducing the number of failed runs in the actual CI environment.

## Supported Features

- [x] Run pipeline
  - [x] Sequential steps
  - [x] Parallel steps
- [x] Custom Step Image
- [x] Private image repository access
- [x] Artifacts
- [x] Smart caches
- [x] Pipe
- [x] Services
  - [x] Sidecars
  - [x] Default Docker
  - [x] Custom Docker
- [x] Secret variables from file
- [x] Stage
- [x] Conditional steps
- [ ] Settings
  - [x] Default image
  - [x] Timeout
  - [ ] Container Size
- [x] Validate bitbucket-pipelines.yml file
- [x] Integration marketplace search

## Installation

You can install the latest version of Local-BBP using the following command:

```bash
curl -s https://raw.githubusercontent.com/zhex/local-bbp/main/scripts/install.sh | bash
```

### Usages

You need to have Docker installed on your machine to run Local-BBP.

To simulate your Bitbucket pipeline locally, navigate to the project directory containing your bitbucket-pipelines.yml file and run the following command to list the available pipelines:

```bash
bbp list
```

then run the default pipeline:

```bash
bbp run
```

This command will execute your default pipeline steps as defined in the bitbucket-pipelines.yml file.

You can also specify a specific pipeline to run using the -n flag:

```bash
bbp run -n default
```

When first running the command, Local-BBP will init the config file in the `~/.bbp/config.json` also download the Linux docker cli binaries.

example config file:

```json5
{
    // the default workdir in the build container
    "workDir": "/opt/atlassian/pipelines/agent/build",
    
    // the default image to use in the build container if not specified in the bitbucket-pipelines.yaml file
    "defaultImage": "atlassian/default-image:4",
    
    // the default output directory for the pipeline results, base on the project directory if relative path is used 
    "outputDir": "bbp",
    
    // the default linux docker version to use in the build container 
    // download automatically from https://download.docker.com/linux/static/stable/
    "dockerVersion": "19.03.15",
    
    // the default docker image for running the docker daemon for the build container
    "defaultDockerImage": "docker:27.0.3-dind-alpine3.20",
    
    // the path for the cli to download required tools
    "toolDir": "/Users/zhex/.bbp/tools",
    
    // default timeout for a single pipeline step (in minutes)
    "maxStepTimout": 120,
    
    // default timeout for the whole pipeline (in minutes)
    "maxPipeTimout": 240,
}
```

You can also specify a path to the project directory containing the [bitbucket-pipelines.yml](https://support.atlassian.com/bitbucket-cloud/docs/bitbucket-pipelines-configuration-reference/) file:

```bash
bbp run -n "pr/**" -p /path/to/project
```

Support secret variables by providing a path to the secrets file:

```bash
bbp run -n default -s /path/to/secrets
```

The secret file format is the same as dot env file. Sample secrets file:

```dotenv
MY_SECRET="my-secret-value"
MY_OTHER_SECRET="my-other-secret-value"
```

use the -v flag to view the verbose output for more details:

```bash
bbp run -n default -v
```

Also, you can validate your bitbucket-pipelines.yml file using the following command:

```bash
bbp validate
```

The validation rules are based on the official Bitbucket Pipelines configuration [schema](https://bitbucket.org/atlassianlabs/intellij-bitbucket-references-plugin/raw/master/src/main/resources/schemas/bitbucket-pipelines.schema.json).

## Differences between Local-BBP and Bitbucket Pipelines

Local-BBP is designed to simulate Bitbucket Pipelines as closely as possible, but there are some differences between the two:

- **Environment**: Local-BBP runs pipelines on your local machine, so it may not have access to the same resources as the Bitbucket Pipelines environment.
- **Features**: Local-BBP does not support all the features of Bitbucket Pipelines, such as host runners and custom runner size.
- **Service Access**: In Local-BBP, service names are used as hostnames similar to Docker Compose. In Bitbucket Pipelines, sidecar services are accessed via localhost.
- **Step Condition**: Bitbucket Pipeline compares all commits between source and target branches in pull-request pipelines,, while in other pipelines, it compares the last commit. Local-BBP includes uncommitted changes for easier development.

## License

Local-BBP is released under the MIT License. See [LICENSE](LICENSE) for more details.

## Support

For any issues or questions, please open an issue on our GitHub repository.
