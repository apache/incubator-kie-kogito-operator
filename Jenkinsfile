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
        CONTAINER_ENGINE="podman"
    }
    stages {
        stage('Initialize') {
            steps {
               script{
                    sh ' git config --global user.email "jenkins@kie.com" '
                    sh ' git config --global user.name "kie user"'
                    githubscm.checkoutIfExists('kogito-operator', changeAuthor, changeBranch, 'kiegroup', changeTarget, true, ['token' : 'GITHUB_TOKEN', 'usernamePassword' : 'user-kie-ci10'])
                    sh "set +x && oc login --token=\$(oc whoami -t) --server=${OPENSHIFT_API} --insecure-skip-tls-verify"
                    sh """
                        usermod --add-subuids 10000-75535 \$(whoami)
                        usermod --add-subgids 10000-75535 \$(whoami)
                    """
               }
            }
        }
        stage('Test Kogito Operator & CLI') {
            steps {
                sh 'make test'
            }
        }
        stage('Build Kogito Operator') {
            steps {
                sh "make BUILDER=${CONTAINER_ENGINE}"
            }
        }
        stage('Build Kogito CLI') {
            steps {
                sh 'make build-cli'
            }
        }
        stage('Push Operator Image to Openshift Registry') {
            steps {
                sh """
                    set +x && ${CONTAINER_ENGINE} login -u jenkins -p \$(oc whoami -t) --tls-verify=false ${OPENSHIFT_REGISTRY}
                    cd version/ && TAG_OPERATOR=\$(grep -m 1 'Version =' version.go) && TAG_OPERATOR=\$(echo \${TAG_OPERATOR#*=} | tr -d '"')
                    ${CONTAINER_ENGINE} tag quay.io/kiegroup/kogito-operator:\${TAG_OPERATOR} ${OPENSHIFT_REGISTRY}/openshift/kogito-operator:pr-\$(echo \${GIT_COMMIT} | cut -c1-7)
                    ${CONTAINER_ENGINE} push --tls-verify=false ${OPENSHIFT_REGISTRY}/openshift/kogito-operator:pr-\$(echo \${GIT_COMMIT} | cut -c1-7)
                """
            }
        }

        stage('Run BDD tests') {
            options {
                lock("BDD tests ${OPENSHIFT_API}")
            }
            stages {
                stage('Running smoke tests') {
                    steps {
                        // Run just smoke tests to verify basic operator functionality
                        sh """
                            make run-smoke-tests concurrent=5 ${getBDDParameters()}
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
        }
    }
    post {
        always {
            cleanWs()
        }
    }
}

String getBDDParameters() {
    testParamsMap = [:]

    testParamsMap["load_default_config"] = true
    testParamsMap["ci"] = "jenkins"
    testParamsMap["load_factor"] = 3
    testParamsMap['disable_maven_native_build_container'] = true

    testParamsMap["operator_image"] = "${OPENSHIFT_REGISTRY}/openshift/kogito-operator"
    testParamsMap["operator_tag"] = "pr-\$(echo \${GIT_COMMIT} | cut -c1-7)"
    
    if(env.MAVEN_MIRROR_REPOSITORY){
        testParamsMap["maven_mirror"] = env.MAVEN_MIRROR_REPOSITORY
        testParamsMap["maven_ignore_self_signed_certificate"] = true
    }
    
    // Reuse runtime application images from nightly builds
    testParamsMap["image_cache_mode"] = "always"
    testParamsMap["runtime_application_image_registry"] = "quay.io"
    testParamsMap["runtime_application_image_namespace"] = "kiegroup"
    
    testParamsMap['container_engine'] = env.CONTAINER_ENGINE

    String testParams = testParamsMap.collect{ entry -> "${entry.getKey()}=\"${entry.getValue()}\"" }.join(" ")
    echo "BDD parameters = ${testParams}"
    return testParams
}