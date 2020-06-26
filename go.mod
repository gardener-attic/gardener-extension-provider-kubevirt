module github.com/gardener/gardener-extension-provider-kubevirt

go 1.14

require (
	github.com/ahmetb/gen-crd-api-reference-docs v0.2.0
	github.com/gardener/etcd-druid v0.3.0
	github.com/gardener/gardener v1.6.4
	github.com/gardener/machine-controller-manager v0.27.0
	github.com/go-logr/logr v0.1.0
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/golang/mock v1.4.3
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.4
	k8s.io/apiextensions-apiserver v0.17.6
	k8s.io/apimachinery v0.18.4
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.18.4
	k8s.io/component-base v0.18.4
	sigs.k8s.io/controller-runtime v0.5.4
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	k8s.io/api => k8s.io/api v0.16.8 // 1.16.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.8 // 1.16.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.8 // 1.16.8
	k8s.io/apiserver => k8s.io/apiserver v0.16.8 // 1.16.8
	k8s.io/client-go => k8s.io/client-go v0.16.8 // 1.16.8
	k8s.io/code-generator => k8s.io/code-generator v0.16.8 // 1.16.8
	k8s.io/helm => k8s.io/helm v2.13.1+incompatible
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.8 // 1.16.8
)
