package resources

import (
	"errors"
	"fmt"

	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/artifacttypes"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources/accesshandler"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

const (
	moduleImageResourceName = "module-image"
	rawManifestResourceName = "raw-manifest"
	defaultCRResourceName   = "default-cr"
)

var ErrNilTarGenerator = errors.New("tarGenerator must not be nil")

type Service struct {
	tarGenerator accesshandler.TarGenerator
}

func NewService(tarGen accesshandler.TarGenerator) (*Service, error) {
	if tarGen == nil {
		return nil, ErrNilTarGenerator
	}

	return &Service{
		tarGenerator: tarGen,
	}, nil
}

type AccessHandler interface {
	GenerateBlobAccess() (cpi.BlobAccess, error)
}

type Resource struct {
	compdesc.Resource
	AccessHandler AccessHandler
}

func (s *Service) GenerateModuleResources(moduleConfig *contentprovider.ModuleConfig,
	manifestPath, defaultCRPath string,
) ([]Resource, error) {
	moduleImageResource := GenerateModuleImageResource()
	metadataResource, err := GenerateMetadataResource(moduleConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate metadata resource: %w", err)
	}
	rawManifestResource := GenerateRawManifestResource(s.tarGenerator, manifestPath)
	resources := []Resource{moduleImageResource, metadataResource, rawManifestResource}
	if defaultCRPath != "" {
		defaultCRResource := GenerateDefaultCRResource(s.tarGenerator, defaultCRPath)
		resources = append(resources, defaultCRResource)
	}

	for idx := range resources {
		resources[idx].Version = moduleConfig.Version
	}
	return resources, nil
}

func GenerateModuleImageResource() Resource {
	return Resource{
		Resource: compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name: moduleImageResourceName,
				},
				Type:     artifacttypes.OCI_ARTIFACT,
				Relation: ocmv1.ExternalRelation,
			},
		},
	}
}

func GenerateRawManifestResource(tarGen accesshandler.TarGenerator, manifestPath string) Resource {
	return Resource{
		Resource: compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name: rawManifestResourceName,
				},
				Type:     artifacttypes.DIRECTORY_TREE,
				Relation: ocmv1.LocalRelation,
			},
		},
		AccessHandler: accesshandler.NewTar(tarGen, manifestPath),
	}
}

func GenerateDefaultCRResource(tarGen accesshandler.TarGenerator, defaultCRPath string) Resource {
	return Resource{
		Resource: compdesc.Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name: defaultCRResourceName,
				},
				Type:     artifacttypes.DIRECTORY_TREE,
				Relation: ocmv1.LocalRelation,
			},
		},
		AccessHandler: accesshandler.NewTar(tarGen, defaultCRPath),
	}
}
