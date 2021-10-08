FROM quay.io/kiegroup/kogito-runtime-jvm:latest

ENV RUNTIME_TYPE springboot
# How to use this image with a simple example:
# 1. Clone kogito-examples repository
# 2. Build the process-springboot-example with `mvn clean package -DskipTests=true
# 3. Copy this file to the project root
# 4. Build the image: docker build --tag quay.io/<yournamespace>/process-springboot-example:latest -f springboot.Dockerfile .
#   4.1 Optionally test the image locally: docker run --rm -it -p 8080:8080 quay.io/<yournamespace>/process-springboot-example:latest
# 5. Push it: docker push quay.io/<yournamespace>/process-springboot-example:latest
# 6. Deploy it on Kubernetes with the Kogito Operator, as a reference use process-quarkus-example-runtime.yaml (works for both runtimes)

# you must replace the *.jar with the application binary, such as process-springboot-example.jar
COPY target/*.jar $KOGITO_HOME/bin
