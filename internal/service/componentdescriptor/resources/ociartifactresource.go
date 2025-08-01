package resources

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	ociartifacttypes "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"

	"github.com/kyma-project/modulectl/internal/service/image"
)

const (
	// Semantic versioning format following e.g: x.y.z or vx.y,z
	semverPattern             = `^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`
	secScanBaseLabelKey       = "scan.security.kyma-project.io"
	typeLabelKey              = "type"
	thirdPartyImageLabelValue = "third-party-image"
	ocmVersion                = "v1"
)

var ErrInvalidImageFormat = errors.New("invalid image url format")

func NewOciArtifactResource(imageInfo *image.ImageInfo) (*compdesc.Resource, error) {
	if imageInfo == nil || imageInfo.FullURL == "" {
		return nil, fmt.Errorf("image info is nil or empty: %w", ErrInvalidImageFormat)
	}

	typeLabel, err := createLabel()
	if err != nil {
		return nil, err
	}
	version, resourceName := generateOCMVersionAndName(imageInfo)
	access := ociartifact.New(imageInfo.FullURL)
	access.SetType(ociartifact.Type)

	return &compdesc.Resource{
		ResourceMeta: compdesc.ResourceMeta{
			Type:     ociartifacttypes.TYPE,
			Relation: ocmv1.ExternalRelation,
			ElementMeta: compdesc.ElementMeta{
				Name:    resourceName,
				Labels:  []ocmv1.Label{*typeLabel},
				Version: version,
			},
		},
		Access: access,
	}, nil
}

func AddResourceIfNotExists(descriptor *compdesc.ComponentDescriptor, resource *compdesc.Resource) {
	for _, r := range descriptor.Resources {
		if r.Name == resource.Name && r.Version == resource.Version {
			return // Already exists, skip
		}
	}
	descriptor.Resources = append(descriptor.Resources, *resource)
	compdesc.DefaultResources(descriptor)
}

func createLabel() (*ocmv1.Label, error) {
	labelKey := fmt.Sprintf("%s/%s", secScanBaseLabelKey, typeLabelKey)
	label, err := ocmv1.NewLabel(labelKey, thirdPartyImageLabelValue, ocmv1.WithVersion(ocmVersion))
	if err != nil {
		return nil, fmt.Errorf("failed to create OCM label: %w", err)
	}
	return label, nil
}

func generateOCMVersionAndName(info *image.ImageInfo) (string, string) {
	if info.Digest != "" {
		shortDigest := info.Digest[:12]
		var version string
		switch {
		case info.Tag != "" && isValidSemverForOCM(info.Tag):
			version = fmt.Sprintf("%s+sha256.%s", info.Tag, shortDigest)
		case info.Tag != "":
			version = fmt.Sprintf("0.0.0-%s+sha256.%s", normalizeTagForOCM(info.Tag), shortDigest)
		default:
			version = "0.0.0+sha256." + shortDigest
		}
		resourceName := fmt.Sprintf("%s-%s", info.Name, info.Digest[:8])
		return version, resourceName
	}

	var version string
	if isValidSemverForOCM(info.Tag) {
		version = info.Tag
	} else {
		version = "0.0.0-" + normalizeTagForOCM(info.Tag)
	}

	resourceName := info.Name
	return version, resourceName
}

func normalizeTagForOCM(tag string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9.-]`)
	normalized := reg.ReplaceAllString(tag, "-")
	normalized = strings.Trim(normalized, "-.")
	if normalized == "" {
		normalized = "unknown"
	}
	return normalized
}

func isValidSemverForOCM(version string) bool {
	matched, _ := regexp.MatchString(semverPattern, version)
	return matched
}
