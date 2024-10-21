# nbde-tang-server

## License

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Contents

- [Introduction](#introduction)
- [Versions](#versions)
- [Installation](#installation)
- [Compilation](#compilation)
- [Cross Compilation](#cross-compilation)
- [Cleanup](#cleanup)
- [Tests](#tests)
- [Function Tests](#function-tests)
- [CI/CD](#cicd)
- [Scorecard](#scorecard)
- [Links](#links)

## Introduction

The NBDE Tang server operator helps to provide [NBDE](https://access.redhat.com/articles/6987053)
technology for OpenShift/K8S. It deploys one or several Tang servers automatically.
The Tang server container image to launch is configurable, and will use the latest one
available by default. This operator has been developed using [operator-sdk](https://sdk.operatorframework.io/).

The operator avoids having to follow all Tang manual installation steps,
and leverages some of the features provided by OpenShift: multi-replica deployment,
scale-in/out, scale up/down or traffic load balancing.

This operator also allows automation of certain operations, which are error-prone
if executed manually. Examples of these operations are:
- server deployment and configuration
- key rotation
- hidden keys deletion

Up to date, it can be deployed as a CRD, containing its proper
configuration values to perform appropriate Tang server operations.

An introductory video can be seen in next link:
[NBDE in OpenShift: tang-operator basics](https://youtu.be/hmMSIkBoGoY)

## Versions

Versions released up to date of the operator and the operator-bundle are:

- v0.0.1:  Hello world version
- v0.0.2:  Basic version with no fields still updated
- v0.0.3:  First release correct version
- v0.0.4:  Version that fixes issues with deployments/pods/services permissions
- v0.0.5:  Version that publishes the service and exposes it on configurable port
- v0.0.6:  Types refactoring. Initial ginkgo based test
- v0.0.7:  Include finalizers to make deletion quicker
- v0.0.8:  Operator metadata homogenization
- v0.0.9:  Operator shared storage
- v0.0.10: Code Refactoring
- v0.0.11: Extend tests
- v0.0.12: Fix default key path
- v0.0.13: Add type for Persistent Volume Claim attach
- v0.0.14: Fix issue on non 8080 service port deployment
- v0.0.15: Add resource request/limits
- v0.0.16: Fix scale up issues
- v0.0.17: Key rotation/deletion management through spec file
- v0.0.18: Advertise only signing keys
- v0.0.19: Add Events writing with important information
- v0.0.20: Use tangd-healthcheck only for aliveness
- v0.0.21: Use tangd-healthcheck for aliveness and readiness, separating intervals
- v0.0.22: Remove personal accounts and use organization ones
- v0.0.23: Selective hidden keys deletion
- v0.0.24: Execute Tang container pod as non-root user
- v0.0.25: Allow key handling without cluster role configuration
- v0.0.26: Use RHEL9 Tang container version
- v0.0.27: Update operator-sdk and supported Go version (1.19.6 and higher)
- v0.0.28: Code refactor
-  v1.0.0: GA release candidate
-  v1.0.1: new GA release candidate. Update Go version (1.16->1.21) and dependencies
-  v1.0.2: new GA release candidate. Fix CVE-2023-39325
-  v1.0.3: new GA release candidate. Update all mods to latest version
-  v1.0.4: new GA release candidate. Adjust to naming conventions
-  v1.0.5: new GA release candidate. Use latest `kube-rbac-proxy` version
-  v1.0.6: GA release. Fix Url to URL in CRD. GA Released Version in OpenShift
-  v1.0.7: GA re-release. Fix channel ("alpha" to "stable")
-  v1.0.8: ServiceType / ClusterIP configuration through TangServer CR
-  v1.0.9: Golang dependencies update
- v1.0.10: Libraries update
- v1.0.11: Libraries update, Update Go version (1.22 -> 1.22.5)
- v1.0.12: Fix issues reported by gosec tool
- v1.0.13: Libraries update, Update Go version (1.22.5 -> 1.23.2)
- v1.0.14: Rename tang-operator to nbde-tang-server. Use nbde.openshift.io instead of daemons.redhat.com domain

## Installation

In order to install this operator, you must have previously installed
an OpenShift/K8S cluster. For small computers, **CRC** (Code Ready Containers)
project is recommended. In case normal OpenShift cluster is used,
installation should not differ from the CRC one. Ultimately, the operator has
been included in Red Hat OpenShift catalog, named as
[NBDE Tang Server](https://catalog.redhat.com/software/container-stacks/detail/651c310f3b4c44380c45b7c9).

Instructions for **CRC** installation can be observed
in the [Links](#links) section.
Apart from cluster, the corresponding client is required to check
the status of the different Pods, Deployments and Services. Required
OpenShift client to install is `oc`, whose installation can be
checked in the [Links](#links) section.

Once OpenShift/K8S cluster is installed, operator can be installed
with operator-sdk.
operator-sdk installation is described in the [Links](#links) section.

In order to deploy the latest version of the operator, check latest released
version in the [Versions](#versions) section, and install the appropriate version
bundle. For example, in case latest version is **1.0.14**, the command to execute
will be:

```bash
TODO: Insert operator installation here
```

To install latest multi-arch image, execute:
```bash
$ operator-sdk run bundle quay.io/sec-eng-special/tang-operator-bundle:multi-arch
```

If the message **OLM has successfully installed** is displayed, it is normally a
sign of a proper installation of the operator.

If a message similar to **"failed open: failed to do request: context deadline exceeded"**,
it is possible that a timeout is taking place. Try to increase the timeout in case
your cluster takes long time to deploy. To do so, the option **--timeout** can be
used (if not used, default time is 2m, which stands for two minutes):

```bash
$ operator-sdk run bundle --timeout 3m quay.io/sec-eng-special/nbde-tang-server-bundle:v1.0.14
INFO[0008] Successfully created registry pod: quay-io-sec-eng-special-nbde-tang-server-bundle-v1.0.14
...
INFO[0031] OLM has successfully installed "nbde-tang-server.v1.0.14"
```

Additionally, correct installation can be observed if an output like
the following is observed when prompting for installed pods:

```bash
$ oc get pods
NAME                                                  READY STATUS    RESTARTS AGE
dbbd1837106ec169542546e7ad251b95d27c3542eb0409c1e     0/1   Completed 0        82s
quay-io-tang-nbde-tang-server-bundle-v1.0.14          1/1   Running   0        90s
nbde-tang-server-controller-manager-5c9488d8dd-mgmsf  2/2   Running   0        52s
```

Note the **Completed** and **Running** state for the different operator pods.

Once operator is correctly installed, appropriate configuration can be applied
from `config` directory. Minimal installation, that just provides the number
of replicas (1) to use, is the recommended operator configuration to apply:

```bash
$ oc apply -f operator_configs/minimal
namespace/nbde created
tangserver.nbde.openshift.io/tangserver created
```

In case operator is appropriately executed, **nbde** namespace should contain
the service, deployment and its related pods:

```
$ oc -n nbde get services
NAME               TYPE         CLUSTER-IP     EXTERNAL-IP    PORT(S)        AGE
service-tangserver LoadBalancer 172.30.167.129 34.133.181.172 8080:30831/TCP 58s

$ oc -n nbde get deployments
NAME              READY   UP-TO-DATE   AVAILABLE   AGE
tsdp-tangserver   1/1     1            1           63s

$ oc -n nbde get pods
NAME                               READY   STATUS    RESTARTS   AGE
tsdp-tangserver-55f747757c-599j5   1/1     Running   0          40s
```

Note the **Running** state for the `tangserver` pods.

## Compilation

Requirements for operator compilation are as follows:
* Go compiler (v1.19.6+). Recommended version: v1.23.2.
* Docker (v24.0.7+ recommended). Podman (v4.9.0+ recommended) can be used as an alternative to Docker.

Compilation of operator needs to be performed in top directory, by executing
**make docker-build**. The name of the image must be provided. In case there
is no requirement to update the version, same version compared to the last
version can be used. Otherwise, if new version of the operator is going
to be released, it is recommended to increase version appropriately.

In this case, same version is used. Last released version can be observed in
[Versions](#versions) section.

To summarize, taking into account that the last released version is **v1.0.14**,
compilation can be done with next command:

```bash
$ make docker-build docker-push IMG="quay.io/sec-eng-special/nbde-tang-server:v1.0.14"
...
Successfully built 4a88ba8e6426
Successfully tagged sec-eng-special/nbde-tang-server:v1.0.14
docker push sec-eng-special/nbde-tang-server:v1.0.14
The push refers to repository [quay.io/sec-eng-special/nbde-tang-server]
7910991.0.14a: Pushed
417cb9b79ade: Layer already exists
v1.0.14: digest: sha256:c97bed08ab71556542602b008888bdf23ce4afd86228a07 size: 739
```

It is possible to use `podman` instead of `docker`:

```bash
$ make podman-build podman-push IMG="quay.io/sec-eng-special/nbde-tang-server:v1.0.14"
...
Successfully built 4a88ba8e6426
Successfully tagged sec-eng-special/nbde-tang-server:v1.0.14
podman push sec-eng-special/nbde-tang-server:v1.0.14
The push refers to repository [quay.io/sec-eng-special/nbde-tang-server]
7910991.0.14a: Pushed
417cb9b79ade: Layer already exists
v1.0.14: digest: sha256:c97bed08ab71556542602b008888bdf23ce4afd86228a07 size: 739
```

In case a new release is planned to be done, the steps to follow will be:

- <ins>Modify Makefile so that it contains the new version</ins>:

```bash
$ git diff Makefile
diff --git a/Makefile b/Makefile
index 9a41c6a..db12a82 100644
--- a/Makefile
+++ b/Makefile
@@ -3,7 +3,7 @@
# To re-generate a bundle for another specific version without changing the
# standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g. make bundle VERSION=1.0.14)
# - use environment variables to overwrite this value (e.g. export VERSION=1.0.14)
-VERSION ?= 1.0.13
+VERSION ?= 1.0.14
```

Apart from previous changes, it is recommended to generate a "latest" tag for nbde-tang-server bundle:

```bash
$ docker tag quay.io/sec-eng-special/perator-bundle:v1.0.14 quay.io/sec-eng-special/nbde-tang-server-bundle:latest
$ docker push quay.io/sec-eng-special/nbde-tang-server-bundle:latest
```

In case `podman` is being used:

```bash
$ podman tag quay.io/sec-eng-special/nbde-tang-server-bundle:v1.0.14 quay.io/sec-eng-special/nbde-tang-server-bundle:latest
$ podman push quay.io/sec-eng-special/nbde-tang-server-bundle:latest
```

- <ins>Compile operator</ins>:

Compile operator code, specifying new version, by using **make docker-build** command:

```bash
$ make docker-build docker-push IMG="quay.io/sec-eng-special/nbde-tang-server:v1.0.14"
...
Successfully tagged sec-eng-special/nbde-tang-server:v1.0.14
docker push sec-eng-special/nbde-tang-server:v1.0.14
The push refers to repository [quay.io/sec-eng-special/nbde-tang-server]
9ff8a4099c67: Pushed
417cb9b79ade: Layer already exists
v1.0.14: digest: sha256:01620ab19faae54fb382a2ff285f589cf0bde6e168f14f07 size: 739
```

And, in case `podman` is being used instead of `docker`:

```bash
$ make podman-build podman-push IMG="quay.io/sec-eng-special/nbde-tang-server:v1.0.14"
...
Successfully built 4a88ba8e6426
Successfully tagged sec-eng-special/nbde-tang-server:v1.0.14
podman push sec-eng-special/nbde-tang-server:v1.0.14
The push refers to repository [quay.io/sec-eng-special/nbde-tang-server]
7910991.0.14a: Pushed
417cb9b79ade: Layer already exists
v1.0.14: digest: sha256:c97bed08ab71556542602b008888bdf23ce4afd86228a07 size: 739
```

- <ins>Bundle push</ins>:

In case the operator bundle is required to be pushed, generate
the bundle with **make bundle**, specifying appropriate image,
and push it with **make bundle-build bundle-push**:

```bash
$ make bundle IMG="quay.io/sec-eng-special/nbde-tang-server:v1.0.14"
$ make bundle-build bundle-push BUNDLE_IMG="quay.io/sec-eng-special/nbde-tang-server-bundle:v1.0.14"
...
docker push sec-eng-special/nbde-tang-server-bundle:v1.0.14
The push refers to repository [quay.io/sec-eng-special/nbde-tang-server-bundle]
02e3768cfc56: Pushed
df0c8060d328: Pushed
84774958bcf4: Pushed
v1.0.14: digest: sha256:925c2f844f941db2b53ce45cba9db7ee0be613321da8f0f05d size: 939
```

In case `podman` has been used for container generation, bundle push must be done through `podman`.
In case the operator bundle is required to be pushed, generate the bundle with **make bundle**,
specifying appropriate image, and push it with **make podman-bundle-build podman-bundle-push**:

```bash
$ make bundle IMG="quay.io/sec-eng-special/nbde-tang-server:v1.0.14"
$ make podman-bundle-build podman-bundle-push BUNDLE_IMG="quay.io/sec-eng-special/nbde-tang-server-bundle:v1.0.14"
...
podman push sec-eng-special/nbde-tang-server-bundle:v1.0.14
The push refers to repository [quay.io/sec-eng-special/nbde-tang-server-bundle]
02e3768cfc56: Pushed
df0c8060d328: Pushed
84774958bcf4: Pushed
v1.0.14: digest: sha256:925c2f844f941db2b53ce45cba9db7ee0be613321da8f0f05d size: 939
```

**IMPORTANT NOTE**: After bundle generation, next change will appear on the bundle directory:

```bash
diff --git a/bundle/manifests/nbde-tang-server.clusterserviceversion.yaml b/bundle/manifests/nbde-tang-server.clusterserviceversion.yaml
index 041d9b0..f6f2bb9 100644
--- a/bundle/manifests/nbde-tang-server.clusterserviceversion.yaml
+++ b/bundle/manifests/nbde-tang-server.clusterserviceversion.yaml
@@ -38,17 +38,6 @@ spec:
       displayName: Tang Server
       kind: TangServer
       name: tangservers.nbde.openshift.io
-      resources:
-      - kind: Deployment
-        version: v1
-      - kind: ReplicaSet
-        version: v1
-      - kind: Pod
-        version: v1
-      - kind: Secret
-        version: v1
-      - kind: Service
-        version: v1
```

**DO NOT COMMIT PREVIOUS CHANGE**, as this metadata information is required by
scorecard tests to pass successfully

- <ins>Commit changes</ins>:

Remember to **modify README.md** to include the new release version, and commit changes
performed in the operator, together with README.md and Makefile changes

## Cross Compilation

In order to cross compile the operator, prepend **GOARCH** with required architecture to
**make docker-build**:

```bash
$ GOARCH=ppc64le make docker-build docker-push IMG="quay.io/sec-eng-special/nbde-tang-server:v1.0.14"
...
Successfully built 4a88ba8e6426
Successfully tagged sec-eng-special/nbde-tang-server:v1.0.14
docker push sec-eng-special/nbde-tang-server:v1.0.14
The push refers to repository [quay.io/sec-eng-special/nbde-tang-server]
7910991.0.14a: Pushed
417cb9b79ade: Layer already exists
v1.0.14: digest: sha256:c97bed08ab71556542602b008888bdf23ce4afd86228a07 size: 739
```
Examples of architectures to cross-compile are:
* `ppc64le`
* `s390x`
* `arm64`
* `mips64`
* `riscv64`
* `amd64`

## Cleanup

For operator removal, execution of option **cleanup** from sdk-operator is the
recommended way:

```bash
$ operator-sdk cleanup nbde-tang-server
INFO[0002] operatorgroup "operator-sdk-og" deleted
...
INFO[0002] Operator "nbde-tang-server" uninstalled
```

## Tests

Execution of operator tests is pretty simple. These tests don't require any OpenShift/K8S infrastructure installed.
Execute **make test** from top directory and available tests will be executed:

```bash
$ make test
...
go fmt ./...
go vet ./...
...
setting up env vars
?       github.com/openshift/nbde-tang-server       [no test files]
?       github.com/openshift/nbde-tang-server/api/v1alpha1  [no test files]
ok      github.com/openshift/nbde-tang-server/controllers   9.720s  coverage: 14.0% of statements
```

As shown previously, coverage is calculated after test execution. Coverage data is dumped
to file **coverage.out**. To inspect coverage graphically, it can be observed by executing
next command:

```bash
$ go tool cover -html=cover.out
```

Previous command will open a web browser with the different coverage reports of the different
files that are part of the controller.

## Function Tests

Function tests are developed in [Tang Operator Test Suite Repository](https://github.com/RedHat-SP-Security/tang-operator-tests)

Please, follow instructions in previous repository to execute NBDE Tang Server function tests

## CI/CD

:exclamation: CI/CD is in a continuous "work in progress" state :exclamation:

CI/CD operator is based in Konflux. For more information, please check [Konflux Documentation](https://konflux-ci.dev/docs/)

## Scorecard

As described previously, scorecard test is executed as part of the CI/CD jobs.
However, scorecard tests can be executed manually. In order to execute these tests,
run next command:

```bash
$ operator-sdk scorecard -w 60s quay.io/sec-eng-special/nbde-tang-server-bundle:v1.0.14
...
Results:
Name: olm-status-descriptors
State: pass
...
Results:
Name: olm-spec-descriptors
State: pass
...
Results:
Name: olm-crds-have-resources
State: pass
...
Results:
Name: basic-check-spec
State: pass
...
Results:
Name: olm-crds-have-validation
State: pass
...
Results:
Name: olm-bundle-validation
State: pass
```

As it can be observed, execution of operator-sdk scorecard tests are passing successfully in latest released version (v1.0.14).

## Links

[NBDE Technology](https://access.redhat.com/articles/6987053)\
[Tang-Operator: Providing NBDE in OpenShift](https://cloud.redhat.com/blog/tang-operator-providing-nbde-in-openshift)\
[Tang-Operator demo on Killercoda](https://killercoda.com/sarroutb/scenario/tang-operator)\
[CodeReady Containers Installation](https://access.redhat.com/documentation/en-us/red_hat_codeready_containers/1.29/html/getting_started_guide/installation_gsg)\
[Minikube Installation](https://minikube.sigs.k8s.io/docs/start/)\
[operator-sdk Installation](https://sdk.operatorframework.io/docs/building-operators/golang/installation/)\
[Golang installation](https://go.dev/doc/install)\
[OpenShift CLI Installation](https://docs.openshift.com/container-platform/4.2/cli_reference/openshift_cli/getting-started-cli.html#cli-installing-cli_cli-developer-commands)\
[Validating Operators using the scorecard tool](https://docs.okd.io/latest/operators/operator_sdk/osdk-scorecard.html)\
[Tang Operator Test Suite](https://github.com/RedHat-SP-Security/tang-operator-tests)\
[Konflux Documentation](https://konflux-ci.dev/docs/)\
[Konflux: Building an Operator Lifecycle Manager (OLM) Operator](https://konflux-ci.dev/docs/advanced-how-tos/building-olm/)
