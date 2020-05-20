FROM quay.io/kiegroup/kogito-quarkus-jvm-ubi8:latest

# How to use this image with a simple example:
# 1. Clone kogito-examples repository
# 2. Build the process-quarkus-example with `mvn clean package -DskipTests=true
# 3. Copy this file to the project root
# 4. Build the image: docker build --tag quay.io/<yournamespace>/process-quarkus-example:latest -f quarkus-jvm.Dockerfile .
#   4.1 Optionally test the image locally: docker run --rm -it -p 8080:8080 quay.io/<yournamespace>/process-quarkus-example:latest
# 5. Push it: docker push quay.io/<yournamespace>/process-quarkus-example:latest
# 6. Deploy it on Kubernetes with the kogito operator, as a reference use process-quarkus-example-runtime.yaml

COPY target/*-runner.jar $KOGITO_HOME/bin
COPY target/lib $KOGITO_HOME/bin/lib
COPY target/classes/persistence/ $KOGITO_HOME/data/protobufs