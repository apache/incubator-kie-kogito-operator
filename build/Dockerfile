FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

# install operator binary
COPY build/_output/bin/kogito-cloud-operator /usr/local/bin/kogito-cloud-operator

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER 1001
