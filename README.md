# Local-BBP

Local-BBP is an open-source CLI tool written in Go, designed to simulate Bitbucket pipelines on your local machine. This tool allows developers to test and debug their pipeline configurations faster and more efficiently before pushing to Bitbucket.


## Why Local-BBP?

As a developer, I often find that the first 5-10 pipeline runs for a new project fail due to configuration issues. This can be incredibly time-consuming and frustrating, as each failure requires a push to Bitbucket, a wait for the pipeline to run, and then reviewing the results to identify and fix issues. With Local-BBP, I can run and debug these pipelines locally, saving time and reducing the number of failed runs in the actual CI environment.

## Installation

You can install Local-BBP using Homebrew:

```bash
brew install bbp
```
Alternatively, you can install it using Go:

```bash
go install github.com/zhex/local-bbp
```

Ensure your $GOPATH/bin is in your PATH to run Local-BBP directly.

### Usage

You need to have Docker installed on your machine to run Local-BBP.

To simulate your Bitbucket pipeline locally, navigate to the directory containing your bitbucket-pipelines.yml file and run:

```bash
bbp
```

This command will execute your default pipeline steps as defined in the bitbucket-pipelines.yml file.

You can also specify a specific pipeline to run using the -n flag:

```bash
bbp -n my-pipeline
```

You can also specify a path to the project directory containing the bitbucket-pipelines.yml file:

```bash
bbp -p /path/to/project
```

## License

Local-BBP is released under the MIT License. See [LICENSE](LICENSE) for more details.

## Support

For any issues or questions, please open an issue on our GitHub repository.
