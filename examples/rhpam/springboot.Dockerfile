FROM registry.redhat.io/rhpam-7/rhpam-kogito-runtime-jvm-rhel8:7.12.0

ENV RUNTIME_TYPE springboot
# How to use this image with a simple example:
# 1. Clone kogito-examples repository
# 2. Build the ruleunit-springboot-example with `mvn clean package -DskipTests=true
# 3. Copy this file to the project root
# 4. Build the image: docker build --tag quay.io/<yournamespace>/ruleunit-springboot-example:latest -f springboot.Dockerfile .
#   4.1 Optionally test the image locally: docker run --rm -it -p 8080:8080 quay.io/<yournamespace>/ruleunit-springboot-example:latest
# 5. Push it: docker push quay.io/<yournamespace>/ruleunit-springboot-example:latest
# 6. Deploy it on OpenShift with the Kogito Operator, as a reference use ruleunit-quarkus-example-runtime.yaml (works for both runtimes)

# the *.jar was left to make this file project agnostic, but ideally you would need only the application binary, such as ruleunit-springboot-example.jar
COPY target/*.jar $KOGITO_HOME/bin