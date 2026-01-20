package build

import (
	"fmt"

	xapiextv1 "github.com/crossplane/crossplane/v2/apis/apiextensions/v1"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	errEmptyCompositionname                 = "composition name must not be empty"
	errFmtBuildComposedTemplate             = "cannot build composed template at index %d"
	errFmtInvalidPatch                      = "invalid patch at index %d"
	errPatchFromFieldPath                   = "fromFieldPath is invalid"
	errPatchToFieldPath                     = "toFieldPath is invalid"
	errPatchRequireField                    = "missing field %s"
	errPatchCombineEmptyVariables           = "no variables given"
	errFmtPatchCombineVariableFromFieldPath = "fromFieldPath of variable at index %d is invalid"
	errUnknownPatchType                     = "unknown patch type %s"
	errParseRegisteredCompositePaths        = "cannot parse registered composite paths"
	errParseRegisteredComposedPaths         = "cannot parse registered composed paths"

	labelKeyClaimName      = "crossplane.io/claim-name"
	labelKeyClaimNamespace = "crossplane.io/claim-namespace"
)

// CompositionSkeleton represents the build time state of a composition.
type CompositionSkeleton interface {
	// WithName sets the metadata.name of the composition to be built.
	WithName(name string) CompositionSkeleton

	// WithWriteConnectionSecretsToNamespace sets the
	// WriteConnectionSecretsToNamespace of this compositionSkeleton.
	WithWriteConnectionSecretsToNamespace(namespace *string) CompositionSkeleton

	// WithPipelineSteps sets the pipeline steps of this compositionSkeleton.
	WithPipelineSteps(steps ...xapiextv1.PipelineStep) CompositionSkeleton
}

// Object is an extension of the k8s runtime.Object with additional functions
// that are required by Crossbuildec.
type Object interface {
	runtime.Object
	SetGroupVersionKind(gvk schema.GroupVersionKind)
}

// ObjectKindReference contains the group version kind and instance of a
// runtime.Object.
type ObjectKindReference struct {
	// GroupVersionKind is the GroupVersionKind for the composite type.
	GroupVersionKind schema.GroupVersionKind

	// Object is an instance of the composite type.
	Object Object
}

type compositionSkeleton struct {
	composite 								         ObjectKindReference
	name                               string
	writeConnectionSecretsToNamespace *string
	pipelineSteps                     []xapiextv1.PipelineStep
}

// WithName sets the metadata.name of the composition to be built.
func (c *compositionSkeleton) WithName(name string) CompositionSkeleton {
	c.name = name
	return c
}

// WithWriteConnectionSecretsToNamespace sets the
// WriteConnectionSecretsToNamespace of this compositionSkeleton.
func (c *compositionSkeleton) WithWriteConnectionSecretsToNamespace(namespace *string) CompositionSkeleton {
	c.writeConnectionSecretsToNamespace = namespace
	return c
}

// WithPipelineSteps sets the pipeline steps of this compositionSkeleton.
func (c *compositionSkeleton) WithPipelineSteps(steps ...xapiextv1.PipelineStep) CompositionSkeleton {
	c.pipelineSteps = steps
	return c
}

// ToComposition generates a Crossplane compositionSkeleton from this compositionSkeleton.
func (c *compositionSkeleton) ToComposition() (xapiextv1.Composition, error) {
	if c.name == "" {
		return xapiextv1.Composition{}, errors.New(errEmptyCompositionname)
	}

	comp := xapiextv1.Composition{
		Spec: xapiextv1.CompositionSpec{
			CompositeTypeRef:                  xapiextv1.TypeReferenceTo(c.composite.GroupVersionKind),
			Mode:                              xapiextv1.CompositionModePipeline,
			WriteConnectionSecretsToNamespace: c.writeConnectionSecretsToNamespace,
		},
	}

	comp.SetGroupVersionKind(xapiextv1.CompositionGroupVersionKind)
	comp.SetName(c.name)
	comp.SetCreationTimestamp(v1.Time{})
	comp.Spec.Pipeline = c.pipelineSteps
	return comp, nil
}

func makeLabelPaths(keys []string) []string {
	paths := make([]string, len(keys))
	for i, k := range keys {
		paths[i] = fmt.Sprintf("metadata.labels[%s]", k)
	}
	return paths
}

func makeAnnotationPaths(keys []string) []string {
	paths := make([]string, len(keys))
	for i, k := range keys {
		paths[i] = fmt.Sprintf("metadata.annotations[%s]", k)
	}
	return paths
}
