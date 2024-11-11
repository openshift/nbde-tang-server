FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_9_1.22 as builder
#
# TODO: Set image to correct version
ARG IMG=registry.redhat.io/nbde-tang-server/tang-rhel9-operator@sha256:562e5f1677dbf5cd9feb00a7270e78c57b28f5177a7bf4bd2d39b2a5cd451da8
ARG ORIGINAL_IMG=quay.io/sec-eng-special/nbde-tang-server:v1.1.0
WORKDIR /code
COPY ./ ./
#
RUN echo "SNAPSHOT=${SNAPSHOT}"
RUN bash -c printenv
# Replace the bundle image in the repository with the one specified by the IMG build argument.
RUN chmod -R g+rwX ./ && find bundle/ && find bundle -type f -exec sed -i \
   "s|${ORIGINAL_IMG}|${IMG})|g" {} \+; grep -rq "${ORIGINAL_IMG}" bundle/ && \
   { echo "Failed to replace image references"; exit 1; } || echo "Image references replaced" && \
   grep -r "${IMG}" bundle/

FROM registry.access.redhat.com/ubi9/ubi-micro@sha256:7f376b75faf8ea546f28f8529c37d24adcde33dca4103f4897ae19a43d58192b

# Include required labels (for Konflux deployment)
LABEL com.redhat.component="NBDE Tang Server (Bundle)"
LABEL distribution-scope="public"
LABEL name="nbde-tang-server-bundle"
LABEL release="1.1.0"
LABEL version="1.1.0"
LABEL url="https://github.com/openshift/nbde-tang-server"
LABEL vendor="Red Hat, Inc."
LABEL description="The NBDE Tang Server Operator allows NBDE technology deployment on OpenShift"
LABEL io.k8s.description="The NBDE Tang Server Operator allows NBDE technology deployment on OpenShift"
LABEL summary="The NBDE Tang Server Operator allows NBDE technology deployment on OpenShift"
LABEL io.k8s.display-name="NBDE Tang Server"
LABEL io.openshift.tags="openshift,operator,nbde,network,security,storage,disk,unlocking"

# Core bundle labels.
LABEL operators.operatorframework.io.bundle.mediatype.v1=registry+v1
LABEL operators.operatorframework.io.bundle.manifests.v1=manifests/
LABEL operators.operatorframework.io.bundle.metadata.v1=metadata/
LABEL operators.operatorframework.io.bundle.package.v1=nbde-tang-server
LABEL operators.operatorframework.io.bundle.channels.v1=stable
LABEL operators.operatorframework.io.metrics.builder=operator-sdk-v1.37.0
LABEL operators.operatorframework.io.metrics.mediatype.v1=metrics+v1
LABEL operators.operatorframework.io.metrics.project_layout=go.kubebuilder.io/v4

# Labels for testing.
LABEL operators.operatorframework.io.test.mediatype.v1=scorecard+v1
LABEL operators.operatorframework.io.test.config.v1=tests/scorecard/

### Copy files to locations specified by labels
COPY --from=builder /code/bundle/manifests /manifests/
COPY --from=builder /code/bundle/metadata /metadata/
COPY --from=builder /code/bundle/tests/scorecard /tests/scorecard/

# Copy LICENSE to /licenses directory
COPY LICENSE /licenses/
