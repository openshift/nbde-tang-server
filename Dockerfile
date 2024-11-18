# Build the manager binary
FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:v1.22.7-202410111609.gc451559.el9 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
ARG goarch=amd64

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY LICENSE LICENSE

# Build
RUN echo "GOARCH=${goarch}"
RUN GOFLAGS='' CGO_ENABLED=0 GOOS=linux GOARCH=${goarch} go build -a -o manager main.go

# Use well-known UBI9 micro image
FROM registry.access.redhat.com/ubi9/ubi-micro@sha256:31f00ba1d79523e182624c96e05b2f5ca66ea35d64959d84acdc8b670429415f

# Include Konflux required labels
LABEL com.redhat.component="NBDE Tang Server"
LABEL distribution-scope="public"
LABEL name="nbde-tang-server"
LABEL release="1.1.0"
LABEL version="1.1.0"
LABEL url="https://github.com/openshift/nbde-tang-server"
LABEL vendor="Red Hat, Inc."
LABEL description="The NBDE Tang Server Operator allows NBDE technology deployment on OpenShift"
LABEL io.k8s.description="The NBDE Tang Server Operator allows NBDE technology deployment on OpenShift"
LABEL summary="The NBDE Tang Server Operator allows NBDE technology deployment on OpenShift"
LABEL io.k8s.display-name="NBDE Tang Server"
LABEL io.openshift.tags="openshift,operator,nbde,network,security,storage,disk,unlocking"

WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=builder /workspace/LICENSE /licenses/
USER 65532:65532

ENTRYPOINT ["/manager"]
