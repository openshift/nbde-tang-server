FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_9_1.22 as builder
#
# TODO: Set image to correct version
# ARG IMG=registry.redhat.io/nbde-tang-server/tang-rhel9-operator@sha256:562e5f1677dbf5cd9feb00a7270e78c57b28f5177a7bf4bd2d39b2a5cd451da8
ARG IMG=quay.io/redhat-user-workloads/konflux-sec-eng-spec-tenant/nbde-tang-server-multiarch:9ed11bd3865e986ff30eebeabd38934b07e50e53
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

FROM registry.access.redhat.com/ubi9/ubi-micro@sha256:ac53e091e8bcaf2e33877dc26fd118aea7f31ba82b7c6f083e39ad9cea6691fc

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

# Include referenced image so that it can be easily verified in the bundle
LABEL konflux.referenced.image="quay.io/redhat-user-workloads/konflux-sec-eng-spec-tenant/nbde-tang-server-multiarch@sha256:1c5be3b432e968c8ff495a30ef0bf730aefd5c508e75299f7497afce182894f9"

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
