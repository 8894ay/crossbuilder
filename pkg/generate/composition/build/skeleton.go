package build

import (
	"fmt"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
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

var (
	// KnownCompositeAnnotations are annotations that will be registered by
	// default
	KnownCompositeAnnotations = []string{}

	// KnownCompositeLabels are labels that will be registered by default.
	KnownCompositeLabels = []string{
		labelKeyClaimName,
		labelKeyClaimNamespace,
	}
	// KnownResourceAnnotations are annotations that will be registered by
	// default
	KnownResourceAnnotations = []string{
		meta.AnnotationKeyExternalName,
		meta.AnnotationKeyExternalCreatePending,
		meta.AnnotationKeyExternalCreateSucceeded,
		meta.AnnotationKeyExternalCreateFailed,
	}
	// KnownResourceLabels are labels that will be registered by default.
	KnownResourceLabels = []string{}
)

// ComposedTemplateSkeleton represents the draft for a compositionSkeleton composeTemplateSkeleton.
type ComposedTemplateSkeleton interface {
	// WithName sets the name of this composeTemplateSkeleton.
	WithName(name string) ComposedTemplateSkeleton

	// RegisterAnnotations marks the given resource annotations as safe
	// so they will be treated as a valid field in patch paths.
	RegisterAnnotations(annotationKeys ...string) ComposedTemplateSkeleton

	// RegisterLabels marks the given resource label as safe
	// so they will be treated as a valid field in patch paths.
	RegisterLabels(labelsKeys ...string) ComposedTemplateSkeleton

	// RegisterFieldPaths marks the given resource paths as safe so ti will
	// be treated a valid in patch paths.
	RegisterFieldPaths(paths ...string) ComposedTemplateSkeleton
}

// CompositionSkeleton represents the build time state of a composition.
type CompositionSkeleton interface {
	// WithName sets the metadata.name of the composition to be built.
	WithName(name string) CompositionSkeleton

	// WithWriteConnectionSecretsToNamespace sets the
	// WriteConnectionSecretsToNamespace of this compositionSkeleton.
	WithWriteConnectionSecretsToNamespace(namespace *string) CompositionSkeleton

	// RegisterCompositeAnnotations marks the given composite annotations as safe
	// so it will be treated as a valid field in patch paths.
	RegisterCompositeAnnotations(annotationKeys ...string) CompositionSkeleton

	// RegisterCompositeLabels marks the given composite labels as safe
	// so it will be treated as a valid field in patch paths.
	RegisterCompositeLabels(labelKeys ...string) CompositionSkeleton

	// RegisterCompositeFieldPaths marks the given composite paths as safe so
	// they will be treated a valid in patch paths.
	RegisterCompositeFieldPaths(paths ...string) CompositionSkeleton
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
	composite ObjectKindReference

	registeredPaths                         []string
	name                                    string
	writeConnectionSecretsToNamespace       *string
}

// RegisterCompositeAnnotations marks the given composite annotations as safe so
// it will be treated as a valid field in patch paths.
func (c *compositionSkeleton) RegisterCompositeAnnotations(annotionKeys ...string) CompositionSkeleton {
	paths := make([]string, len(annotionKeys))
	for i, k := range annotionKeys {
		paths[i] = fmt.Sprintf("metadata.annotations[%s]", k)
	}
	return c.RegisterCompositeFieldPaths(paths...)
}

// RegisterCompositeLabels marks the given composite labels as safe so it
// will be treated as a valid field in patch paths.
func (c *compositionSkeleton) RegisterCompositeLabels(labelKeys ...string) CompositionSkeleton {
	paths := make([]string, len(labelKeys))
	for i, k := range labelKeys {
		paths[i] = fmt.Sprintf("metadata.labels[%s]", k)
	}
	return c.RegisterCompositeFieldPaths(paths...)
}

// RegisterCompositeFieldPaths marks the given composite paths as safe so ti will
// be treated a valid in patch paths.
func (c *compositionSkeleton) RegisterCompositeFieldPaths(path ...string) CompositionSkeleton {
	c.registeredPaths = append(c.registeredPaths, path...)
	return c
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

// ToComposition generates a Crossplane compositionSkeleton from this compositionSkeleton.
func (c *compositionSkeleton) ToComposition() (xapiextv1.Composition, error) {
	if c.name == "" {
		return xapiextv1.Composition{}, errors.New(errEmptyCompositionname)
	}

	c.RegisterCompositeAnnotations(KnownCompositeAnnotations...)
	c.RegisterCompositeLabels(KnownCompositeLabels...)


	comp := xapiextv1.Composition{
		Spec: xapiextv1.CompositionSpec{
			CompositeTypeRef:                           xapiextv1.TypeReferenceTo(c.composite.GroupVersionKind),
			Mode:                                        xapiextv1.CompositionModePipeline,
			WriteConnectionSecretsToNamespace:          c.writeConnectionSecretsToNamespace,
		},
	}
	comp.SetGroupVersionKind(xapiextv1.CompositionGroupVersionKind)
	comp.SetName(c.name)
	comp.SetCreationTimestamp(v1.Time{})
	return comp, nil
}

type composeTemplateSkeleton struct {
	compositionSkeleton *compositionSkeleton

	registeredPaths   []string
	name              *string
	base              ObjectKindReference
}

// RegisterAnnotations marks the given resource annotations as safe
// so they will be treated as a valid field in patch paths.
func (c *composeTemplateSkeleton) RegisterAnnotations(annotionKeys ...string) ComposedTemplateSkeleton {
	return c.RegisterFieldPaths(makeAnnotationPaths(annotionKeys)...)
}

// RegisterLabels marks the given resource labels as safe
// so they will be treated as a valid field in patch paths.
func (c *composeTemplateSkeleton) RegisterLabels(labelKeys ...string) ComposedTemplateSkeleton {
	return c.RegisterFieldPaths(makeLabelPaths(labelKeys)...)
}

// RegisterFieldPaths marks the given resource paths as safe so they will
// be treated a valid in patch paths.
func (c *composeTemplateSkeleton) RegisterFieldPaths(paths ...string) ComposedTemplateSkeleton {
	c.registeredPaths = append(c.registeredPaths, paths...)
	return c
}

// WithName sets the name of this composeTemplateSkeleton.
func (c *composeTemplateSkeleton) WithName(name string) ComposedTemplateSkeleton {
	c.name = &name
	return c
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
