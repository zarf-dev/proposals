// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package v1beta1 holds the definition of the v1beta1 Zarf Package.
package v1beta1

// VariableType represents a type of a Zarf package variable.
type VariableType string

const (
	// RawVariableType is the default type for a Zarf package variable.
	RawVariableType VariableType = "raw"
	// FileVariableType loads the variable contents from a file.
	FileVariableType VariableType = "file"
)

// PackageKind is an enum of the different kinds of Zarf packages.
type PackageKind string

const (
	// ZarfInitConfig is the kind of Zarf package used during `zarf init`.
	ZarfInitConfig PackageKind = "ZarfInitConfig"
	// ZarfPackageConfig is the default kind of Zarf package.
	ZarfPackageConfig PackageKind = "ZarfPackageConfig"
	// APIVersion is the api version of this package.
	APIVersion string = "zarf.dev/v1beta1"
)

// Package is the top-level structure of a Zarf package definition.
type Package struct {
	// The API version of the Zarf package.
	APIVersion string `json:"apiVersion" jsonschema:"enum=zarf.dev/v1beta1"`
	// The kind of Zarf package.
	Kind PackageKind `json:"kind" jsonschema:"enum=ZarfInitConfig,enum=ZarfPackageConfig,default=ZarfPackageConfig"`
	// Package metadata.
	Metadata Metadata `json:"metadata,omitempty"`
	// Zarf-generated package build data.
	Build BuildData `json:"build,omitempty"`
	// List of components to deploy in this package.
	Components []Component `json:"components" jsonschema:"minItems=1"`
	// Constant template values applied on deploy.
	Constants []Constant `json:"constants,omitempty"`
	// Variable template values applied on deploy.
	Variables []InteractiveVariable `json:"variables,omitempty"`
	// Values imports Zarf values files for templating and overriding Helm values.
	Values Values `json:"values,omitempty"`
	// Documentation files included in the package.
	Documentation map[string]string `json:"documentation,omitempty"`
}

// Metadata holds information about the package.
type Metadata struct {
	// Name to identify this Zarf package.
	Name string `json:"name" jsonschema:"pattern=^[a-z0-9][a-z0-9\\-]*$"`
	// Additional information about this Zarf package.
	Description string `json:"description,omitempty"`
	// Generic string set by a package author to track the package version.
	Version string `json:"version,omitempty"`
	// Disable compression of this package.
	Uncompressed bool `json:"uncompressed,omitempty"`
	// The target cluster architecture for this package.
	Architecture string `json:"architecture,omitempty" jsonschema:"example=arm64,example=amd64"`
	// Annotations are key-value pairs that can be used to store metadata about the package.
	Annotations map[string]string `json:"annotations,omitempty"`
	// Whether to allow namespace overrides for this package.
	AllowNamespaceOverride *bool `json:"allowNamespaceOverride,omitempty"`
}

// BuildData is written during package create to track details of the created package.
type BuildData struct {
	// The machine name that created this package.
	Terminal string `json:"terminal,omitempty"`
	// The username who created this package.
	User string `json:"user,omitempty"`
	// The architecture this package was created on.
	Architecture string `json:"architecture"`
	// The timestamp when this package was created.
	Timestamp string `json:"timestamp"`
	// The version of Zarf used to build this package.
	Version string `json:"version"`
	// Any migrations that have been run on this package.
	Migrations []string `json:"migrations,omitempty"`
	// Any registry domains that were overridden on package create when pulling images.
	RegistryOverrides map[string]string `json:"registryOverrides,omitempty"`
	// Whether this package was created with differential components.
	Differential bool `json:"differential,omitempty"`
	// The flavor of Zarf used to build this package.
	Flavor string `json:"flavor,omitempty"`
	// Requirements for specific Zarf versions needed to deploy this package.
	VersionRequirements []VersionRequirement `json:"versionRequirements,omitempty"`
	// Checksum of a checksums.txt file that contains checksums all the layers within the package.
	AggregateChecksum string `json:"aggregateChecksum,omitempty"`
}

// VersionRequirement specifies a minimum Zarf version needed and the reason for the requirement.
type VersionRequirement struct {
	// The minimum version of Zarf required.
	Version string `json:"version"`
	// The reason this version is required.
	Reason string `json:"reason"`
}

// Values defines values files and schema for templating and overriding Helm values.
type Values struct {
	// List of values file paths to include.
	Files []string `json:"files,omitempty"`
	// Path to a JSON schema file for validating values.
	Schema string `json:"schema,omitempty"`
}

// Variable represents a variable that has a value set programmatically.
type Variable struct {
	// The name to be used for the variable.
	Name string `json:"name" jsonschema:"pattern=^[A-Z0-9_]+$"`
	// Whether to mark this variable as sensitive to not print it in the log.
	Sensitive bool `json:"sensitive,omitempty"`
	// Whether to automatically indent the variable's value (if multiline) when templating.
	AutoIndent bool `json:"autoIndent,omitempty"`
	// An optional regex pattern that a variable value must match before a package deployment can continue.
	Pattern string `json:"pattern,omitempty"`
	// Changes the handling of a variable to load contents differently.
	Type VariableType `json:"type,omitempty" jsonschema:"enum=raw,enum=file"`
}

// InteractiveVariable is a variable that can be used to prompt a user for more information.
type InteractiveVariable struct {
	Variable `json:",inline"`
	// A description of the variable to be used when prompting the user a value.
	Description string `json:"description,omitempty"`
	// The default value to use for the variable.
	Default string `json:"default,omitempty"`
	// Whether to prompt the user for input for this variable.
	Prompt bool `json:"prompt,omitempty"`
}

// Constant is a value used to dynamically template resources or run in actions.
type Constant struct {
	// The name to be used for the constant.
	Name string `json:"name" jsonschema:"pattern=^[A-Z0-9_]+$"`
	// The value to set for the constant during deploy.
	Value string `json:"value"`
	// A description of the constant.
	Description string `json:"description,omitempty"`
	// Whether to automatically indent the constant's value (if multiline) when templating.
	AutoIndent bool `json:"autoIndent,omitempty"`
	// An optional regex pattern that a constant value must match before a package can be created.
	Pattern string `json:"pattern,omitempty"`
}

// Component is the primary functional grouping of assets to deploy by Zarf.
type Component struct {
	// The name of the component.
	Name string `json:"name" jsonschema:"pattern=^[a-z0-9][a-z0-9\\-]*$"`
	// Message to include during package deploy describing the purpose of this component.
	Description string `json:"description,omitempty"`
	// Determines the default Y/N state for installing this component on package deploy.
	Default bool `json:"default,omitempty"`
	// Do not install this component unless explicitly requested. Defaults to false, meaning the component is required.
	Optional *bool `json:"optional,omitempty"`
	// Filter when this component is included in package creation or deployment.
	Target ComponentTarget `json:"target,omitempty"`
	// Import a component from another Zarf component config.
	Import ComponentImport `json:"import,omitempty"`
	// Zarf CLI services and infrastructure such as the registry, injector, and agent.
	Services ComponentServices `json:"services,omitempty"`
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
}

// ComponentTarget filters a component to only apply for a given local OS, architecture, or flavor.
type ComponentTarget struct {
	// Only deploy component to specified OS.
	LocalOS string `json:"localOS,omitempty" jsonschema:"enum=linux,enum=darwin,enum=windows"`
	// Only include component for the given package architecture.
	Architecture string `json:"architecture,omitempty" jsonschema:"enum=amd64,enum=arm64"`
	// Only include this component when a matching '--flavor' is specified on 'zarf package create'.
	Flavor string `json:"flavor,omitempty"`
}

// ComponentImport is a reference to an imported Zarf component config.
type ComponentImport struct {
	// The path to the component config file to import.
	Path string `json:"path,omitempty"`
	// The URL to a Zarf component config to import via OCI.
	URL string `json:"url,omitempty" jsonschema:"pattern=^oci://.*$"`
}

// ComponentServices defines Zarf CLI services to enable for an init component.
type ComponentServices struct {
	// Whether this component provides a registry.
	IsRegistry bool `json:"isRegistry,omitempty"`
	// Injector configuration for the component.
	Injector *Injector `json:"injector,omitempty"`
	// Whether this component provides an agent.
	IsAgent bool `json:"isAgent,omitempty"`
}

// Injector defines the configuration for the Zarf injector.
type Injector struct {
	// Whether the injector is enabled.
	Enabled bool `json:"enabled"`
	// Values for the injector.
	Values *InjectorValues `json:"values,omitempty"`
}

// InjectorValues defines configurable values for the Zarf injector.
type InjectorValues struct {
	// Tolerations for the injector pod.
	Tolerations string `json:"tolerations,omitempty"`
}

// Manifest defines raw manifests Zarf will deploy as a helm chart.
type Manifest struct {
	// A name to give this collection of manifests; this will become the name of the dynamically-created helm chart.
	Name string `json:"name" jsonschema:"maxLength=40"`
	// The namespace to deploy the manifests to.
	Namespace string `json:"namespace,omitempty"`
	// List of local K8s YAML files or remote URLs to deploy (in order).
	Files []string `json:"files,omitempty"`
	// Allow traversing directory above the current directory if needed for kustomization.
	KustomizeAllowAnyDirectory bool `json:"kustomizeAllowAnyDirectory,omitempty"`
	// List of local kustomization paths or remote URLs to include in the package.
	Kustomizations []string `json:"kustomizations,omitempty"`
	// Whether to not wait for manifest resources to be ready before continuing.
	NoWait *bool `json:"noWait,omitempty"`
	// Template enables go-template processing on these manifests during deploy.
	Template *bool `json:"template,omitempty"`
}

// Chart defines a helm chart to be deployed.
type Chart struct {
	// The name of the chart within Zarf; note that this must be unique and does not need to be the same as the name in the chart repository.
	Name string `json:"name"`
	// The version of the chart. This field is removed from the schema, but kept as a backwards compatibility shim so v1alpha1 packages can be converted to v1beta1.
	version string
	// The Helm repository where the chart is stored.
	HelmRepository *HelmRepositorySource `json:"helmRepository,omitempty"`
	// The Git repository where the chart is stored.
	Git *GitSource `json:"git,omitempty"`
	// The local path where the chart is stored.
	Local *LocalSource `json:"local,omitempty"`
	// The OCI registry where the chart is stored.
	OCI *OCISource `json:"oci,omitempty"`
	// The namespace to deploy the chart to.
	Namespace string `json:"namespace,omitempty"`
	// The name of the Helm release to create (defaults to the Zarf name of the chart).
	ReleaseName string `json:"releaseName,omitempty"`
	// Whether to not wait for chart resources to be ready before continuing.
	NoWait *bool `json:"noWait,omitempty"`
	// List of local values file paths or remote URLs to include in the package; these will be merged together when deployed.
	ValuesFiles []string `json:"valuesFiles,omitempty"`
	// List of value sources mapped to their Helm override targets.
	Values []ChartValue `json:"values,omitempty"`
	// Whether to validate the chart's values against its JSON schema. Defaults to true.
	SchemaValidation *bool `json:"schemaValidation,omitempty"`
}

// ChartValue maps a values source path to a Helm chart target path.
type ChartValue struct {
	// The source path for the value.
	SourcePath string `json:"sourcePath"`
	// The target path within the Helm chart values.
	TargetPath string `json:"targetPath"`
}

// HelmRepositorySource represents a Helm chart stored in a Helm repository.
type HelmRepositorySource struct {
	// The name of a chart within a Helm repository.
	Name string `json:"name,omitempty"`
	// The URL of the chart repository where the Helm chart is stored.
	URL string `json:"url"`
	// The version of the chart in the Helm repository.
	Version string `json:"version"`
}

// GitSource represents a Helm chart stored in a Git repository.
type GitSource struct {
	// The URL of the Git repository where the Helm chart is stored.
	URL string `json:"url"`
	// The subdirectory containing the chart within a Git repo.
	Path string `json:"path,omitempty"`
}

// LocalSource represents a Helm chart stored locally.
type LocalSource struct {
	// The path to a local chart's folder or .tgz archive.
	Path string `json:"path"`
}

// OCISource represents a Helm chart stored in an OCI registry.
type OCISource struct {
	// The URL of the OCI registry where the Helm chart is stored.
	URL string `json:"url"`
	// The version of the chart in the OCI registry.
	Version string `json:"version"`
}

// File defines a file to deploy.
type File struct {
	// Local folder or file path or remote URL to pull into the package.
	Source string `json:"source"`
	// Optional SHA256 checksum of the file.
	Shasum string `json:"shasum,omitempty"`
	// The absolute or relative path where the file or folder should be copied to during package deploy.
	Target string `json:"target"`
	// Determines if the file should be made executable during package deploy.
	Executable bool `json:"executable,omitempty"`
	// List of symlinks to create during package deploy.
	Symlinks []string `json:"symlinks,omitempty"`
	// Local folder or file to be extracted from a 'source' archive.
	ExtractPath string `json:"extractPath,omitempty"`
	// Template enables go-template processing on this file during deploy.
	Template *bool `json:"template,omitempty"`
}

// Image defines an OCI image to include in the package.
type Image struct {
	// The image reference.
	Name string `json:"name"`
	// The source to pull the image from. Defaults to "registry".
	Source string `json:"source,omitempty" jsonschema:"enum=registry,enum=daemon"`
}

// ImageArchive defines a tar archive of images to include in the package.
type ImageArchive struct {
	// The path to the tar archive.
	Path string `json:"path"`
	// The list of images contained in the archive.
	Images []string `json:"images"`
}

// ComponentActions are ActionSets that map to different Zarf package operations.
type ComponentActions struct {
	// Actions to run during package creation.
	OnCreate ComponentActionSet `json:"onCreate,omitempty"`
	// Actions to run during package deployment.
	OnDeploy ComponentActionSet `json:"onDeploy,omitempty"`
	// Actions to run during package removal.
	OnRemove ComponentActionSet `json:"onRemove,omitempty"`
}

// ComponentActionSet is a set of actions to run during a Zarf package operation.
type ComponentActionSet struct {
	// Default configuration for all actions in this set.
	Defaults ComponentActionDefaults `json:"defaults,omitempty"`
	// Actions to run at the start of an operation.
	Before []ComponentAction `json:"before,omitempty"`
	// Actions to run at the end of an operation.
	After []ComponentAction `json:"after,omitempty"`
	// Actions to run if any operation in this set fails.
	OnFailure []ComponentAction `json:"onFailure,omitempty"`
}

// ComponentActionDefaults sets the default configs for child actions.
type ComponentActionDefaults struct {
	// Hide the output of commands during execution (default false).
	Mute bool `json:"mute,omitempty"`
	// Default timeout in seconds for commands (default to 0, no timeout).
	MaxTotalSeconds int `json:"maxTotalSeconds,omitempty"`
	// Retry commands a given number of times if they fail (default 0).
	Retries int `json:"retries,omitempty"`
	// Working directory for commands (default CWD).
	Dir string `json:"dir,omitempty"`
	// Additional environment variables for commands.
	Env []string `json:"env,omitempty"`
	// Indicates a preference for a shell for the provided cmd to be executed in on supported operating systems.
	Shell Shell `json:"shell,omitempty"`
}

// ComponentAction represents a single action to run during a Zarf package operation.
type ComponentAction struct {
	// Hide the output of the command during package deployment (default false).
	Mute *bool `json:"mute,omitempty"`
	// Timeout in seconds for the command (default to 0, no timeout for cmd actions and 300, 5 minutes for wait actions).
	MaxTotalSeconds *int `json:"maxTotalSeconds,omitempty"`
	// Retry the command if it fails up to a given number of times (default 0).
	Retries int `json:"retries,omitempty"`
	// The working directory to run the command in (default is CWD).
	Dir *string `json:"dir,omitempty"`
	// Additional environment variables to set for the command.
	Env []string `json:"env,omitempty"`
	// The command to run. Must specify either cmd or wait for the action to do anything.
	Cmd string `json:"cmd,omitempty"`
	// Indicates a preference for a shell for the provided cmd.
	Shell *Shell `json:"shell,omitempty"`
	// An array of variables to update with the output of the command.
	SetVariables []Variable `json:"setVariables,omitempty"`
	// An array of values to set with the output of the command.
	SetValues []SetValue `json:"setValues,omitempty"`
	// Description of the action to be displayed during package execution instead of the command.
	Description string `json:"description,omitempty"`
	// Wait for a condition to be met before continuing.
	Wait *ComponentActionWait `json:"wait,omitempty"`
	// Template enables go-template processing on the cmd field.
	Template *bool `json:"template,omitempty"`
}

// SetValueType declares the expected input back from the cmd, allowing structured data to be parsed.
type SetValueType string

const (
	// SetValueYAML enables YAML parsing.
	SetValueYAML SetValueType = "yaml"
	// SetValueJSON enables JSON parsing.
	SetValueJSON SetValueType = "json"
	// SetValueString sets the raw value.
	SetValueString SetValueType = "string"
)

// SetValue declares a value that can be set during a package deploy.
type SetValue struct {
	// Key represents which value to assign to.
	Key string `json:"key,omitempty"`
	// Value is the current value at the key.
	Value any `json:"value,omitempty"`
	// Type declares the kind of data being stored in the value.
	Type SetValueType `json:"type,omitempty"`
}

// ComponentActionWait specifies a condition to wait for before continuing.
type ComponentActionWait struct {
	// Wait for a condition to be met in the cluster before continuing. Only one of cluster or network can be specified.
	Cluster *ComponentActionWaitCluster `json:"cluster,omitempty"`
	// Wait for a condition to be met on the network before continuing. Only one of cluster or network can be specified.
	Network *ComponentActionWaitNetwork `json:"network,omitempty"`
}

// ComponentActionWaitCluster specifies a cluster-level condition to wait for.
type ComponentActionWaitCluster struct {
	// The kind of resource to wait for.
	Kind string `json:"kind" jsonschema:"example=Pod,example=Deployment"`
	// The name of the resource or selector to wait for.
	Name string `json:"name" jsonschema:"example=podinfo,example=app=podinfo"`
	// The namespace of the resource to wait for.
	Namespace string `json:"namespace,omitempty"`
	// The condition or jsonpath state to wait for; defaults to kstatus readiness checks.
	Condition string `json:"condition,omitempty" jsonschema:"example=Available,'{.status.availableReplicas}'=23"`
}

// ComponentActionWaitNetwork specifies a network-level condition to wait for.
type ComponentActionWaitNetwork struct {
	// The protocol to wait for.
	Protocol string `json:"protocol" jsonschema:"enum=tcp,enum=http,enum=https"`
	// The address to wait for.
	Address string `json:"address" jsonschema:"example=localhost:8080,example=1.1.1.1"`
	// The HTTP status code to wait for if using http or https.
	Code int `json:"code,omitempty" jsonschema:"example=200,example=404"`
}

// Shell represents the desired shell to use for a given command.
type Shell struct {
	// Windows shell preference.
	Windows string `json:"windows,omitempty" jsonschema:"example=powershell,example=cmd,example=pwsh"`
	// Linux shell preference.
	Linux string `json:"linux,omitempty" jsonschema:"example=sh,example=bash,example=zsh"`
	// Darwin (macOS) shell preference.
	Darwin string `json:"darwin,omitempty" jsonschema:"example=sh,example=bash,example=zsh"`
}
