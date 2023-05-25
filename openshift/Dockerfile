FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.19-openshift-4.14 AS builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=${GOOS} GOPROXY=${GOPROXY} go build \
  -o catalogd-controller \
  ./cmd/manager/main.go

FROM registry.ci.openshift.org/ocp/4.14:base
COPY --from=builder /build/catalogd-controller /usr/bin/catalogd-controller
LABEL io.k8s.display-name="OpenShift Operator Lifecycle Manager Catalog Controller" \
      io.k8s.description="This is a component of OpenShift Container Platform that provides operator catalog support."
