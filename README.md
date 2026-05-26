<!-- #GFCFilterMarkerStart# -->

[[_TOC_]] 

<!-- #GFCFilterMarkerEnd# --> 

# MongoDB Operator

## Repository Structure

* `./api` - directory with API for Mongodb Operator, it contains all parameters definition required for MongoDB operator to start reconciliation process.
* `./bin` - directory with `controller-gen` binary that is called from Makefile to generate CRD and deep copy methods.
* `./build` - the directory with etrypoint for Docker image of Mongodb Operator.
* `./charts` - the directory with HELM chart for Mongodb Operator.
* `./docs` - the directory with actual documentation for the service.
* `./examples` - the irectory with deploy parameters examples.
* `./extras` - the directory with files that might be useful for development.
* `./hack` - the directory with licence file.
* `./mongo` - the directory with `description.yaml` and `build.sh` file for promotion process.
* `./config` - the directory contains k8s templates, not used in our project.
* `./controller` - the directory with operator sdk controller.
* `./pkg` - the directory with source code.
* `.gitlab-ci.yml` - the CI/CD pipelines configuration.
* `./build.sh` - the entrypoint for build job, it starts docker image build and copies charts.
* `./description.yaml` - descibes buld sructure of Mongodb Operator docker image.
* `./Dockerfile` - the Dockerfile for Mongodb Operator docker image.
* `./go.mod` - the go.mod of the project.
* `./go.sum` - the go.sum of the project.
* `./main.go` - the entrypoint of the Mongodb Operator.
* `./Makefile` - the Makefile to generate CRD and other code.
* `./module_test.go` - the unit tests of Mongodb Operator.
* `./PROJECT` - the file is used to track the info used to scaffold the project.
* `./renovate.json` - the configuration for renovate bot.


### Smoke Tests

There are no smoke tests.

### How to Debug

The information on debug is provided below.

#### Mongodb Operator

To debug Operator in VSCode you can use `Launch Mongodb Operator` configuration which is already defined in 
`.vscode/launch.json` file. 

The developer should configure environment variables: 

* `KUBECONFIG` - developer **need to define** `KUBECONFIG` environment variable
  which should contains path to the kube-config file. It can be defined on configuration level
  or on the level of user's environment variables.
* `NAMESPACE` - namespace, in which custom resources should be proceeded.


### How to Troubleshoot

There are no well-defined rules for troubleshooting, as each task is unique, but there are some tips that can do:

* Deploy parameters.
* Application manifest.
* Logs from all Mongodb pods.

Also, developer can take a look on [Troubleshooting guide](/docs/public/troubleshooting.md).


## Useful Links

* [Installation guide](/docs/public/installation_guide.md).
* [Troubleshooting guide](/docs/public/troubleshooting.md).
* [Architecture Guide](/docs/public/architecture.md).
