FROM quay.io/kiegroup/kogito-runtime-jvm:latest

ENV RUNTIME_TYPE quarkus
# How to use this image with a simple example:
# 1. Clone kogito-examples repository
# 2. Build the process-quarkus-example with `mvn clean package -DskipTests=true
# 3. Copy this file to the project root
# 4. Build the image: docker build --tag quay.io/<yournamespace>/process-quarkus-example:latest -f quarkus-jvm.Dockerfile .
#   4.1 Optionally test the image locally: docker run --rm -it -p 8080:8080 quay.io/<yournamespace>/process-quarkus-example:latest
# 5. Push it: docker push quay.io/<yournamespace>/process-quarkus-example:latest
# 6. Deploy it on Openshift with the kogito operator, as a reference use process-quarkus-example-runtime.yaml

COPY target/quarkus-app/lib/ $KOGITO_HOME/bin/lib/
COPY target/quarkus-app/*.jar $KOGITO_HOME/bin
COPY target/quarkus-app/app/ $KOGITO_HOME/bin/app/
COPY target/quarkus-app/quarkus/ $KOGITO_HOME/bin/quarkus/

# For the legacy quarkus application jar use the commands below
# COPY target/*-runner.jar $KOGITO_HOME/bin
# COPY target/lib $KOGITO_HOME/bin/lib
