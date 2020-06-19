@Library('jenkins-pipeline-shared-libraries')_

def changeAuthor = env.ghprbPullAuthorLogin ?: CHANGE_AUTHOR
def changeBranch = env.ghprbSourceBranch ?: CHANGE_BRANCH
def changeTarget = env.ghprbTargetBranch ?: CHANGE_TARGET

pipeline {
    agent { label 'kogito-operator-slave && !master'}
    options {
        buildDiscarder logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '', numToKeepStr: '10')
        timeout(time: 90, unit: 'MINUTES')
    }
    tools {
            jdk 'kie-jdk11'
            maven 'kie-maven-3.6.3'
        }
    environment {
        TEMP_TAG="""pr-${sh(
                returnStdout: true,
                script: 'echo \${ghprbActualCommit} | cut -c1-7'
            ).trim()}"""
        JAVA_HOME = "${GRAALVM_HOME}"

        OPERATOR_IMAGE_NAME="kogito-cloud-operator"
        OPENSHIFT_API = credentials("OPENSHIFT_API")
        OPENSHIFT_REGISTRY = credentials("OPENSHIFT_REGISTRY")
        OPENSHIFT_INTERNAL_REGISTRY = "image-registry.openshift-image-registry.svc:5000"
        // OPENSHIFT_CREDS => Credentials to access the Openshift cluster. Use in `loginOpenshift()`
    }
    stages {
        stage('Initialize') {
            steps {
                script{
                    githubscm.checkoutIfExists('kogito-cloud-operator', changeAuthor, changeBranch, 'kiegroup', changeTarget, true)
                    // Make sure Openshift is available and can authenticate before continuing
                     loginOpenshift()
                }
            }
        }
        stage('Build Kogito Operator') {
            steps {
                sh """
                    go get -u golang.org/x/lint/golint
                    make image_builder=podman
                """
            }
        }
        stage('Build Kogito CLI') {
            steps {
                sh """
                    go get -u github.com/gobuffalo/packr/v2/packr2
                    make build-cli
                """
            }
        }
        stage('Push Operator Image to Openshift Registry') {
            steps {
                loginOpenshiftRegistry()
                sh """
                    podman tag quay.io/kiegroup/${OPERATOR_IMAGE_NAME}:${getOperatorVersion()} ${buildTempOpenshiftImageFullName()}
                    podman push --tls-verify=false ${buildTempOpenshiftImageFullName()}
                """
            }
        }
        stage('Running Smoke Testing') {
            steps {
                sh """
                        make run-smoke-tests load_factor=3 load_default_config=true operator_image=${getTempOpenshiftImageName(true)} operator_tag=${TEMP_TAG} maven_mirror=${MAVEN_MIRROR_REPOSITORY} concurrent=3
                """
            }
        }
    }
    post {
        always {
            archiveArtifacts artifacts: 'test/logs/**/*.log', allowEmptyArchive: true
            junit testResults: 'test/logs/**/junit.xml', allowEmptyResults: true
            sh "cd test && go run scripts/prune_namespaces.go"
            cleanWs()
        }
    }
}
void loginOpenshift(){
    withCredentials([usernamePassword(credentialsId: "OPENSHIFT_CREDS", usernameVariable: 'OC_USER', passwordVariable: 'OC_PWD')]){
        sh "oc login --username=${OC_USER} --password=${OC_PWD} --server=${OPENSHIFT_API} --insecure-skip-tls-verify"
    }
}
void loginOpenshiftRegistry(){
    loginOpenshift()
    sh "set +x && podman login -u jenkins -p \$(oc whoami -t) --tls-verify=false ${OPENSHIFT_REGISTRY}"
}
String getOperatorVersion(){
    return sh(script: "cd version/ && TAG_OPERATOR=\$(grep -m 1 'Version =' version.go) && TAG_OPERATOR=\$(echo \${TAG_OPERATOR#*=} | tr -d '\"') && echo \${TAG_OPERATOR}", returnStdout: true).trim()
}
String buildTempOpenshiftImageFullName(boolean internal=false){
    return "${getTempOpenshiftImageName(internal)}:${TEMP_TAG}"
}
String getTempOpenshiftImageName(boolean internal=false){
    String registry = internal ? env.OPENSHIFT_INTERNAL_REGISTRY : env.OPENSHIFT_REGISTRY
    return "${registry}/openshift/${OPERATOR_IMAGE_NAME}"
}