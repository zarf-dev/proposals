// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

package v1beta1

// ComponentConfig is the top-level structure of a Zarf component config file.
type ComponentConfig struct {
	// The API version of the component config.
	APIVersion string `json:"apiVersion" jsonschema:"enum=zarf.dev/v1beta1"`
	// The kind of component config.
	Kind PackageKind `json:"kind" jsonschema:"enum=ZarfComponentConfig,default=ZarfComponentConfig"`
	// Component metadata.
	Metadata ComponentMetadata `json:"metadata"`
	// A single component definition that applies in all contexts. Exactly one of Component or Variants must be set.
	Component *ImportableComponent `json:"component,omitempty" jsonschema:"oneof_required=component"`
	// A list of component variants, each with a distinct .target filter. Use this when the
	// component has different definitions for different flavors, OSes, or architectures.
	Variants []Variant `json:"variants,omitempty" jsonschema:"oneof_required=variants"`
	// Constant template values applied on deploy.
	Constants []Constant `json:"constants,omitempty"`
	// Variable template values applied on deploy.
	Variables []InteractiveVariable `json:"variables,omitempty"`
	// Values imports Zarf values files for templating and overriding Helm values.
	Values Values `json:"values,omitempty"`
	// Zarf-generated publish data for the component config.
	PublishData ComponentPublishData `json:"publishData,omitempty"`
}

// ImportableComponent is a reduced component definition used in component configs.
type ImportableComponent struct {
	// Import a component from another Zarf component config.
	Import ComponentImport `json:"import,omitempty"`
	// Kubernetes manifests to be included in a generated Helm chart on package deploy.
	Manifests []Manifest `json:"manifests,omitempty"`
	// Helm charts to install during package deploy.
	Charts []Chart `json:"charts,omitempty"`
	// Files or folders to place on disk during package deployment.
	Files []File `json:"files,omitempty"`
	// List of OCI images to include in the package.
	Images []Image `json:"images,omitempty"`
	// List of tar archives of images to include in the package.
	ImageArchives []ImageArchive `json:"imageArchives,omitempty"`
	// List of git repositories to include in the package.
	Repositories []string `json:"repositories,omitempty"`
	// Custom commands to run at various stages of a package lifecycle.
	Actions ComponentActions `json:"actions,omitempty"`
	// The Zarf CLI service this component provides, such as the registry, injector, or agent.
	Service Service `json:"service,omitempty" jsonschema:"enum=registry,enum=seed-registry,enum=injector,enum=agent,enum=git-server"`
}

// Variant is a component definition with a required filter for when it applies.
type Variant struct {
	ImportableComponent
	// Filter when this variant is included in package creation or deployment.
	Target ComponentTarget `json:"target"`
}

// ComponentMetadata holds metadata about a component config.
type ComponentMetadata struct {
	// Name to identify this component config.
	Name string `json:"name" jsonschema:"pattern=^[a-z0-9][a-z0-9\\-]*$"`
	// Additional information about this component config.
	Description string `json:"description,omitempty"`
	// Generic string to track the component config version.
	Version string `json:"version,omitempty"`
	// Annotations contains arbitrary metadata about the component config.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ComponentPublishData is written during publish to track details of the component config.
type ComponentPublishData struct {
	// The version of Zarf used to build this component config.
	ZarfVersion string `json:"zarfVersion"`
	// Any migrations that have been run on this component config.
	Migrations []string `json:"migrations,omitempty"`
	// Requirements for specific package operations.
	VersionRequirements []VersionRequirement `json:"versionRequirements,omitempty"`
}
