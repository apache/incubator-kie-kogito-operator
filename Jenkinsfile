pipeline {
    agent { label 'go'}
    options {
        buildDiscarder logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '', numToKeepStr: '6')
        timeout(time: 90, unit: 'MINUTES')
    }
    stages {
        stage('Clean Workspace') {
            steps {
                sh 'rm -rf /home/jenkins/go/src/github.com/kiegroup/kogito-cloud-operator/'
            }
        }
        stage('Clone repository') {
            steps {
              sh 'mkdir -p /home/jenkins/go/src/github.com/kiegroup/kogito-cloud-operator/'
              sh 'git clone https://github.com/kiegroup/kogito-cloud-operator.git /home/jenkins/go/src/github.com/kiegroup/kogito-cloud-operator/'
            }    
        }
        stage('Initialize') {
            steps {
                sh 'oc login --token=\$(cat /var/run/secrets/kubernetes.io/serviceaccount/token) --server=https://api.kogito.automation.rhmw.io:6443 --insecure-skip-tls-verify'
            }
        }
        stage('Build Kogito Operator') {
            steps {
                 sh """
                 export GOPATH=/home/jenkins/go
                 export GOCACHE=\${GOPATH}/.cache/go-build
                 export GOROOT=`go env GOROOT`
                 GO111MODULE=on 
                 go get -u golang.org/x/lint/golint
                 touch /etc/sub{u,g}id
                 usermod --add-subuids 10000-75535 \$(whoami)
                 usermod --add-subgids 10000-75535 \$(whoami)
                 cat /etc/subuid
                 cat /etc/subgid
                 cd /home/jenkins/go/src/github.com/kiegroup/kogito-cloud-operator/ && ./hack/go-build.sh
                 """
            }
            
        }
        stage('Build Kogito CLI') {
            steps {
                sh """
                cd /home/jenkins/go/src/github.com/kiegroup/kogito-cloud-operator/ && make build-cli
                """
            }
        }
        stage('Tag Image for Smoke Testing') {
            steps {
                  sh """
                  cd /home/jenkins/go/src/github.com/kiegroup/kogito-cloud-operator/version/ && TAG_OPERATOR=\$(grep -m 1 'Version =' version.go) && TAG_OPERATOR=\$(echo \${TAG_OPERATOR#*=} | tr -d '"')
                  podman tag quay.io/kiegroup/kogito-cloud-operator:\${TAG_OPERATOR} quay.io/kiegroup/kogito-cloud-operator-nightly:nightly-\$(echo \${GIT_COMMIT} | cut -c1-7)
                  """
            }
        }
        stage('Running Smoke Testing') {
            steps {
                  sh """
                  set +e
                  cd /home/jenkins/go/src/github.com/kiegroup/kogito-cloud-operator/ && make run-smoke-tests operator_image=quay.io/kiegroup/kogito-cloud-operator-nightly operator_tag=nightly-\$(echo \${GIT_COMMIT} | cut -c1-7) maven_mirror=http://nexus3-kogito-tools.apps.kogito.automation.rhmw.io/repository/maven-public concurrent=3 examples_ref=0.7.1 || time
                  """   
            }
        }
        
        stage('Push to Quay') {
            steps {
                withDockerRegistry([ credentialsId: "quay", url: "https://quay.io" ]) {
                  sh """
                  podman push quay.io/kiegroup/kogito-cloud-operator-nightly:nightly-\$(echo \${GIT_COMMIT} | cut -c1-7)
                  """
                }
            }
        }
    }
}
