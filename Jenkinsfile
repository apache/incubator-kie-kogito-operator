@Library('jenkins-pipeline-shared-libraries')_

def changeAuthor = env.ghprbPullAuthorLogin ?: CHANGE_AUTHOR
def changeBranch = env.ghprbSourceBranch ?: CHANGE_BRANCH
def changeTarget = env.ghprbTargetBranch ?: CHANGE_TARGET

pipeline {
    agent { label 'operator-slave'}
    options {
        buildDiscarder logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '', numToKeepStr: '10')
        timeout(time: 90, unit: 'MINUTES')
    }
    environment {
        OPENSHIFT_INTERNAL_REGISTRY = "image-registry.openshift-image-registry.svc:5000"

        // Use buildah container engine in this pipeline
        CONTAINER_ENGINE="buildah"
    }
    stages {
        stage('Initialize') {
            steps {
               script{
                    sh ' git config --global user.email "jenkins@kie.com" '
                    sh ' git config --global user.name "kie user"'
                    githubscm.checkoutIfExists('kogito-cloud-operator', changeAuthor, changeBranch, 'kiegroup', changeTarget, true, ['token' : 'GITHUB_TOKEN', 'usernamePassword' : 'user-kie-ci10'])
                    sh "set +x && oc login --token=\$(oc whoami -t) --server=${OPENSHIFT_API} --insecure-skip-tls-verify"
               }
            }
        }
        stage('Build Kogito Operator') {
            steps {
                sh """
                    go get -u golang.org/x/lint/golint
                    usermod --add-subuids 10000-75535 \$(whoami)
                    usermod --add-subgids 10000-75535 \$(whoami)
                    make image_builder=${CONTAINER_ENGINE}
                """
            }
        }
        stage('Build Kogito CLI') {
            steps {
                sh "make build-cli"
            }
        }
        stage('Push Operator Image to Openshift Registry') {
            steps {
                sh """
                    set +x && ${CONTAINER_ENGINE} login -u jenkins -p \$(oc whoami -t) --tls-verify=false ${OPENSHIFT_REGISTRY}
                    cd version/ && TAG_OPERATOR=\$(grep -m 1 'Version =' version.go) && TAG_OPERATOR=\$(echo \${TAG_OPERATOR#*=} | tr -d '"')
                    ${CONTAINER_ENGINE} tag quay.io/kiegroup/kogito-cloud-operator:\${TAG_OPERATOR} ${OPENSHIFT_REGISTRY}/openshift/kogito-cloud-operator:pr-\$(echo \${GIT_COMMIT} | cut -c1-7)
                    ${CONTAINER_ENGINE} push --tls-verify=false docker://${OPENSHIFT_REGISTRY}/openshift/kogito-cloud-operator:pr-\$(echo \${GIT_COMMIT} | cut -c1-7)
                """
            }
        }
        stage("Build examples' images for testing"){
            steps {
                // Do not build native images for the PR checks
                sh "make build-examples-images tags='~@native' concurrent=1 ${getBDDParameters('never', false)}"
            }
            post {
                always {
                    archiveArtifacts artifacts: 'test/logs/**/*.log', allowEmptyArchive: true
                    junit testResults: 'test/logs/**/junit.xml', allowEmptyResults: true
                }
            }
        }
        stage('Running Smoke Testing') {
            steps {
                sh """
                    make run-smoke-tests concurrent=3 ${getBDDParameters('always', true)}
                """
            }
            post {
                always {
                    archiveArtifacts artifacts: 'test/logs/**/*.log', allowEmptyArchive: true
                     junit testResults: 'test/logs/**/junit.xml', allowEmptyResults: true
                      sh "cd test && go run scripts/prune_namespaces.go"
                }
            }
        }
    }
    post {
        always {
            cleanWs()
        }
    }
}

String getBDDParameters(String image_cache_mode, boolean runtime_app_registry_internal=false) {
    testParamsMap = [:]

    testParamsMap["load_default_config"] = true
    testParamsMap["ci"] = "jenkins"
    testParamsMap["load_factor"] = 3

    testParamsMap["operator_image"] = "${OPENSHIFT_REGISTRY}/openshift/kogito-cloud-operator"
    testParamsMap["operator_tag"] = "pr-\$(echo \${GIT_COMMIT} | cut -c1-7)"
    testParamsMap["maven_mirror"] = env.MAVEN_MIRROR_REPOSITORY
    
    // runtime_application_image are built in this pipeline so we can just use Openshift registry for them
    testParamsMap["image_cache_mode"] = image_cache_mode
    testParamsMap["runtime_application_image_registry"] = runtime_app_registry_internal ? env.OPENSHIFT_INTERNAL_REGISTRY : env.OPENSHIFT_REGISTRY
    testParamsMap["runtime_application_image_namespace"] = "openshift"
    testParamsMap["runtime_application_image_version"] = "pr-\$(echo \${GIT_COMMIT} | cut -c1-7)"
    
    testParamsMap['container_engine'] = env.CONTAINER_ENGINE

    String testParams = testParamsMap.collect{ entry -> "${entry.getKey()}=\"${entry.getValue()}\"" }.join(" ")
    echo "BDD parameters = ${testParams}"
    return testParams
}