package component

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kyma-project/modulectl/internal/common"
	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/image"
)

const (
	GithubSourceType = "Github"
	GithubAccessType = "gitHub"

	OCIArtifactResourceType     = "ociArtifact"
	OCIArtifactResourceRelation = "external"
	OCIArtifactAccessType       = "ociArtifact"

	DirectoryTreeResourceType = "directoryTree"
	DirectoryInputType        = "dir"

	PlainTextResourceType = "PlainText"
	FileResourceInput     = "file"
)

type Provider struct {
	Name   string  `yaml:"name"`
	Labels []Label `yaml:"labels,omitempty"`
}

type Input struct {
	Type         string   `yaml:"type"`
	Path         string   `yaml:"path,omitempty"`
	Data         string   `yaml:"data,omitempty"`
	Compress     bool     `yaml:"compress,omitempty"`
	IncludeFiles []string `yaml:"includeFiles,omitempty"`
}

type Access struct {
	Type           string `yaml:"type"`
	ImageReference string `yaml:"imageReference,omitempty"`
	RepoUrl        string `yaml:"repoUrl,omitempty"`
	Commit         string `yaml:"commit,omitempty"`
}

type Label struct {
	Name    string `yaml:"name"`
	Value   any    `yaml:"value"`
	Version string `yaml:"version,omitempty"`
}

// ResponsibleEntry represents a responsible team entry in OCM component descriptor.
// Field names use snake_case to match OCM specification format.
//
//nolint:tagliatelle // OCM spec requires snake_case field names
type ResponsibleEntry struct {
	GitHubHostname string `yaml:"github_hostname"`
	TeamName       string `yaml:"teamname"`
	Type           string `yaml:"type"`
}

type Resource struct {
	Name     string  `yaml:"name"`
	Type     string  `yaml:"type"`
	Version  string  `yaml:"version,omitempty"`
	Relation string  `yaml:"relation,omitempty"`
	Labels   []Label `yaml:"labels,omitempty"`
	Input    *Input  `yaml:"input,omitempty"`
	Access   *Access `yaml:"access,omitempty"`
}

type Source struct {
	Name    string  `yaml:"name"`
	Type    string  `yaml:"type"`
	Version string  `yaml:"version,omitempty"`
	Labels  []Label `yaml:"labels,omitempty"`
	Input   *Input  `yaml:"input,omitempty"`
	Access  *Access `yaml:"access,omitempty"`
}

type Component struct {
	Name      string     `yaml:"name"`
	Version   string     `yaml:"version"`
	Provider  Provider   `yaml:"provider"`
	Labels    []Label    `yaml:"labels,omitempty"`
	Resources []Resource `yaml:"resources"`
	Sources   []Source   `yaml:"sources,omitempty"`
}

type Constructor struct {
	Components []Component `yaml:"components"`
}

func NewConstructor(componentName, componentVersion string) *Constructor {
	return &Constructor{
		Components: []Component{
			{
				Name:    componentName,
				Version: componentVersion,
				Provider: Provider{
					Name: common.ProviderName,
					Labels: []Label{
						{common.BuiltByLabelKey, common.BuiltByLabelValue, common.VersionV1},
					},
				},
				Resources: make([]Resource, 0),
				Sources:   make([]Source, 0),
			},
		},
	}
}

func (c *Constructor) AddGitSource(gitRepoURL, commitHash string) {
	source := Source{
		Name:    common.OCMIdentityName,
		Type:    GithubSourceType,
		Version: c.Components[0].Version,
		Labels:  []Label{},
		Access: &Access{
			Type:    GithubAccessType,
			RepoUrl: gitRepoURL,
			Commit:  commitHash,
		},
	}

	c.Components[0].Sources = append(c.Components[0].Sources, source)
}

func (c *Constructor) AddLabel(key, value, version string) {
	labels := c.Components[0].Labels
	labelValue := Label{
		Name:    key,
		Value:   value,
		Version: version,
	}
	labels = append(labels, labelValue)
	c.Components[0].Labels = labels
}

func (c *Constructor) AddImageAsResource(imageInfos []*image.ImageInfo) {
	for _, imageInfo := range imageInfos {
		version, resourceName := generateOCMVersionAndName(imageInfo)
		resource := Resource{
			Name:     resourceName,
			Type:     OCIArtifactResourceType,
			Relation: OCIArtifactResourceRelation,
			Version:  version,
			Labels: []Label{
				{
					Name:    fmt.Sprintf("%s/%s", common.SecScanBaseLabelKey, common.TypeLabelKey),
					Value:   common.ThirdPartyImageLabelValue,
					Version: common.OCMVersion,
				},
				{
					Name:    common.OriginalImageReferenceLabelKey,
					Value:   imageInfo.FullURL,
					Version: common.OCMVersion,
				},
			},
			Access: &Access{
				Type:           OCIArtifactAccessType,
				ImageReference: imageInfo.FullURL,
			},
		}
		c.Components[0].Resources = append(c.Components[0].Resources, resource)
	}
}

func (c *Constructor) AddFileResource(resourceName, filePath string) error {
	switch resourceName {
	case common.RawManifestResourceName, common.DefaultCRResourceName:
		return c.addFileAsDirResource(resourceName, filePath)
	case common.ModuleTemplateResourceName:
		return c.addFileAsPlainTextResource(resourceName, filePath)
	default:
		return fmt.Errorf("%w: %s", commonerrors.ErrUnknownResourceName, resourceName)
	}
}

func (c *Constructor) addFileAsDirResource(resourceName, filePath string) error {
	dir, err := getAbsPath(filepath.Dir(filePath))
	if err != nil {
		return err
	}

	c.Components[0].Resources = append(c.Components[0].Resources, Resource{
		Name:    resourceName,
		Type:    DirectoryTreeResourceType,
		Version: c.Components[0].Version,
		Input: &Input{
			Type:         DirectoryInputType,
			Path:         dir,
			Compress:     false,
			IncludeFiles: []string{filepath.Base(filePath)},
		},
	})
	return nil
}

func (c *Constructor) addFileAsPlainTextResource(resourceName, filePath string) error {
	filePath, err := getAbsPath(filePath)
	if err != nil {
		return err
	}

	c.Components[0].Resources = append(c.Components[0].Resources, Resource{
		Name:    resourceName,
		Type:    PlainTextResourceType,
		Version: c.Components[0].Version,
		Input: &Input{
			Type: FileResourceInput,
			Path: filePath,
		},
	})
	return nil
}

func getAbsPath(filePath string) (string, error) {
	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path for %s: %w", filePath, err)
		}
		filePath = absPath
	}
	return filePath, nil
}

// Semantic versioning format following e.g: x.y.z or vx.y,z.
const semverPattern = `^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$` //nolint:revive,lll // for readability

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
