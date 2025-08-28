package markers

import (
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xapiext "github.com/crossplane/crossplane/v2/apis/apiextensions/v2"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// XRDMarkers lists all markers that directly modify the XRD (not validation
// schemas).
var XRDMarkers = []*definitionWithHelp{
	must(markers.MakeDefinition("crossbuilder:generate:xrd:defaultCompositionRef", markers.DescribesType, DefaultCompositionRef{})),
	must(markers.MakeDefinition("crossbuilder:generate:xrd:enforcedCompositionRef", markers.DescribesType, EnforcedCompositionRef{})),
	must(markers.MakeDefinition("crossbuilder:generate:xrd:defaultCompositeDeletePolicy", markers.DescribesType, DefaultCompositeDeletePolicy{})),
	must(markers.MakeDefinition("crossbuilder:generate:xrd:connectionSecretKeys", markers.DescribesType, ConnectionSecretKeys(nil))),
}

func init() {
	AllDefinitions = append(AllDefinitions, XRDMarkers...)
}

// +controllertools:marker:generateHelp:category=XRD

// DefaultCompositionRef is a marker to specify the default composition ref of
// an XRD.
type DefaultCompositionRef struct {
	Name string `marker:"name"`
}

// ApplyToXRD applies the default composition ref to the XRD.
func (c DefaultCompositionRef) ApplyToXRD(xrd *xapiext.CompositeResourceDefinition, version string) error {
	xrd.Spec.DefaultCompositionRef = &xapiext.CompositionReference{
		Name: c.Name,
	}
	// test(c)
	return nil
}

// +controllertools:marker:generateHelp:category=XRD

// EnforcedCompositionRef is a marker to specify the enforced composition ref of
// an XRD.
type EnforcedCompositionRef struct {
	Name string `marker:"name"`
}

// ApplyToXRD applies the enforced composition ref to the XRD.
func (c EnforcedCompositionRef) ApplyToXRD(xrd *xapiext.CompositeResourceDefinition, version string) error {
	xrd.Spec.EnforcedCompositionRef = &xapiext.CompositionReference{
		Name: c.Name,
	}
	// test(c)
	return nil
}

// +controllertools:marker:generateHelp:category=XRD

// DefaultCompositeDeletePolicy is a marker to specify the default composite
// delete policy of an XRD.
type DefaultCompositeDeletePolicy struct {
	Policy xpv1.CompositeDeletePolicy `marker:"policy"`
}

// ApplyToXRD applies the enforced composition ref to the XRD.
func (c DefaultCompositeDeletePolicy) ApplyToXRD(xrd *xapiext.CompositeResourceDefinition, version string) error {
	xrd.Spec.DefaultCompositeDeletePolicy = &c.Policy
	// test(c)
	return nil
}

// ConnectionSecretKeys is a marker to specify connection secret keys of an XRD
type ConnectionSecretKeys []string

func (c ConnectionSecretKeys) ApplyToXRD(xrd *xapiext.CompositeResourceDefinition, version string) error {
	xrd.Spec.ConnectionSecretKeys = c
	return nil
}
