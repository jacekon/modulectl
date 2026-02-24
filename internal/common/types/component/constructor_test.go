package component_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/internal/common"
	"github.com/kyma-project/modulectl/internal/common/types/component"
	"github.com/kyma-project/modulectl/internal/service/image"
)

func TestNewConstructor(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	require.NotNil(t, constructor)
	require.Len(t, constructor.Components, 1)

	moduleComponent := constructor.Components[0]
	require.Equal(t, "test-component", moduleComponent.Name)
	require.Equal(t, "1.0.0", moduleComponent.Version)
}

func TestConstructor_Initialize(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	require.Len(t, constructor.Components, 1)
	moduleComponent := constructor.Components[0]
	require.Equal(t, "test-component", moduleComponent.Name)
	require.Equal(t, "1.0.0", moduleComponent.Version)
	require.Equal(t, common.ProviderName, moduleComponent.Provider.Name)
	require.Len(t, moduleComponent.Provider.Labels, 1)
	providerLabel := moduleComponent.Provider.Labels[0]
	require.Equal(t, common.BuiltByLabelKey, providerLabel.Name)
	require.Equal(t, common.BuiltByLabelValue, providerLabel.Value)
	require.Equal(t, common.VersionV1, providerLabel.Version)
	require.Empty(t, moduleComponent.Resources)
	require.Empty(t, moduleComponent.Sources)
}

func TestConstructor_AddGitSource(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	constructor.AddGitSource("https://github.com/kyma-project/modulectl", "abc123def456")

	require.Len(t, constructor.Components[0].Sources, 1)
	source := constructor.Components[0].Sources[0]
	require.Equal(t, common.OCMIdentityName, source.Name)
	require.Equal(t, component.GithubSourceType, source.Type)
	require.Equal(t, "1.0.0", source.Version)
	require.Empty(t, source.Labels)
	require.Equal(t, component.GithubAccessType, source.Access.Type)
	require.Equal(t, "https://github.com/kyma-project/modulectl", source.Access.RepoUrl)
	require.Equal(t, "abc123def456", source.Access.Commit)
}

func TestConstructor_AddLabel(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	initialLabelCount := len(constructor.Components[0].Labels)

	constructor.AddLabel("test-key", "test-value", common.VersionV1)

	require.Len(t, constructor.Components[0].Labels, initialLabelCount+1)

	var addedLabel *component.Label
	for _, label := range constructor.Components[0].Labels {
		if label.Name == "test-key" {
			addedLabel = &label
			break
		}
	}

	require.NotNil(t, addedLabel, "added label not found")
	require.Equal(t, "test-key", addedLabel.Name)
	require.Equal(t, "test-value", addedLabel.Value)
	require.Equal(t, common.VersionV1, addedLabel.Version)
}

func TestConstructor_AddLabel_Multiple(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	labels := []struct {
		key, value, version string
	}{
		{"environment", "production", common.VersionV1},
		{"team", "platform", common.VersionV2},
		{"criticality", "high", common.VersionV1},
		{"region", "us-east-1", common.VersionV1},
	}

	for _, label := range labels {
		constructor.AddLabel(label.key, label.value, label.version)
	}

	require.Len(t, constructor.Components[0].Labels, len(labels))

	for _, expectedLabel := range labels {
		found := false
		for _, actualLabel := range constructor.Components[0].Labels {
			if actualLabel.Name == expectedLabel.key &&
				actualLabel.Value == expectedLabel.value &&
				actualLabel.Version == expectedLabel.version {
				found = true
				break
			}
		}
		require.True(t, found, "label with key %s not found in component labels", expectedLabel.key)
	}
}

func TestConstructor_AddImageAsResource(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	imageInfo := &image.ImageInfo{
		Name:    "test-image",
		Tag:     "1.0.0",
		Digest:  "sha256:abc123",
		FullURL: "registry.io/test-image:1.0.0",
	}

	constructor.AddImageAsResource([]*image.ImageInfo{imageInfo})

	require.Len(t, constructor.Components[0].Resources, 1)
	resource := constructor.Components[0].Resources[0]
	require.Equal(t, component.OCIArtifactResourceType, resource.Type)
	require.Equal(t, component.OCIArtifactResourceRelation, resource.Relation)
	require.Len(t, resource.Labels, 1)

	expectedLabelName := common.SecScanBaseLabelKey + "/" + common.TypeLabelKey
	require.Equal(t, expectedLabelName, resource.Labels[0].Name)
	require.Equal(t, common.ThirdPartyImageLabelValue, resource.Labels[0].Value)
	require.Equal(t, common.OCMVersion, resource.Labels[0].Version)
	require.Equal(t, component.OCIArtifactAccessType, resource.Access.Type)
	require.Equal(t, imageInfo.FullURL, resource.Access.ImageReference)
}

func TestConstructor_AddImageAsResource_Multiple(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	imageInfos := []*image.ImageInfo{
		{
			Name:    "image1",
			Tag:     "1.0.0",
			Digest:  "sha256:abc123",
			FullURL: "registry.io/image1:1.0.0",
		},
		{
			Name:    "image2",
			Tag:     "2.0.0",
			Digest:  "sha256:def456",
			FullURL: "registry.io/image2:2.0.0",
		},
	}

	constructor.AddImageAsResource(imageInfos)

	require.Len(t, constructor.Components[0].Resources, 2)

	for i, resource := range constructor.Components[0].Resources {
		require.Equal(t, imageInfos[i].FullURL, resource.Access.ImageReference, "resource %d: image reference mismatch",
			i)
	}
}

func TestConstructor_AddFileResource_ModuleTemplate(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	err := constructor.AddFileResource(common.ModuleTemplateResourceName, "/path/to/file.yaml")
	require.NoError(t, err)

	require.Len(t, constructor.Components[0].Resources, 1)
	resource := constructor.Components[0].Resources[0]
	require.Equal(t, common.ModuleTemplateResourceName, resource.Name)
	require.Equal(t, component.PlainTextResourceType, resource.Type)
	require.Equal(t, "1.0.0", resource.Version)
	require.Equal(t, component.FileResourceInput, resource.Input.Type)
	require.Equal(t, "/path/to/file.yaml", resource.Input.Path)
}

func TestConstructor_AddFileResource_RawManifest(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	err := constructor.AddFileResource(common.RawManifestResourceName, "/path/to/manifest.yaml")
	require.NoError(t, err)

	require.Len(t, constructor.Components[0].Resources, 1)
	resource := constructor.Components[0].Resources[0]
	require.Equal(t, common.RawManifestResourceName, resource.Name)
	require.Equal(t, component.DirectoryTreeResourceType, resource.Type)
	require.Equal(t, "1.0.0", resource.Version)
	require.Equal(t, component.DirectoryInputType, resource.Input.Type)
	require.Equal(t, "/path/to", resource.Input.Path)
	require.Equal(t, []string{"manifest.yaml"}, resource.Input.IncludeFiles)
	require.False(t, resource.Input.Compress)
}

func TestConstructor_AddFileResource_DefaultCR(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	err := constructor.AddFileResource(common.DefaultCRResourceName, "/path/to/cr.yaml")
	require.NoError(t, err)

	require.Len(t, constructor.Components[0].Resources, 1)
	resource := constructor.Components[0].Resources[0]
	require.Equal(t, common.DefaultCRResourceName, resource.Name)
	require.Equal(t, component.DirectoryTreeResourceType, resource.Type)
	require.Equal(t, "1.0.0", resource.Version)
	require.Equal(t, component.DirectoryInputType, resource.Input.Type)
	require.Equal(t, "/path/to", resource.Input.Path)
	require.Equal(t, []string{"cr.yaml"}, resource.Input.IncludeFiles)
	require.False(t, resource.Input.Compress)
}

func TestConstructor_AddFileResource_UnknownResourceName(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	err := constructor.AddFileResource("unknown-resource", "/path/to/file.yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown resource name: unknown-resource")
	require.Empty(t, constructor.Components[0].Resources)
}

func TestConstructor_AddFileResource_ModuleTemplate_RelativePath(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	err := constructor.AddFileResource(common.ModuleTemplateResourceName, "relative/path/file.yaml")
	require.NoError(t, err)

	require.Len(t, constructor.Components[0].Resources, 1)
	resource := constructor.Components[0].Resources[0]
	require.Equal(t, common.ModuleTemplateResourceName, resource.Name)
	require.Equal(t, component.PlainTextResourceType, resource.Type)
	require.True(t, filepath.IsAbs(resource.Input.Path), "path should be converted to absolute")
	require.Contains(t, resource.Input.Path, "relative/path/file.yaml")
}

func TestConstructor_AddFileResource_RawManifest_RelativePath(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	err := constructor.AddFileResource(common.RawManifestResourceName, "relative/path/manifest.yaml")
	require.NoError(t, err)

	require.Len(t, constructor.Components[0].Resources, 1)
	resource := constructor.Components[0].Resources[0]
	require.Equal(t, common.RawManifestResourceName, resource.Name)
	require.Equal(t, component.DirectoryTreeResourceType, resource.Type)
	require.True(t, filepath.IsAbs(resource.Input.Path), "directory path should be converted to absolute")
	require.Contains(t, resource.Input.Path, "relative/path")
	require.Equal(t, "manifest.yaml", resource.Input.IncludeFiles[0])
}

func TestConstructor_AddFileResource_DefaultCR_CurrentDirectory(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	err := constructor.AddFileResource(common.DefaultCRResourceName, "./cr.yaml")
	require.NoError(t, err)

	require.Len(t, constructor.Components[0].Resources, 1)
	resource := constructor.Components[0].Resources[0]
	require.Equal(t, common.DefaultCRResourceName, resource.Name)
	require.Equal(t, component.DirectoryTreeResourceType, resource.Type)
	require.True(t, filepath.IsAbs(resource.Input.Path), "directory path should be converted to absolute")
	require.Equal(t, "cr.yaml", resource.Input.IncludeFiles[0])
}

func TestConstructor_AddFileResource_RawManifest_ParentDirectory(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	err := constructor.AddFileResource(common.RawManifestResourceName, "../manifest.yaml")
	require.NoError(t, err)

	require.Len(t, constructor.Components[0].Resources, 1)
	resource := constructor.Components[0].Resources[0]
	require.Equal(t, common.RawManifestResourceName, resource.Name)
	require.Equal(t, component.DirectoryTreeResourceType, resource.Type)
	require.True(t, filepath.IsAbs(resource.Input.Path), "directory path should be converted to absolute")
	require.Equal(t, "manifest.yaml", resource.Input.IncludeFiles[0])
}

func TestConstructor_AddLabel_WithStringValue(t *testing.T) {
	constructor := component.NewConstructor("test-component", "1.0.0")

	// Test backward compatibility - string values should still work
	constructor.AddLabel("string-label", "simple-string-value", common.VersionV1)

	require.Len(t, constructor.Components[0].Labels, 1)

	label := constructor.Components[0].Labels[0]
	require.Equal(t, "string-label", label.Name)
	require.Equal(t, "simple-string-value", label.Value)
}
