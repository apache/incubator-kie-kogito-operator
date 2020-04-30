pipeline {
    agent { label 'go'}

    parameters {
        booleanParam(name: 'NIGHTLY_BUILD', defaultValue: false, description: 'If nightly build')
    }

    options {
        buildDiscarder logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '', numToKeepStr: '10')
        // timeout(time: 90, unit: 'MINUTES')
    }
    environment {
        WORKING_DIR = "/home/jenkins/go/src/github.com/kiegroup/kogito-cloud-operator/"
        GOPATH = "/home/jenkins/go"
        GOCACHE = "/home/jenkins/go/.cache/go-build"
    }
    stages {
        stage('Clean Workspace') {
            steps {
                dir ("${WORKING_DIR}") {
                    deleteDir()
                }
            }
        }
        stage('Initialize') {
            steps {
                sh "mkdir -p ${WORKING_DIR} && cd ${WORKSPACE} && cp -Rap * ${WORKING_DIR}"
                sh "set +x && oc login --token=\$(oc whoami -t) --server=${OPENSHIFT_API} --insecure-skip-tls-verify"
            }
        }
        stage('Build Kogito Operator') {
            steps {
                dir ("${WORKING_DIR}") {
                    sh """
                        export GOROOT=`go env GOROOT`
                        GO111MODULE=on 
                        go get -u golang.org/x/lint/golint
                        touch /etc/sub{u,g}id
                        usermod --add-subuids 10000-75535 \$(whoami)
                        usermod --add-subgids 10000-75535 \$(whoami)
                        cat /etc/subuid
                        cat /etc/subgid
                        make image_builder=podman
                    """
                }
            }
            
        }
        stage('Build Kogito CLI') {
            steps {
                dir ("${WORKING_DIR}") {
                    sh "make build-cli"
                }
            }
        }
        stage('Push Operator Image to Openshift Registry') {
            steps {
                dir ("${WORKING_DIR}") {
                    sh """
                        set +x && podman login -u jenkins -p \$(oc whoami -t) --tls-verify=false ${OPENSHIFT_REGISTRY}
                        cd version/ && TAG_OPERATOR=\$(grep -m 1 'Version =' version.go) && TAG_OPERATOR=\$(echo \${TAG_OPERATOR#*=} | tr -d '"')
                        podman tag quay.io/kiegroup/kogito-cloud-operator:\${TAG_OPERATOR} ${OPENSHIFT_REGISTRY}/openshift/kogito-cloud-operator:pr-\$(echo \${GIT_COMMIT} | cut -c1-7)
                        podman push --tls-verify=false docker://${OPENSHIFT_REGISTRY}/openshift/kogito-cloud-operator:pr-\$(echo \${GIT_COMMIT} | cut -c1-7)
                    """
                }
            }
        }
        stage('Running Testing') {
            steps {
                dir ("${WORKING_DIR}") {
                    test_params = "load_default_config=true operator_image=${OPENSHIFT_REGISTRY}/openshift/kogito-cloud-operator operator_tag=pr-\$(echo \${GIT_COMMIT} | cut -c1-7) maven_mirror=${MAVEN_MIRROR_REPOSITORY} concurrent=3"

                    if (params.NIGHTLY_BUILD) {
                        sh "make run-tests ${test_params}"
                    } else {
                        sh "make run-smoke-tests ${test_params}"
                    }
                }
            }
            post {
                always {
                    dir("${WORKING_DIR}") {
                        archiveArtifacts artifacts: 'test/logs/**/*.log', allowEmptyArchive: true
                        junit testResults: 'test/logs/**/junit.xml', allowEmptyResults: true
                    }
                }
            }
        }
        stage('Push to Quay') {
            when {
                expression {
                    return params.NIGHTLY_BUILD;
                }
            }
            steps {
                withDockerRegistry([ credentialsId: "quay", url: "https://quay.io" ]) {
                  sh """
                      podman tag ${OPENSHIFT_REGISTRY}/openshift/kogito-cloud-operator:pr-\$(echo \${GIT_COMMIT} | cut -c1-7) quay.io/kiegroup/kogito-cloud-operator-nightly:\$(echo \${GIT_COMMIT} | cut -c1-7)
                      podman push quay.io/kiegroup/kogito-cloud-operator-nightly:\$(echo \${GIT_COMMIT} | cut -c1-7)
                  """
                }
            }
        }
    }
}