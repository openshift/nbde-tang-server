FROM golang:1.25.2 as builder

# Set original rbac image
ARG IMG=gcr.io/kubebuilder/kube-rbac-proxy:v0.16.0
ARG ORIGINAL_IMG=registry.redhat.io/openshift4/ose-kube-rbac-proxy-rhel9@sha256:c15928ecc2f0ff370f567a2ca909b7a71e2becaff0af649275396e59232b7b8c
WORKDIR /code
COPY ./ ./

# Replace the bundle image in the repository with the one specified by the IMG build argument.
RUN chmod -R g+rwX ./ && find bundle/ && find bundle -type f -exec sed -i \
   "s|${ORIGINAL_IMG}|${IMG}|g" {} \+; grep -rq "${ORIGINAL_IMG}" bundle/ && \
   { echo "Failed to replace image references"; exit 1; } || echo "Image references replaced" && \
   grep -r "${IMG}" bundle/


FROM scratch

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

# Copy files to locations specified by labels.
COPY --from=builder /code/bundle/manifests /manifests/
COPY --from=builder /code/bundle/metadata /metadata/
COPY --from=builder /code/bundle/tests/scorecard /tests/scorecard/
