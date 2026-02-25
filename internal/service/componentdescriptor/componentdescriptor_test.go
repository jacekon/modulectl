package componentdescriptor_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	ociartifacttypes "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"

	"github.com/kyma-project/modulectl/internal/service/componentdescriptor"
)

func Test_InitializeComponentDescriptor_ReturnsCorrectDescriptor(t *testing.T) {
	moduleName := "github.com/test-module"
	moduleVersion := "0.0.1"
	team := "test-team"
	descriptor, err := componentdescriptor.InitializeComponentDescriptor(moduleName, moduleVersion, team, true)
	expectedProviderLabel := json.RawMessage(`"modulectl"`)

	require.NoError(t, err)
	require.Equal(t, moduleName, descriptor.GetName())
	require.Equal(t, moduleVersion, descriptor.GetVersion())
	require.Equal(t, "v2", descriptor.Metadata.ConfiguredVersion)
	require.Equal(t, ocmv1.ProviderName("kyma-project.io"), descriptor.Provider.Name)
	require.Len(t, descriptor.Provider.Labels, 1)
	require.Equal(t, "kyma-project.io/built-by", descriptor.Provider.Labels[0].Name)
	require.Equal(t, expectedProviderLabel, descriptor.Provider.Labels[0].Value)
	require.Equal(t, "v1", descriptor.Provider.Labels[0].Version)
	require.Empty(t, descriptor.Resources)

	// Verify responsibles label
	require.Len(t, descriptor.Labels, 2) // responsibles + security scan
	require.Equal(t, "cloud.gardener.cnudie/responsibles", descriptor.Labels[0].Name)
	require.Equal(t, "v1", descriptor.Labels[0].Version)
}

func Test_InitializeComponentDescriptor_ReturnsErrWhenInvalidName(t *testing.T) {
	moduleName := "test-module"
	moduleVersion := "0.0.1"
	team := "test-team"
	_, err := componentdescriptor.InitializeComponentDescriptor(moduleName, moduleVersion, team, true)

	expectedError := errors.New("failed to validate component descriptor")
	require.ErrorContains(t, err, expectedError.Error())
}

func Test_InitializeComponentDescriptor_LabelCreationFails(t *testing.T) {
	badName := string([]byte{0x7f})
	moduleVersion := "0.0.1"
	team := "test-team"
	_, err := componentdescriptor.InitializeComponentDescriptor(badName, moduleVersion, team, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to validate component descriptor")
}

func Test_InitializeComponentDescriptor_EmptyVersion_ReturnsError(t *testing.T) {
	moduleName := "github.com/test-module"
	moduleVersion := ""
	team := "test-team"
	_, err := componentdescriptor.InitializeComponentDescriptor(moduleName, moduleVersion, team, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to validate component descriptor")
}

func Test_InitializeComponentDescriptor_WithSecurityScanEnabled_AddsSecurityLabel(t *testing.T) {
	moduleName := "github.com/test-module"
	moduleVersion := "0.0.1"
	team := "test-team"
	descriptor, err := componentdescriptor.InitializeComponentDescriptor(moduleName, moduleVersion, team, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Labels, 2) // responsibles + security scan
	require.Equal(t, "cloud.gardener.cnudie/responsibles", descriptor.Labels[0].Name)
	require.Equal(t, "v1", descriptor.Labels[0].Version)
	require.Equal(t, "security.kyma-project.io/scan", descriptor.Labels[1].Name)
	require.Equal(t, `"enabled"`, string(descriptor.Labels[1].Value))
	require.Equal(t, "v1", descriptor.Labels[1].Version)
}

func Test_InitializeComponentDescriptor_WithSecurityScanDisabled_DoesNotAddSecurityLabel(t *testing.T) {
	moduleName := "github.com/test-module"
	moduleVersion := "0.0.1"
	team := "test-team"
	descriptor, err := componentdescriptor.InitializeComponentDescriptor(moduleName, moduleVersion, team, false)

	require.NoError(t, err)
	require.Empty(t, descriptor.Labels) // no labels when security scan is disabled
}

func Test_InitializeComponentDescriptor_AddsResponsiblesLabel(t *testing.T) {
	moduleName := "github.com/test-module"
	moduleVersion := "0.0.1"
	team := "my-awesome-team"
	descriptor, err := componentdescriptor.InitializeComponentDescriptor(moduleName, moduleVersion, team, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Labels, 2) // responsibles + security scan labels

	responsiblesLabel := descriptor.Labels[0]
	require.Equal(t, "cloud.gardener.cnudie/responsibles", responsiblesLabel.Name)
	require.Equal(t, "v1", responsiblesLabel.Version)

	// Parse the label value to verify structure
	var responsibles []map[string]any
	err = json.Unmarshal(responsiblesLabel.Value, &responsibles)
	require.NoError(t, err)
	require.Len(t, responsibles, 1)

	responsible := responsibles[0]
	require.Equal(t, "github.tools.sap", responsible["github_hostname"])
	require.Equal(t, "my-awesome-team", responsible["teamname"])
	require.Equal(t, "githubTeam", responsible["type"])
}

func TestAddImagesToOcmDescriptor_WhenCalledWithValidImages_AppendsResources(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine:3.15.4",
		"nginx:1.21.0",
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 2)

	resource1 := descriptor.Resources[0]
	require.Equal(t, "alpine", resource1.Name)
	require.Equal(t, "3.15.4", resource1.Version)
	require.Equal(t, ociartifacttypes.TYPE, resource1.Type)
	require.Equal(t, ocmv1.ExternalRelation, resource1.Relation)
	require.Len(t, resource1.Labels, 1)
	require.Equal(t, "scan.security.kyma-project.io/type", resource1.Labels[0].Name)

	var labelValue1 string
	err = json.Unmarshal(resource1.Labels[0].Value, &labelValue1)
	require.NoError(t, err)
	require.Equal(t, "third-party-image", labelValue1)

	resource2 := descriptor.Resources[1]
	require.Equal(t, "nginx", resource2.Name)
	require.Equal(t, "1.21.0", resource2.Version)
	require.Equal(t, ociartifacttypes.TYPE, resource2.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithComplexRegistryPath_AppendsResource(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"europe-docker.pkg.dev/kyma-project/prod/external/istio/proxyv2:1.25.3-distroless",
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "proxyv2", resource.Name)
	require.Equal(t, "1.25.3-distroless", resource.Version)
	require.Equal(t, ociartifacttypes.TYPE, resource.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithGcrImage_AppendsResource(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"gcr.io/kubebuilder/kube-rbac-proxy:v0.13.1",
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "kube-rbac-proxy", resource.Name)
	require.Equal(t, "v0.13.1", resource.Version)
	require.Equal(t, ociartifacttypes.TYPE, resource.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithInvalidImage_ReturnsError(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{"invalid-image-no-tag"}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.Error(t, err)
	require.Contains(t, err.Error(), "no tag or digest found")
}

func TestAddImagesToOcmDescriptor_WhenCalledWithEmptyImageList_DoesNothing(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Empty(t, descriptor.Resources)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithRegistryPortImage_AppendsResource(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"localhost:5000/myimage:v1.0.0",
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "myimage", resource.Name)
	require.Equal(t, "v1.0.0", resource.Version)
	require.Equal(t, ociartifacttypes.TYPE, resource.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithDockerHubImage_AppendsResource(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"istio/proxyv2:1.19.0",
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "proxyv2", resource.Name)
	require.Equal(t, "1.19.0", resource.Version)
	require.Equal(t, ociartifacttypes.TYPE, resource.Type)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithMultipleImages_CreatesCorrectLabels(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine:3.15.4",
		"nginx:1.21.0",
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 2)

	for _, resource := range descriptor.Resources {
		require.Len(t, resource.Labels, 1)
		require.Equal(t, "scan.security.kyma-project.io/type", resource.Labels[0].Name)
		require.Equal(t, "v1", resource.Labels[0].Version)
		require.NotNil(t, resource.Access)
		require.Equal(t, ociartifact.Type, resource.Access.GetType())
	}
}

func TestAddImagesToOcmDescriptor_WhenCalledWithDigestImage_AppendsResourceWithConvertedVersion(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine@sha256:abcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab",
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "alpine-abcd1234", resource.Name)
	require.Equal(t, "0.0.0+sha256.abcd12345678", resource.Version)
	require.Equal(t, ociartifacttypes.TYPE, resource.Type)

	access, ok := resource.Access.(*ociartifact.AccessSpec)
	if !ok {
		t.Fatalf("expected AccessSpec type, got %T", resource.Access)
	}
	require.Equal(
		t,
		"alpine@sha256:abcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab",
		access.ImageReference,
	)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithMalformedImage_ReturnsError(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"",
		"alpine:",
		"alpine@",
	}

	for _, img := range images {
		err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, []string{img}, true)
		require.Error(t, err)
	}
}

func TestAddImagesToOcmDescriptor_WhenCalledWithExistingResources_AppendsToExisting(t *testing.T) {
	existingResource := compdesc.Resource{
		ResourceMeta: compdesc.ResourceMeta{
			Type:     "existing-type",
			Relation: ocmv1.LocalRelation,
			ElementMeta: compdesc.ElementMeta{
				Name:    "existing",
				Version: "1.0.0",
			},
		},
		Access: ociartifact.New("existing:1.0.0"),
	}

	descriptor := createDescriptorWithResource(existingResource)
	images := []string{
		"alpine:3.15.4",
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 2)
	require.Equal(t, "existing", descriptor.Resources[0].Name)
	require.Equal(t, "alpine", descriptor.Resources[1].Name)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithNilDescriptor_Panics(t *testing.T) {
	var descriptor *compdesc.ComponentDescriptor
	images := []string{"alpine:3.15.4"}

	require.Panics(t, func() {
		_ = componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)
	})
}

func TestAddImagesToOcmDescriptor_WhenCalledWithImageWithoutTag_ReturnsError(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine",
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.Error(t, err)
	require.Contains(t, err.Error(), "no tag or digest found in alpine")
}

func TestAddImagesToOcmDescriptor_WhenCalledWithValidImageAfterError_StopsProcessing(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine:3.15.4",
		"",
		"nginx:1.21.0",
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.Error(t, err)
	require.Len(t, descriptor.Resources, 1)
	require.Equal(t, "alpine", descriptor.Resources[0].Name)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithVariousTagFormats_AppendsResourcesWithCorrectVersions(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"myapp:v1.0.0",
		"myapp:1.0.0",
		"myapp:123",
		"myapp:feature-branch",
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 4)

	expectedVersions := []string{
		"v1.0.0",
		"1.0.0",
		"0.0.0-123",
		"0.0.0-feature-branch",
	}

	for i, expected := range expectedVersions {
		require.Equal(t, expected, descriptor.Resources[i].Version,
			"Resource %d version mismatch", i)
	}
}

func TestAddImagesToOcmDescriptor_WhenCalledAfterDefaults_MaintainsDescriptorValidity(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine:3.15.4",
		"nginx:1.21.0",
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)

	err = compdesc.Validate(descriptor)
	require.NoError(t, err)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithImageWithMultipleSlashes_ExtractsCorrectName(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"registry.example.com/team/project/subproject/app:v1.0.0",
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)

	resource := descriptor.Resources[0]
	require.Equal(t, "app", resource.Name)
	require.Equal(t, "v1.0.0", resource.Version)
}

func TestAddImagesToOcmDescriptor_WhenCalledWithShortDigest_ReturnsError(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine@sha256:short",
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid reference format")
}

func TestAddOciArtifactsToDescriptor_WhenImageListIsEmpty_LeavesResourcesUnchanged(t *testing.T) {
	descriptor := createEmptyDescriptor()
	descriptor.Resources = append(descriptor.Resources, compdesc.Resource{
		ResourceMeta: compdesc.ResourceMeta{
			Type:     ociartifacttypes.TYPE,
			Relation: ocmv1.ExternalRelation,
			ElementMeta: compdesc.ElementMeta{
				Name:    "existing",
				Version: "1.0.0",
				Labels: []ocmv1.Label{
					{
						Name:    "scan.security.kyma-project.io/type",
						Value:   json.RawMessage(`"third-party-image"`),
						Version: "v1",
					},
				},
			},
		},
		Access: ociartifact.New("existing:1.0.0"),
	})
	images := []string{}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)
	require.Equal(t, "existing", descriptor.Resources[0].Name)
}

func TestAddOciArtifactsToDescriptor_WhenDuplicateImages_DoesNotAddDuplicates(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{"alpine:3.15.4", "alpine:3.15.4"}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)
	require.Equal(t, "alpine", descriptor.Resources[0].Name)
}

func TestAddOciArtifactsToDescriptor_WhenManyInvalidAndOneValidImage_AddsOnlyValid(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{"", ":", "notvalid", "alpine:3.15.4"}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.Error(t, err)
	require.Empty(t, descriptor.Resources)
}

func TestAddOciArtifactsToDescriptor_WhenAllImagesValid_AddsAll(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{"alpine:3.15.4", "nginx:1.21.0"}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 2)
	names := []string{descriptor.Resources[0].Name, descriptor.Resources[1].Name}
	require.Contains(t, names, "alpine")
	require.Contains(t, names, "nginx")
}

func TestAddOciArtifactsToDescriptor_WhenDescriptorHasInvalidResource_ReturnsValidationError(t *testing.T) {
	descriptor := createEmptyDescriptor()
	descriptor.Resources = append(descriptor.Resources, compdesc.Resource{
		ResourceMeta: compdesc.ResourceMeta{
			ElementMeta: compdesc.ElementMeta{
				Name:    "invalid",
				Version: "1.0.0",
			},
		},
	})
	images := []string{"alpine:3.15.4"}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to validate component descriptor")
}

func TestAddOciArtifactsToDescriptor_WhenImagesResultInDuplicateResources_DoesNotAddDuplicates(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{
		"alpine:3.15.4",
		"library/alpine:3.15.4",
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 1)
}

func TestAddOciArtifactsToDescriptor_WhenCompdescValidateFailsAfterResourceAddition_ReturnsError(t *testing.T) {
	descriptor := createEmptyDescriptor()
	descriptor.SetName("invalid name with spaces")
	images := []string{"alpine:3.15.4"}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to validate component descriptor")
}

func TestAddOciArtifactsToDescriptor_WhenLargeNumberOfImages_AddsAllResources(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := make([]string, 0, 50)
	for i := range 50 {
		images = append(images, fmt.Sprintf("alpine:3.15.%d", i))
	}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 50)
	for i := range 50 {
		require.Equal(t, fmt.Sprintf("3.15.%d", i), descriptor.Resources[i].Version)
	}
}

func TestAddOciArtifactsToDescriptor_WhenImagesContainEmptyOrWhitespace_SkipsAndReturnsError(t *testing.T) {
	descriptor := createEmptyDescriptor()
	images := []string{"alpine:3.15.4", "", "   "}

	err := componentdescriptor.AddOciArtifactsToDescriptor(descriptor, images, true)

	require.Error(t, err)
	require.Len(t, descriptor.Resources, 1)
	require.Equal(t, "alpine", descriptor.Resources[0].Name)
}

// Test helper functions.
func createEmptyDescriptor() *compdesc.ComponentDescriptor {
	descriptor := &compdesc.ComponentDescriptor{
		ComponentSpec: compdesc.ComponentSpec{
			ObjectMeta: ocmv1.ObjectMeta{
				Name:     "kyma-project.io/module/telemetry",
				Version:  "1.0.0",
				Provider: ocmv1.Provider{Name: "kyma-project.io"},
			},
			Resources: []compdesc.Resource{},
		},
	}
	compdesc.DefaultResources(descriptor)
	return descriptor
}

func createDescriptorWithResource(resource compdesc.Resource) *compdesc.ComponentDescriptor {
	descriptor := &compdesc.ComponentDescriptor{
		ComponentSpec: compdesc.ComponentSpec{
			ObjectMeta: ocmv1.ObjectMeta{
				Name:     "kyma-project.io/module/telemetry",
				Version:  "1.0.0",
				Provider: ocmv1.Provider{Name: "kyma-project.io"},
			},
			Resources: []compdesc.Resource{resource},
		},
	}
	compdesc.DefaultResources(descriptor)
	return descriptor
}
