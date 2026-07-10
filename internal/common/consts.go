package common

const (
	ProviderName         = "kyma-project.io"
	BuiltByLabelKey      = ProviderName + "/built-by"
	BuiltByLabelValue    = "modulectl"
	VersionV1, VersionV2 = "v1", "v2"

	OCMIdentityName = "module-sources"
	OCMVersion      = "v1"

	SecurityScanLabelKey      = "security.kyma-project.io/scan"
	SecurityScanEnabledValue  = "enabled"
	SecScanBaseLabelKey       = "scan.security.kyma-project.io"
	TypeLabelKey              = "type"
	ThirdPartyImageLabelValue = "third-party-image"

	ResponsiblesLabelKey      = "cloud.gardener.cnudie/responsibles"
	GitHubHostname            = "github.tools.sap"
	ResponsibleTypeGitHubTeam = "githubTeam"

	ModuleImageResourceName    = "module-image"
	RawManifestResourceName    = "raw-manifest"
	DefaultCRResourceName      = "default-cr"
	ModuleTemplateResourceName = "moduletemplate"

	RequiresDowntimeLabelKey = "operator.kyma-project.io/requiresDowntime"

	OriginalImageReferenceLabelKey = ProviderName + "/original-image-reference"
)
