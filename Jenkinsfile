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
                    make image_builder=buildah
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
                    set +x && buildah login -u jenkins -p \$(oc whoami -t) --tls-verify=false ${OPENSHIFT_REGISTRY}
                    cd version/ && TAG_OPERATOR=\$(grep -m 1 'Version =' version.go) && TAG_OPERATOR=\$(echo \${TAG_OPERATOR#*=} | tr -d '"')
                    buildah tag quay.io/kiegroup/kogito-cloud-operator:\${TAG_OPERATOR} ${OPENSHIFT_REGISTRY}/openshift/kogito-cloud-operator:pr-\$(echo \${GIT_COMMIT} | cut -c1-7)
                    buildah push --tls-verify=false docker://${OPENSHIFT_REGISTRY}/openshift/kogito-cloud-operator:pr-\$(echo \${GIT_COMMIT} | cut -c1-7)
                """
            }
        }
        stage('Running Smoke Testing') {
            steps {
                sh """
                     make run-smoke-tests load_factor=3 load_default_config=true operator_image=${OPENSHIFT_REGISTRY}/openshift/kogito-cloud-operator operator_tag=pr-\$(echo \${GIT_COMMIT} | cut -c1-7) maven_mirror=${MAVEN_MIRROR_REPOSITORY} concurrent=3
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