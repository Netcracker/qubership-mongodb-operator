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

## Artifacts Described

The delivery of this repository contains several artifacts:

* **Mongodb Operator Docker image**, for instance `artifactorycn.netcracker.com:17008/product/prod.platform.databases_mongodb-operator:master_20241029-200359` - the image to be deployed to Kubernetes/Openshift, contains all logic of the Operator written on Go. Usually we do not need to provide it separately, but for development purpose it might be convinient to update only the image of Operator deployment in Kubernetes/Openshift to test changes 

* **Mongodb Operator Manifest Version**, for instance `master-mongo_7.0.5-20241030.001209-70` - the artifact contains Helm chart of Mongodb Operator and all its integrations like Mongodb Operator docker image, Disaster Recovery Daemon docker image and so on. In other words this is complete build of the project and promoted version of this artifact is delivered to customers.

## How to Start

### Build On-commit

The Mongodb operator Docker image and Mongodb Operator manifest are built automatically by CI pipelines on each commit

To get the Mongodb operator image:
1. Navigate to the Pipeline page of Gitlab and find your pipeline.
1. Find and open the stage Operator.
1. In the bottom of the job log find the Docker image address.

![Docker image of Mongodb Operator](/docs/internal/images/pipeline_docker_image.png)

To get the Mongodb operator manifest version:
1. Navigate to the Pipeline page of Gitlab and find your pipeline.
1. Find and open the stage Manifest Mongo <Version>.
1. In the bottom of the job log find the App Manifest.

![Version](/docs/internal/images/pipeline_app_version.png)

### Build Manually

The manual steps to build Mongodb Operator docker image and Manifest are described below.

#### Mongodb Operator Docker Image

1. Go to [DP-Builder](https://cisrvrecn.netcracker.com/job/DP.Pub.Microservice_builder_v2/)
2. Go to the "Build with parameters" tab.
3. Specify the following:

   * REPOSITORY_NAME - `PROD.Platform.Databases/mongodb-operator`.
   * LOCATION - your dev branch or `master`.
   
4. Click the **Build** button.
5. Find your running build in the “Build History” tab in the DP-Builder page.
6. Wait for the job to finish.
7. Navigate to the job Report section and find the docker image address.

![Version](/docs/internal/images/docker_job.png)

#### Mongodb Operator Manifest

1. Actualize locations in `integration` section in `/mongo/<MONGO_VERSION>/description.yaml` file.
   If one of the integrations should be taken from other location, then need to specify related branch in `location`.
   For example, if you need to integrate another version of `PROD.Platform.Streaming/disaster-recovery-daemon`, then you need to specify the new branch or tag in its integration section:

    ```yaml
    - type: find-latest-deployment-descriptor
      repo: PROD.Platform.Streaming/disaster-recovery-daemon
      location: <NEW_BRANCH_OR_TAG>
      docker-image-id: 'timestamp'
      deploy-param: disasterRecoveryImage
      service-name: disasterRecoveryImage
    ```

2. Go to [DP-Builder](https://cisrvrecn.netcracker.com/job/DP.Pub.Microservice_builder_v2/).
3. Go to the "Build with parameters" tab.
4. Specify the following:

   * REPOSITORY_NAME - `PROD.Platform.Databases/mongodb-operator`.
   * LOCATION - your dev branch or `master`.
   * PREFIX - `charts`.
   * PARAMETERS:

   ```json
    {integration: [{type: find-latest-deployment-descriptor, repo: "PROD.Platform.Databases/mongodb-operator", location: <YOU_CURRENT_BRANCH_OF_OPERATOR>, docker-image-id: timestamp, deploy-param: operator}, {type: scripts, repo: "PROD.Platform.Databases/mongodb-operator", location: <YOU_CURRENT_BRANCH_OF_OPERATOR>, alias: timestamp, zip-name: migration-artifacts.zip, name: migration-artifacts, pre-deploy-entry-point: migration-artifacts/migration-artifacts.sh}]}
   ```
   **Important!** Update location parameter in PARAMETRS json as well
   
5. Click the **Build** button.
6. Find your running build in the “Build History” tab in the DP-Builder page.
7. Wait for the job to finish.
8. In finished build job find: `Services` column --> `independent` topic --> `operator` line.
9. Follow **the first link** in the `Artifact` column in the `operator` line.

![dmp artifact](/docs/internal/images/dmp_artifact.png)

10. Copy the `version` from the `Dependency Declaration` section.

![Version](/docs/internal/images/app_version.png)

11. Combine **the service name** and **the version** by the following way: `mongodb:<version>`.
12. The combination result is an **application manifest**.

#### Definition of Done

The changes might be marked as fully done, if it accepts the following criteria:

1. The ticket's task done.
2. The solution is deployed to dev environment, where it can be tested.
3. Created merge request has:
   1. "Green" pipeline (linter, build, deploy & test jobs are passed).
   2. The title should follow the naming conversation: `<TMS-TICKET-ID>: <CHANGES-SHORT-DESCRIPTION>`.
   3. The description is **fully** filled.

### Deploy to k8s

#### Pure Helm

1. Build operator and integration tests, if you need non-master versions.
2. Prepare kubeconfig on you host machine to work with target cluster.
3. Prepare `sample.yaml` file with deployment parameters, which should contains custom docker images if it is needed.
4. Store `sample.yaml` file in `/charts/helm/mongodb-operator` directory.
5. Go to `/charts/helm/mongodb-operator` directory.
6. Run the following command if you deploy Mongodb:

     ```sh
     # Run in /charts/helm/mongodb-operator directory
     helm install mongodb-operator ./ -f sample.yaml -n <TARGET_NAMESPACE>
     ```

#### Application deployer

To deploy Mongodb you needed to run a deployer jobs: 

The algorithm is the following:

1. Build a manifest using [Mongodb Operator Manifest](#mongodb-operator-manifest) on-commit.
2. Prepare Cloud Deploy Configuration:
   1. Go to the [APP-Deployer infrastructure configuration](https://cloud-deployer.netcracker.com/job/INFRA/job/groovy.deploy.v3/).
   2. INFRA --> Clouds -->  find your cloud --> Namespaces --> find your namespace.
   3. If the namespace is **not presented** then:
      1. Click `Create` button and specify: namespace and credentials. 
      2. Click `Save`.
      3. Go to your namespace configuration and specify the parameters for service installing.
   4. If the namespace is presented then: just check the parameters or change them.
3. To deploy service using APP-Deployer:
   1. Go to the [APP-Deployer groovy deploy page](https://cloud-deployer.netcracker.com/job/INFRA/job/groovy.deploy.v3/).
   2. Go to the `Build with Parameters` tab.
   3. Specify:
      1. Project: it is your cloud and namespace.
      2. The list is based on the information from the [APP-Deployer infrastructure configuration](https://cloud-deployer.netcracker.com/job/INFRA/job/groovy.deploy.v3/). 
      3. Deploy mode - `Clean Install` or `Rolling Update`. The `Clean Install` will clear the namespace before installation of Mongodb 
      4. Artifact descriptor version --> the **application manifest** (eg. `mongodb:master-mongo_7.0.5-20241030.001209-70`).

#### ArgoCD

The information about ArgoCD deployment can be found in [Platform ArgoCD guide](https://bass.netcracker.com/display/PLAT/ArgoCD+guide).

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

### How to Promote a Release

A promoted artifact is a finalized release that is ready for delivery to customers. Follow these steps to ensure a successful promotion:

1. **Verify Merge Requests**: Ensure that all relevant merge requests (MRs) are merged and have both `milestone` and `labels` set:
   - The **milestone** is crucial as it defines which MRs will be included in the release notes. For instance, if you set the milestone to `0.10.1` for specific MRs and create a tag `0.10.1`, details of those MRs will be listed in the release notes for the `0.10.1` release.
   - **Labels** determine the section in which an MR will appear in the release notes, such as `Documentation`, `Features`, or `Bug fix`.

2. **Review Integration Tags**: Ensure that `mongo/<VERSION>/description.yaml` includes the desired tags for all integration components.

3. **Create a Tag**: Tag the branch to be released, typically `master`.

4. **Automatic Promotion**: The CI pipeline will promote the release and automatically update the release notes.

#### Troubleshooting Promotion

- **Error: `Previous version is not set in snippet`**  
  This error indicates that the snippet containing the last promoted version is incorrect or missing. Check the snippets and validate the version to resolve this.

- **Error: `Artifacts not Promotable`**  
  This message suggests issues with artifact validation, such as Docker image or manifest validation. If new third-party artifacts are involved, request approval as necessary.

### How to Troubleshoot

There are no well-defined rules for troubleshooting, as each task is unique, but there are some tips that can do:

* Deploy parameters.
* Application manifest.
* Logs from all Mongodb pods.

Also, developer can take a look on [Troubleshooting guide](/docs/public/troubleshooting.md).

#### APP-Deployer Job Typical Errors

The errors occur during the app-deployer job are specified below.

##### Application does not exist in the CMDB

The error like "ERROR: Application does not exist in the CMDB" means that the APP-Deployer doesn't have
configuration related to the "service-name" from application manifest.

**Solution**: check that the correct manifest is used.

##### CUSTOM_RESOURCE timeout

The error like "CUSTOM_RESOURCE timeout" means the service was deployed to the namespace, but the Custom Resource doesn't have SUCCESS status.
Usually, it is related to the service state - it might be unhealthy or repeatedly crushing.

**Solution**: there is no ready answer, need to go to namespace & check service state, operator logs to find a root cause and fix it.

## CI/CD

The main CI/CD pipeline is designed to automize all basic developer routines start from code quality and finish with
deploying to k8s cluster.

- test - the stage runs go and helm unit tests.
- buildDocker - the stage builds docker image of operators.
- manifest - the stage builds Mongodb Operator manifest for current branch or release manifest using DP-Builder.
- prepare_cloud - the stage clears Vault on `ci-master` Kubernetes cluster.
- cleanDeploy - the stage runs `Clean Install` deploy to `ci-master` or `miniha` Kubernetes clusters.
- updateDeploy - the stage that runs `Clean Install` deploy to `ci-master` or `miniha` Kubernetes clusters.
- manifestValidation - the stage validates that the manifest can be promoted.
- promoteManifest70 - the stage promotes the manifest.
- releaseNotes - the stage generates Release Notes.
- milestones - the stage closes current milstone and creates new one.

## Evergreen Strategy

To keep the component up to date, the following activities should be performed regularly:

* Vulnerabilities fixing.
* Mongodb upgrade.
* Bug-fixing, improvement and feature implementation.

## Useful Links

* [Installation guide](/docs/public/installation_guide.md).
* [Troubleshooting guide](/docs/public/troubleshooting.md).
* [Architecture Guide](/docs/public/architecture.md).
* [Cloud INFRA Development process](https://bass.netcracker.com/display/PLAT/Cloud+Infra+Platform+services%3A+development+process).
* [ArgoCD User guide](https://bass.netcracker.com/display/PLAT/ArgoCD+guide).
