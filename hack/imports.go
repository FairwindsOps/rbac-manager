package hack

/*
Package imports imports dependencies required for "dep ensure" to fetch all of the go package dependencies needed
by kubebuilder commands to work without rerunning "dep ensure".

Example: make sure the testing libraries and apimachinery libraries are fetched by "dep ensure" so that
dep ensure doesn't need to be rerun after "kubebuilder create resource".

This is necessary for subsequent commands - such as building docs, tests, etc - to work without rerunning "dep ensure"
afterward.
*/
import _ "github.com/kubernetes-sigs/kubebuilder/pkg/imports"
