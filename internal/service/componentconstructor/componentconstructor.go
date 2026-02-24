package componentconstructor

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/kyma-project/modulectl/internal/common"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/common/types/component"
	"github.com/kyma-project/modulectl/internal/service/image"
	"github.com/kyma-project/modulectl/tools/filesystem"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) AddResources(
	componentConstructor *component.Constructor,
	resourcePaths *types.ResourcePaths,
) error {
	err := componentConstructor.AddFileResource(common.RawManifestResourceName, resourcePaths.RawManifest)
	if err != nil {
		return fmt.Errorf("failed to create raw manifest resource: %w", err)
	}
	if resourcePaths.DefaultCR != "" {
		err = componentConstructor.AddFileResource(common.DefaultCRResourceName, resourcePaths.DefaultCR)
		if err != nil {
			return fmt.Errorf("failed to create default CR resource: %w", err)
		}
	}
	err = componentConstructor.AddFileResource(common.ModuleTemplateResourceName, resourcePaths.ModuleTemplate)
	if err != nil {
		return fmt.Errorf("failed to create moduletemplate resource: %w", err)
	}
	return nil
}

func (s *Service) CreateConstructorFile(componentConstructor *component.Constructor, filePath string) error {
	marshal, err := yaml.Marshal(componentConstructor)
	if err != nil {
		return fmt.Errorf("unable to marshal component constructor: %w", err)
	}

	helper := &filesystem.Helper{}
	if err = helper.WriteFile(filePath, string(marshal)); err != nil {
		return fmt.Errorf("unable to write component constructor file: %w", err)
	}
	return nil
}

func (s *Service) AddImagesToConstructor(
	componentConstructor *component.Constructor,
	images []string,
) error {
	imageInfos := make([]*image.ImageInfo, 0, len(images))
	for _, img := range images {
		imageInfo, err := image.ValidateAndParseImageInfo(img)
		if err != nil {
			return fmt.Errorf("image validation failed for %s: %w", img, err)
		}
		imageInfos = append(imageInfos, imageInfo)
	}
	componentConstructor.AddImageAsResource(imageInfos)
	return nil
}

// SetComponentLabel sets a top-level label on the first component of the constructor.
// The label version defaults to "v1".
// The label is overwritten if it already exists.
func (s *Service) SetComponentLabel(
	componentConstructor *component.Constructor,
	labelName, value string,
) {
	labels := componentConstructor.Components[0].Labels
	for i, label := range labels {
		if label.Name == labelName {
			labels[i].Value = value
			return
		}
	}

	componentConstructor.AddLabel(labelName, value, common.VersionV1)
}

// SetResponsiblesLabel sets the cloud.gardener.cnudie/responsibles label with team information.
func (s *Service) SetResponsiblesLabel(
	componentConstructor *component.Constructor,
	team string,
) {
	label := component.Label{
		Name: common.ResponsiblesLabelKey,
		Value: []component.ResponsibleEntry{{
			GitHubHostname: common.GitHubHostname,
			TeamName:       team,
			Type:           common.ResponsibleTypeGitHubTeam,
		}},
		Version: common.VersionV1,
	}
	componentConstructor.Components[0].Labels = append(componentConstructor.Components[0].Labels, label)
}
