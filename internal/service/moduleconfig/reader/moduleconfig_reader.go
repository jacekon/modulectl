package moduleconfigreader

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/validation"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
}

type Service struct {
	fileSystem FileSystem
}

func NewService(fileSystem FileSystem) (*Service, error) {
	if fileSystem == nil {
		return nil, fmt.Errorf("fileSystem must not be nil: %w", commonerrors.ErrInvalidArg)
	}

	return &Service{
		fileSystem: fileSystem,
	}, nil
}

func (s *Service) ParseAndValidateModuleConfig(moduleConfigFile string,
) (*contentprovider.ModuleConfig, error) {
	moduleConfig, err := ParseModuleConfig(moduleConfigFile, s.fileSystem)
	if err != nil {
		return nil, fmt.Errorf("failed to parse module config file: %w", err)
	}

	if err = ValidateModuleConfig(moduleConfig); err != nil {
		return nil, fmt.Errorf("failed to validate module config: %w", err)
	}

	return moduleConfig, nil
}

//nolint:cyclop,funlen // validation function with many checks is acceptable
func ValidateModuleConfig(moduleConfig *contentprovider.ModuleConfig) error {
	if err := validation.ValidateModuleName(moduleConfig.Name); err != nil {
		return fmt.Errorf("failed to validate module name: %w", err)
	}

	if err := validation.ValidateModuleVersion(moduleConfig.Version); err != nil {
		return fmt.Errorf("failed to validate module version: %w", err)
	}

	if moduleConfig.Manifest.IsURL() {
		if moduleConfig.Manifest.URL().Scheme != "https" {
			return fmt.Errorf(
				"failed to validate manifest: %w", fmt.Errorf(
					"'%s' is not using https scheme: %w",
					moduleConfig.Manifest.String(),
					commonerrors.ErrInvalidOption,
				),
			)
		}
	} else {
		if moduleConfig.Manifest.IsEmpty() {
			return fmt.Errorf("failed to validate manifest: must not be empty: %w", commonerrors.ErrInvalidOption)
		}
		if strings.HasPrefix(moduleConfig.Manifest.String(), "/") {
			return fmt.Errorf("failed to validate manifest: must not be an absolute path: %w",
				commonerrors.ErrInvalidOption)
		}
	}

	if err := validation.ValidateIsValidHTTPSURL(moduleConfig.Repository); err != nil {
		return fmt.Errorf("failed to validate repository: %w", err)
	}

	// Team is required only when security scan is enabled (default: true)
	securityScanEnabled := moduleConfig.SecurityScanEnabled == nil || *moduleConfig.SecurityScanEnabled
	if securityScanEnabled && moduleConfig.Team == "" {
		return fmt.Errorf("failed to validate team: must not be empty when security scan is enabled: %w",
			commonerrors.ErrInvalidOption)
	}

	if err := validation.ValidateIsValidHTTPSURL(moduleConfig.Documentation); err != nil {
		return fmt.Errorf("failed to validate documentation: %w", err)
	}

	if len(moduleConfig.Icons) == 0 {
		return fmt.Errorf("failed to validate module icons: must contain at least one icon: %w",
			commonerrors.ErrInvalidOption)
	}

	if err := validation.ValidateMapEntries(moduleConfig.Icons); err != nil {
		return fmt.Errorf("failed to validate module icons: %w", err)
	}

	if err := validation.ValidateMapEntries(moduleConfig.Resources); err != nil {
		return fmt.Errorf("failed to validate resources: %w", err)
	}

	if moduleConfig.DefaultCR.IsURL() {
		if moduleConfig.DefaultCR.URL().Scheme != "https" {
			return fmt.Errorf(
				"failed to validate default CR: %w",
				fmt.Errorf(
					"'%s' is not using https scheme: %w",
					moduleConfig.DefaultCR.String(),
					commonerrors.ErrInvalidOption,
				),
			)
		}
	} else {
		if !moduleConfig.DefaultCR.IsEmpty() && strings.HasPrefix(moduleConfig.DefaultCR.String(), "/") {
			return fmt.Errorf("failed to validate default CR: must not be an absolute path: %w",
				commonerrors.ErrInvalidOption)
		}
	}

	if err := ValidateAssociatedResources(moduleConfig.AssociatedResources); err != nil {
		return fmt.Errorf("failed to validate associated resources: %w", err)
	}

	if err := ValidateManager(moduleConfig.Manager); err != nil {
		return fmt.Errorf("failed to validate manager: %w", err)
	}

	return nil
}

func ValidateAssociatedResources(resources []*metav1.GroupVersionKind) error {
	for _, resource := range resources {
		if err := validation.ValidateGvk(resource.Group, resource.Version, resource.Kind); err != nil {
			return fmt.Errorf("GVK is invalid: %w", err)
		}
	}
	return nil
}

func ValidateManager(manager *contentprovider.Manager) error {
	if manager == nil {
		return nil
	}

	if manager.Name == "" {
		return fmt.Errorf("name must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if err := validation.ValidateGvk(manager.Group, manager.Version, manager.Kind); err != nil {
		return fmt.Errorf("GVK is invalid: %w", err)
	}

	if manager.Namespace != "" {
		if err := validation.ValidateNamespace(manager.Namespace); err != nil {
			return fmt.Errorf("namespace is invalid: %w", err)
		}
	}

	return nil
}

func ParseModuleConfig(configFilePath string, fileSystem FileSystem) (*contentprovider.ModuleConfig, error) {
	moduleConfigData, err := fileSystem.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read module config file: %w", err)
	}

	moduleConfig := &contentprovider.ModuleConfig{}
	if err := yaml.Unmarshal(moduleConfigData, moduleConfig); err != nil {
		return nil, fmt.Errorf("failed to parse module config file: %w", err)
	}

	return moduleConfig, nil
}
