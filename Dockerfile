############# builder
FROM eu.gcr.io/gardener-project/3rd/golang:1.16.2 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-provider-kubevirt
COPY . .
RUN make install

############# base
FROM eu.gcr.io/gardener-project/3rd/alpine:3.13.2 AS base

############# gardener-extension-provider-kubevirt
FROM base AS gardener-extension-provider-kubevirt

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-provider-kubevirt /gardener-extension-provider-kubevirt
ENTRYPOINT ["/gardener-extension-provider-kubevirt"]

############# gardener-extension-admission-kubevirt
FROM base AS gardener-extension-admission-kubevirt

COPY --from=builder /go/bin/gardener-extension-admission-kubevirt /gardener-extension-admission-kubevirt
ENTRYPOINT ["/gardener-extension-admission-kubevirt"]
