/*
* This file is describing all the Jenkins jobs in the DSL format (see https://plugins.jenkins.io/job-dsl/)
* needed by the Kogito pipelines.
*
* The main part of Jenkins job generation is defined into the https://github.com/apache/incubator-kie-kogito-pipelines repository.
*
* This file is making use of shared libraries defined in
* https://github.com/apache/incubator-kie-kogito-pipelines/tree/main/dsl/seed/src/main/groovy/org/kie/jenkins/jobdsl.
*/

import org.kie.jenkins.jobdsl.model.JobType
import org.kie.jenkins.jobdsl.utils.JobParamsUtils
import org.kie.jenkins.jobdsl.KogitoJobTemplate
import org.kie.jenkins.jobdsl.KogitoJobUtils
import org.kie.jenkins.jobdsl.Utils

jenkins_path = '.ci/jenkins'

// Init branch
setupExamplesImagesDeployJob(JobType.SETUP_BRANCH, 'kogito-examples-images', [ JOB_ID: 'Setup branch' ])
createSetupBranchJob()

// Nightly jobs
setupProfilingJob()
setupDeployJob(JobType.NIGHTLY)
setupExamplesImagesDeployJob(JobType.NIGHTLY)

// Release jobs
setupDeployJob(JobType.RELEASE)
setupPromoteJob(JobType.RELEASE)
setupExamplesImagesDeployJob(JobType.RELEASE)
setupExamplesImagesPromoteJob(JobType.RELEASE)

/////////////////////////////////////////////////////////////////
// Methods
/////////////////////////////////////////////////////////////////

void setupProfilingJob() {
    def jobParams = JobParamsUtils.getBasicJobParamsWithEnv(this, 'kogito-operator-profiling', JobType.NIGHTLY, 'sonarcloud', "${jenkins_path}/Jenkinsfile.profiling", 'Kogito Cloud Operator Profiling')
    JobParamsUtils.setupJobParamsAgentDockerBuilderImageConfiguration(this, jobParams)
    jobParams.triggers = [ cron : '@midnight' ]
    jobParams.env.putAll([
        JENKINS_EMAIL_CREDS_ID: "${JENKINS_EMAIL_CREDS_ID}",

        OPERATOR_IMAGE_NAME: 'kogito-operator-profiling',
        MAX_REGISTRY_RETRIES: 3,
        OPENSHIFT_API_KEY: 'OPENSHIFT_API',
        OPENSHIFT_CREDS_KEY: 'OPENSHIFT_CREDS',

        GIT_AUTHOR: "${GIT_AUTHOR_NAME}",
        MAVEN_ARTIFACT_REPOSITORY: "${MAVEN_ARTIFACTS_REPOSITORY}",
    ])
    KogitoJobTemplate.createPipelineJob(this, jobParams)?.with {
        parameters {
            stringParam('BUILD_BRANCH_NAME', "${GIT_BRANCH}", 'Set the Git branch to checkout')
        }
    }
}

void createSetupBranchJob() {
    def jobParams = JobParamsUtils.getBasicJobParams(this, 'kogito-operator', JobType.SETUP_BRANCH, "${jenkins_path}/Jenkinsfile.setup-branch", 'Kogito Cloud Operator Init Branch')
    JobParamsUtils.setupJobParamsAgentDockerBuilderImageConfiguration(this, jobParams)
    jobParams.env.putAll([
        JENKINS_EMAIL_CREDS_ID: "${JENKINS_EMAIL_CREDS_ID}",
        GIT_AUTHOR: "${GIT_AUTHOR_NAME}",
        AUTHOR_CREDS_ID: "${GIT_AUTHOR_CREDENTIALS_ID}",
        GITHUB_TOKEN_CREDS_ID: "${GIT_AUTHOR_TOKEN_CREDENTIALS_ID}",

        IS_MAIN_BRANCH: "${Utils.isMainBranch(this)}"
    ])
    KogitoJobTemplate.createPipelineJob(this, jobParams)?.with {
        parameters {
            stringParam('DISPLAY_NAME', '', 'Setup a specific build display name')

            stringParam('BUILD_BRANCH_NAME', "${GIT_BRANCH}", 'Set the Git branch to checkout')

            stringParam('PROJECT_VERSION', '', 'Version to set.')

            booleanParam('SEND_NOTIFICATION', false, 'In case you want the pipeline to send a notification on CI channel for this run.')
        }
    }
}

void setupDeployJob(JobType jobType) {
    def jobParams = JobParamsUtils.getBasicJobParams(this, 'kogito-operator-deploy', jobType, "${jenkins_path}/Jenkinsfile.deploy", 'Kogito Cloud Operator Deploy')
    JobParamsUtils.setupJobParamsAgentDockerBuilderImageConfiguration(this, jobParams)
    jobParams.env.putAll([
        JENKINS_EMAIL_CREDS_ID: "${JENKINS_EMAIL_CREDS_ID}",

        OPERATOR_IMAGE_NAME: 'kogito-operator',
        MAX_REGISTRY_RETRIES: 3,
        OPENSHIFT_API_KEY: 'OPENSHIFT_API',
        OPENSHIFT_CREDS_KEY: 'OPENSHIFT_CREDS',
        PROPERTIES_FILE_NAME: 'deployment.properties',

        GIT_AUTHOR: "${GIT_AUTHOR_NAME}",

        AUTHOR_CREDS_ID: "${GIT_AUTHOR_CREDENTIALS_ID}",
        GITHUB_TOKEN_CREDS_ID: "${GIT_AUTHOR_TOKEN_CREDENTIALS_ID}",

        DEFAULT_STAGING_REPOSITORY: "${MAVEN_NEXUS_STAGING_PROFILE_URL}",
        MAVEN_ARTIFACT_REPOSITORY: "${MAVEN_ARTIFACTS_REPOSITORY}",
    ])
    KogitoJobTemplate.createPipelineJob(this, jobParams)?.with {
        parameters {
            stringParam('DISPLAY_NAME', '', 'Setup a specific build display name')

            stringParam('BUILD_BRANCH_NAME', "${GIT_BRANCH}", 'Set the Git branch to checkout')
            if (jobType == JobType.PULL_REQUEST) {
                // author can be changed as param only for PR behavior, due to source branch/target, else it is considered as an env
                stringParam('GIT_AUTHOR', "${GIT_AUTHOR_NAME}", 'Set the Git author to checkout')
            }

            booleanParam('CREATE_PR', false, 'Should we create a PR with the changes ?')
            stringParam('PROJECT_VERSION', '', 'Optional if not RELEASE. If RELEASE, cannot be empty.')

            // Build&Test information
            booleanParam('SKIP_TESTS', false, 'Skip tests')
            booleanParam('SMOKE_TESTS_ONLY', false, 'If only smoke tests should be run. Default is full testing.')
            booleanParam('SKIP_NATIVE_TESTS', false, 'Skip native tests')
            booleanParam('SKIP_NON_NATIVE_TESTS', false, 'Skip non native tests')
            stringParam('BDD_TEST_TAGS', '', 'Execute only a subset of BDD tests')

            stringParam('NATIVE_BUILDER_IMAGE', '', 'Force the native builder image')

            // Deploy information
            booleanParam('IMAGE_USE_OPENSHIFT_REGISTRY', false, 'Set to true if image should be deployed in Openshift registry.In this case, IMAGE_REGISTRY_CREDENTIALS, IMAGE_REGISTRY and IMAGE_NAMESPACE parameters will be ignored')
            stringParam('IMAGE_REGISTRY_CREDENTIALS', "${CLOUD_IMAGE_REGISTRY_CREDENTIALS}", 'Image registry credentials to use to deploy images. Will be ignored if no IMAGE_REGISTRY is given')
            stringParam('IMAGE_REGISTRY', "${CLOUD_IMAGE_REGISTRY}", 'Image registry to use to deploy images')
            stringParam('IMAGE_NAMESPACE', "${CLOUD_IMAGE_NAMESPACE}", 'Image namespace to use to deploy images')
            stringParam('IMAGE_NAME_SUFFIX', '', 'Image name suffix to use to deploy images. In case you need to change the final image name, you can add a suffix to it.')
            stringParam('IMAGE_TAG', '', 'Image tag to use to deploy images')
            stringParam('KOGITO_PR_BRANCH', '', 'PR branch name')
            booleanParam('DEPLOY_WITH_LATEST_TAG', false, 'Set to true if you want the deployed image to also be with the `latest` tag')
            booleanParam('SKIP_DEPLOY', false, 'In case you don\'t want to deploy the final image.')

            // Test config if needed specifics. Else test default config will apply.
            booleanParam('KOGITO_IMAGES_USE_OPENSHIFT_REGISTRY', false, 'Set to true if kogito images for tests are in internal Openshift registry.In this case, KOGITO_IMAGES_REGISTRY and KOGITO_IMAGES_NAMESPACE parameters will be ignored')
            stringParam('KOGITO_IMAGES_REGISTRY', "${CLOUD_IMAGE_REGISTRY}", 'Test images registry')
            stringParam('KOGITO_IMAGES_NAMESPACE', "${CLOUD_IMAGE_NAMESPACE}", 'Test images namespace')
            stringParam('KOGITO_IMAGES_NAME_SUFFIX', '', 'Test images name suffix')
            stringParam('KOGITO_IMAGES_TAG', '', 'Test images tag')

            stringParam('EXAMPLES_URI', '', 'Git uri to the kogito-examples repository to use for tests.')
            stringParam('EXAMPLES_REF', '', 'Git reference (branch/tag) to the kogito-examples repository to use for tests.')

            stringParam('EXAMPLES_IMAGES_CACHE_MODE', 'never', 'Set the examples images\' cache mode for the BDD tests. Default it will always build the image.')
            booleanParam('EXAMPLES_IMAGES_USE_OPENSHIFT_REGISTRY', false, 'Set to true if examples images for tests are in internal Openshift registry.In this case, EXAMPLES_IMAGES_REGISTRY and EXAMPLES_IMAGES_NAMESPACE parameters will be ignored')
            stringParam('EXAMPLES_IMAGES_REGISTRY', "${CLOUD_IMAGE_REGISTRY}", 'Examples images registry')
            stringParam('EXAMPLES_IMAGES_NAMESPACE', "${CLOUD_IMAGE_NAMESPACE}", 'Examples images namespace')
            stringParam('EXAMPLES_IMAGES_NAME_PREFIX', '', 'Examples images name prefix')
            stringParam('EXAMPLES_IMAGES_NAME_SUFFIX', '', 'Examples images name suffix')
            stringParam('EXAMPLES_IMAGES_TAG', '', 'Examples images tag')

            booleanParam('SEND_NOTIFICATION', false, 'In case you want the pipeline to send a notification on CI channel for this run.')
        }
    }
}

void setupPromoteJob(JobType jobType) {
    def jobParams = JobParamsUtils.getBasicJobParams(this, 'kogito-operator-promote', jobType, "${jenkins_path}/Jenkinsfile.promote", 'Kogito Cloud Operator Promote')
    JobParamsUtils.setupJobParamsAgentDockerBuilderImageConfiguration(this, jobParams)
    jobParams.env.putAll([
        JENKINS_EMAIL_CREDS_ID: "${JENKINS_EMAIL_CREDS_ID}",

        MAX_REGISTRY_RETRIES: 3,
        OPENSHIFT_API_KEY: 'OPENSHIFT_API',
        OPENSHIFT_CREDS_KEY: 'OPENSHIFT_CREDS',
        PROPERTIES_FILE_NAME: 'deployment.properties',

        GIT_AUTHOR: "${GIT_AUTHOR_NAME}",

        AUTHOR_CREDS_ID: "${GIT_AUTHOR_CREDENTIALS_ID}",
        GITHUB_TOKEN_CREDS_ID: "${GIT_AUTHOR_TOKEN_CREDENTIALS_ID}",
    ])
    KogitoJobTemplate.createPipelineJob(this, jobParams)?.with {
        parameters {
            stringParam('DISPLAY_NAME', '', 'Setup a specific build display name')

            stringParam('BUILD_BRANCH_NAME', "${GIT_BRANCH}", 'Set the Git branch to checkout')

            // Deploy job url to retrieve deployment.properties
            stringParam('DEPLOY_BUILD_URL', '', 'URL to jenkins deploy build to retrieve the `deployment.properties` file. If base parameters are defined, they will override the `deployment.properties` information')

            // Base information which can override `deployment.properties`
            booleanParam('BASE_IMAGE_USE_OPENSHIFT_REGISTRY', false, 'Override `deployment.properties`. Set to true if base image should be deployed in Openshift registry.In this case, BASE_IMAGE_REGISTRY_CREDENTIALS, BASE_IMAGE_REGISTRY and BASE_IMAGE_NAMESPACE parameters will be ignored')
            stringParam('BASE_IMAGE_REGISTRY_CREDENTIALS', "${CLOUD_IMAGE_REGISTRY_CREDENTIALS}", 'Override `deployment.properties`. Base Image registry credentials to use to deploy images. Will be ignored if no BASE_IMAGE_REGISTRY is given')
            stringParam('BASE_IMAGE_REGISTRY', "${CLOUD_IMAGE_REGISTRY}", 'Override `deployment.properties`. Base image registry')
            stringParam('BASE_IMAGE_NAMESPACE', "${CLOUD_IMAGE_NAMESPACE}", 'Override `deployment.properties`. Base image namespace')
            stringParam('BASE_IMAGE_NAME_SUFFIX', '', 'Override `deployment.properties`. Base image name suffix')
            stringParam('BASE_IMAGE_TAG', '', 'Override `deployment.properties`. Base image tag')

            // Promote information
            booleanParam('PROMOTE_IMAGE_USE_OPENSHIFT_REGISTRY', false, 'Set to true if base image should be deployed in Openshift registry.In this case, PROMOTE_IMAGE_REGISTRY_CREDENTIALS, PROMOTE_IMAGE_REGISTRY and PROMOTE_IMAGE_NAMESPACE parameters will be ignored')
            stringParam('PROMOTE_IMAGE_REGISTRY_CREDENTIALS', "${CLOUD_IMAGE_REGISTRY_CREDENTIALS}", 'Promote Image registry credentials to use to deploy images. Will be ignored if no PROMOTE_IMAGE_REGISTRY is given')
            stringParam('PROMOTE_IMAGE_REGISTRY', "${CLOUD_IMAGE_REGISTRY}", 'Promote image registry')
            stringParam('PROMOTE_IMAGE_NAMESPACE', "${CLOUD_IMAGE_NAMESPACE}", 'Promote image namespace')
            stringParam('PROMOTE_IMAGE_NAME_SUFFIX', '', 'Promote image name suffix')
            stringParam('PROMOTE_IMAGE_TAG', '', 'Promote image tag')
            booleanParam('DEPLOY_WITH_LATEST_TAG', false, 'Set to true if you want the deployed images to also be with the `latest` tag')

            // Release information which can override  `deployment.properties`
            stringParam('PROJECT_VERSION', '', 'Override `deployment.properties`. Optional if not RELEASE. If RELEASE, cannot be empty.')
            stringParam('GIT_TAG', '', 'Git tag to set, if different from v{KOGITO_VERSION}')

            booleanParam('SEND_NOTIFICATION', false, 'In case you want the pipeline to send a notification on CI channel for this run.')
        }
    }
}

void setupExamplesImagesDeployJob(JobType jobType, String jobName = 'kogito-examples-images-deploy', Map extraEnv = [:]) {
    def jobParams = JobParamsUtils.getBasicJobParams(this, jobName, jobType, "${jenkins_path}/Jenkinsfile.examples-images.deploy", 'Kogito Examples Images Deploy')
    JobParamsUtils.setupJobParamsAgentDockerBuilderImageConfiguration(this, jobParams)
    jobParams.env.putAll(extraEnv)
    if (jobType == JobType.PULL_REQUEST) {
        jobParams.git.branch = '${BUILD_BRANCH_NAME}'
        jobParams.git.author = '${GIT_AUTHOR}'
        jobParams.git.project_url = Utils.createProjectUrl("${GIT_AUTHOR_NAME}", jobParams.git.repository)
    }
    jobParams.env.putAll([
        JENKINS_EMAIL_CREDS_ID: "${JENKINS_EMAIL_CREDS_ID}",

        MAX_REGISTRY_RETRIES: 3,
        OPENSHIFT_API_KEY: 'OPENSHIFT_API',
        OPENSHIFT_CREDS_KEY: 'OPENSHIFT_CREDS',
        PROPERTIES_FILE_NAME: 'deployment.properties',
    ])
    if (jobType == JobType.PULL_REQUEST) {
        jobParams.env.putAll([
            MAVEN_ARTIFACT_REPOSITORY: "${MAVEN_PR_CHECKS_REPOSITORY_URL}",
        ])
    } else {
        jobParams.env.putAll([
            GIT_AUTHOR: "${GIT_AUTHOR_NAME}",

            DEFAULT_STAGING_REPOSITORY: "${MAVEN_NEXUS_STAGING_PROFILE_URL}",
            MAVEN_ARTIFACT_REPOSITORY: "${MAVEN_ARTIFACTS_REPOSITORY}",
        ])
    }
    KogitoJobTemplate.createPipelineJob(this, jobParams)?.with {
        parameters {
            stringParam('DISPLAY_NAME', '', 'Setup a specific build display name')

            stringParam('BUILD_BRANCH_NAME', "${GIT_BRANCH}", 'Set the Git branch to checkout')
            if (jobType == JobType.PULL_REQUEST) {
                // author can be changed as param only for PR behavior, due to source branch/target, else it is considered as an env
                stringParam('GIT_AUTHOR', "${GIT_AUTHOR_NAME}", 'Set the Git author to checkout')
            }

            // Build&Test information
            booleanParam('SKIP_TESTS', false, 'Skip tests')
            booleanParam('SMOKE_TESTS_ONLY', false, 'If only smoke tests should be run. Default is full testing.')
            booleanParam('SKIP_NATIVE_TESTS', false, 'Skip native tests')
            booleanParam('SKIP_NON_NATIVE_TESTS', false, 'Skip non native tests')
            stringParam('BDD_TEST_TAGS', '', 'Execute only a subset of BDD tests')

            stringParam('NATIVE_BUILDER_IMAGE', '', 'Force the native builder image')

            // Deploy information
            booleanParam('IMAGE_USE_OPENSHIFT_REGISTRY', false, 'Set to true if image should be deployed in Openshift registry.In this case, IMAGE_REGISTRY_CREDENTIALS, IMAGE_REGISTRY and IMAGE_NAMESPACE parameters will be ignored')
            stringParam('IMAGE_REGISTRY_CREDENTIALS', "${CLOUD_IMAGE_REGISTRY_CREDENTIALS}", 'Image registry credentials to use to deploy images. Will be ignored if no IMAGE_REGISTRY is given')
            stringParam('IMAGE_REGISTRY', "${CLOUD_IMAGE_REGISTRY}", 'Image registry to use to deploy images')
            stringParam('IMAGE_NAMESPACE', "${CLOUD_IMAGE_NAMESPACE}", 'Image namespace to use to deploy images')
            stringParam('IMAGE_NAME_PREFIX', '', 'Image name prefix to use to deploy images. In case you need to change the final image name, you can add a prefix to it.')
            stringParam('IMAGE_NAME_SUFFIX', '', 'Image name suffix to use to deploy images. In case you need to change the final image name, you can add a suffix to it.')
            stringParam('IMAGE_TAG', '', 'Image tag to use to deploy images')
            booleanParam('DEPLOY_WITH_LATEST_TAG', false, 'Set to true if you want the deployed image to also be with the `latest` tag')

            // Test config if needed specifics. Else test default config will apply.
            booleanParam('KOGITO_IMAGES_USE_OPENSHIFT_REGISTRY', false, 'Set to true if kogito images for tests are in internal Openshift registry.In this case, KOGITO_IMAGES_REGISTRY and KOGITO_IMAGES_NAMESPACE parameters will be ignored')
            stringParam('KOGITO_IMAGES_REGISTRY', "${CLOUD_IMAGE_REGISTRY}", 'Test images registry')
            stringParam('KOGITO_IMAGES_NAMESPACE', "${CLOUD_IMAGE_NAMESPACE}", 'Test images namespace')
            stringParam('KOGITO_IMAGES_NAME_SUFFIX', '', 'Test images name suffix')
            stringParam('KOGITO_IMAGES_TAG', '', 'Test images tag')

            booleanParam('SEND_NOTIFICATION', false, 'In case you want the pipeline to send a notification on CI channel for this run.')
        }
    }
}

void setupExamplesImagesPromoteJob(JobType jobType) {
    def jobParams = JobParamsUtils.getBasicJobParams(this, 'kogito-examples-images-promote', jobType, "${jenkins_path}/Jenkinsfile.examples-images.promote", 'Kogito Examples Images Promote')
    JobParamsUtils.setupJobParamsAgentDockerBuilderImageConfiguration(this, jobParams)
    jobParams.env.putAll([
        JENKINS_EMAIL_CREDS_ID: "${JENKINS_EMAIL_CREDS_ID}",

        MAX_REGISTRY_RETRIES: 3,
        OPENSHIFT_API_KEY: 'OPENSHIFT_API',
        OPENSHIFT_CREDS_KEY: 'OPENSHIFT_CREDS',
        PROPERTIES_FILE_NAME: 'deployment.properties',

        GIT_AUTHOR: "${GIT_AUTHOR_NAME}",
    ])
    KogitoJobTemplate.createPipelineJob(this, jobParams)?.with {
        parameters {
            stringParam('DISPLAY_NAME', '', 'Setup a specific build display name')

            stringParam('BUILD_BRANCH_NAME', "${GIT_BRANCH}", 'Set the Git branch to checkout')

            // Deploy job url to retrieve deployment.properties
            stringParam('DEPLOY_BUILD_URL', '', 'URL to jenkins deploy build to retrieve the `deployment.properties` file. If base parameters are defined, they will override the `deployment.properties` information')

            // Base information which can override `deployment.properties`
            booleanParam('BASE_IMAGE_USE_OPENSHIFT_REGISTRY', false, 'Override `deployment.properties`. Set to true if base image should be deployed in Openshift registry.In this case, BASE_IMAGE_REGISTRY_CREDENTIALS, BASE_IMAGE_REGISTRY and BASE_IMAGE_NAMESPACE parameters will be ignored')
            stringParam('BASE_IMAGE_REGISTRY_CREDENTIALS', "${CLOUD_IMAGE_REGISTRY_CREDENTIALS}", 'Override `deployment.properties`. Base Image registry credentials to use to deploy images. Will be ignored if no BASE_IMAGE_REGISTRY is given')
            stringParam('BASE_IMAGE_REGISTRY', "${CLOUD_IMAGE_REGISTRY}", 'Override `deployment.properties`. Base image registry')
            stringParam('BASE_IMAGE_NAMESPACE', "${CLOUD_IMAGE_NAMESPACE}", 'Override `deployment.properties`. Base image namespace')
            stringParam('BASE_IMAGE_NAME_PREFIX', '', 'Override `deployment.properties`. Base image name prefix')
            stringParam('BASE_IMAGE_NAMES', '', 'Override `deployment.properties`. Comma separated list of images')
            stringParam('BASE_IMAGE_NAME_SUFFIX', '', 'Override `deployment.properties`. Base image name suffix')
            stringParam('BASE_IMAGE_TAG', '', 'Override `deployment.properties`. Base image tag')

            // Promote information
            booleanParam('PROMOTE_IMAGE_USE_OPENSHIFT_REGISTRY', false, 'Set to true if base image should be deployed in Openshift registry.In this case, PROMOTE_IMAGE_REGISTRY_CREDENTIALS, PROMOTE_IMAGE_REGISTRY and PROMOTE_IMAGE_NAMESPACE parameters will be ignored')
            stringParam('PROMOTE_IMAGE_REGISTRY_CREDENTIALS', "${CLOUD_IMAGE_REGISTRY_CREDENTIALS}", 'Promote Image registry credentials to use to deploy images. Will be ignored if no PROMOTE_IMAGE_REGISTRY is given')
            stringParam('PROMOTE_IMAGE_REGISTRY', "${CLOUD_IMAGE_REGISTRY}", 'Promote image registry')
            stringParam('PROMOTE_IMAGE_NAMESPACE', "${CLOUD_IMAGE_NAMESPACE}", 'Promote image namespace')
            stringParam('PROMOTE_IMAGE_NAME_PREFIX', '', 'Promote image name prefix')
            stringParam('PROMOTE_IMAGE_NAME_SUFFIX', '', 'Promote image name suffix')
            stringParam('PROMOTE_IMAGE_TAG', '', 'Promote image tag')
            booleanParam('DEPLOY_WITH_LATEST_TAG', false, 'Set to true if you want the deployed images to also be with the `latest` tag')

            // Release information which can override  `deployment.properties`
            stringParam('PROJECT_VERSION', '', 'Override `deployment.properties`. If env.RELEASE, cannot be empty.')

            booleanParam('SEND_NOTIFICATION', false, 'In case you want the pipeline to send a notification on CI channel for this run.')
        }
    }
}
