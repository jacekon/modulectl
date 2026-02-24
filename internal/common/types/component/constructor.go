package component

import (
	"fmt"
	"path/filepath"

	"github.com/kyma-project/modulectl/internal/common"
	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources"
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
		version, resourceName := resources.GenerateOCMVersionAndName(imageInfo)
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
