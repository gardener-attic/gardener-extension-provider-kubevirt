############# builder
FROM golang:1.13.9 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-provider-kubevirt
COPY . .
RUN make install

############# base
FROM alpine:3.11.3 AS base

############# gardener-extension-provider-kubevirt
FROM base AS gardener-extension-provider-kubevirt

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-provider-kubevirt /gardener-extension-provider-kubevirt
ENTRYPOINT ["/gardener-extension-provider-kubevirt"]

############# gardener-extension-validator-kubevirt
FROM base AS gardener-extension-validator-kubevirt

COPY --from=builder /go/bin/gardener-extension-validator-kubevirt /gardener-extension-validator-kubevirt
ENTRYPOINT ["/gardener-extension-validator-kubevirt"]
