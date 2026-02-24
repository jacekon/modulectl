package componentconstructor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/internal/common"
	"github.com/kyma-project/modulectl/internal/common/types/component"
	"github.com/kyma-project/modulectl/internal/service/componentconstructor"
)

func TestService_SetComponentLabel_AddIfNotExists(t *testing.T) {
	service := componentconstructor.NewService()

	// given
	constructor := component.NewConstructor(testModuleName, testModuleVersion)
	require.Empty(t, constructor.Components[0].Labels)

	// when
	service.SetComponentLabel(constructor, "foo", "bar")

	// then
	require.Len(t, constructor.Components[0].Labels, 1)
	assert.Equal(t, "foo", constructor.Components[0].Labels[0].Name)
	assert.Equal(t, "bar", constructor.Components[0].Labels[0].Value)
	assert.Equal(t, common.VersionV1, constructor.Components[0].Labels[0].Version)
}

func TestService_SetComponentLabel_AddToExisting(t *testing.T) {
	service := componentconstructor.NewService()

	// given
	constructor := component.NewConstructor(testModuleName, testModuleVersion)
	require.Empty(t, constructor.Components[0].Labels)

	// when
	service.SetComponentLabel(constructor, "foo1", "bar")

	// then
	require.Len(t, constructor.Components[0].Labels, 1)
	assert.Equal(t, "foo1", constructor.Components[0].Labels[0].Name)
	assert.Equal(t, "bar", constructor.Components[0].Labels[0].Value)
	assert.Equal(t, common.VersionV1, constructor.Components[0].Labels[0].Version)

	// and when
	service.SetComponentLabel(constructor, "foo2", "bar")

	// then
	require.Len(t, constructor.Components[0].Labels, 2)
	assert.Equal(t, "foo1", constructor.Components[0].Labels[0].Name)
	assert.Equal(t, "bar", constructor.Components[0].Labels[0].Value)
	assert.Equal(t, common.VersionV1, constructor.Components[0].Labels[0].Version)
	assert.Equal(t, "foo2", constructor.Components[0].Labels[1].Name)
	assert.Equal(t, "bar", constructor.Components[0].Labels[1].Value)
	assert.Equal(t, common.VersionV1, constructor.Components[0].Labels[1].Version)
}

func TestService_SetComponentLabel_Overwrites(t *testing.T) {
	service := componentconstructor.NewService()

	// given
	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	service.SetComponentLabel(constructor, "foo", "bar")
	require.Len(t, constructor.Components[0].Labels, 1)
	assert.Equal(t, "foo", constructor.Components[0].Labels[0].Name)
	assert.Equal(t, "bar", constructor.Components[0].Labels[0].Value)
	assert.Equal(t, common.VersionV1, constructor.Components[0].Labels[0].Version)

	// when
	service.SetComponentLabel(constructor, "foo", "baz")

	// then
	require.Len(t, constructor.Components[0].Labels, 1)
	assert.Equal(t, "foo", constructor.Components[0].Labels[0].Name)
	assert.Equal(t, "baz", constructor.Components[0].Labels[0].Value)
	assert.Equal(t, common.VersionV1, constructor.Components[0].Labels[0].Version)
}

func TestService_SetComponentLabel_OverwritesWhenMultiple(t *testing.T) {
	service := componentconstructor.NewService()

	// given
	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	service.SetComponentLabel(constructor, "foo1", "bar")
	service.SetComponentLabel(constructor, "foo2", "bar")
	require.Len(t, constructor.Components[0].Labels, 2)
	assert.Equal(t, "foo1", constructor.Components[0].Labels[0].Name)
	assert.Equal(t, "bar", constructor.Components[0].Labels[0].Value)
	assert.Equal(t, common.VersionV1, constructor.Components[0].Labels[0].Version)
	assert.Equal(t, "foo2", constructor.Components[0].Labels[1].Name)
	assert.Equal(t, "bar", constructor.Components[0].Labels[1].Value)
	assert.Equal(t, common.VersionV1, constructor.Components[0].Labels[1].Version)

	// when
	service.SetComponentLabel(constructor, "foo2", "baz")

	// then
	require.Len(t, constructor.Components[0].Labels, 2)
	assert.Equal(t, "foo1", constructor.Components[0].Labels[0].Name)
	assert.Equal(t, "bar", constructor.Components[0].Labels[0].Value)
	assert.Equal(t, common.VersionV1, constructor.Components[0].Labels[0].Version)
	assert.Equal(t, "foo2", constructor.Components[0].Labels[1].Name)
	assert.Equal(t, "baz", constructor.Components[0].Labels[1].Value)
	assert.Equal(t, common.VersionV1, constructor.Components[0].Labels[1].Version)
}

func TestService_SetLabel_OnlyFirstComponent(t *testing.T) {
	service := componentconstructor.NewService()

	// given
	constructor2 := component.NewConstructor("second", "v3.2.1")

	// Set label on second component
	service.SetComponentLabel(constructor2, "foo", "baz")
	require.Len(t, constructor2.Components[0].Labels, 1)
	assert.Equal(t, "foo", constructor2.Components[0].Labels[0].Name)
	assert.Equal(t, "baz", constructor2.Components[0].Labels[0].Value)
	assert.Equal(t, common.VersionV1, constructor2.Components[0].Labels[0].Version)

	// when
	constructor := component.NewConstructor(testModuleName, testModuleVersion)
	constructor.Components = append(constructor.Components, constructor2.Components...)
	service.SetComponentLabel(constructor, "foo", "bar")

	// then
	require.Len(t, constructor.Components, 2)

	// label on second component should not be overridden
	require.Len(t, constructor.Components[1].Labels, 1)
	assert.Equal(t, "foo", constructor.Components[1].Labels[0].Name)
	assert.Equal(t, "baz", constructor.Components[1].Labels[0].Value)
	assert.Equal(t, common.VersionV1, constructor.Components[1].Labels[0].Version)

	// label on first component should be set
	require.Len(t, constructor.Components[0].Labels, 1)
	assert.Equal(t, "foo", constructor.Components[0].Labels[0].Name)
	assert.Equal(t, "bar", constructor.Components[0].Labels[0].Value)
	assert.Equal(t, common.VersionV1, constructor.Components[0].Labels[0].Version)
}

func TestService_SetResponsiblesLabel(t *testing.T) {
	service := componentconstructor.NewService()

	// given
	constructor := component.NewConstructor(testModuleName, testModuleVersion)
	require.Empty(t, constructor.Components[0].Labels)

	// when
	service.SetResponsiblesLabel(constructor, "test-team")

	// then
	require.Len(t, constructor.Components[0].Labels, 1)

	label := constructor.Components[0].Labels[0]
	assert.Equal(t, common.ResponsiblesLabelKey, label.Name)
	assert.Equal(t, common.VersionV1, label.Version)

	// Verify the value is stored as a slice
	responsiblesValue, ok := label.Value.([]component.ResponsibleEntry)
	require.True(t, ok, "label value should be []component.ResponsibleEntry, got %T", label.Value)
	require.Len(t, responsiblesValue, 1)

	responsible := responsiblesValue[0]
	assert.Equal(t, common.GitHubHostname, responsible.GitHubHostname)
	assert.Equal(t, "test-team", responsible.TeamName)
	assert.Equal(t, common.ResponsibleTypeGitHubTeam, responsible.Type)
}

func TestService_SetResponsiblesLabel_WithDifferentTeam(t *testing.T) {
	service := componentconstructor.NewService()

	// given
	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	// when
	service.SetResponsiblesLabel(constructor, "another-team")

	// then
	require.Len(t, constructor.Components[0].Labels, 1)

	label := constructor.Components[0].Labels[0]
	responsiblesValue, ok := label.Value.([]component.ResponsibleEntry)
	require.True(t, ok, "label value should be []component.ResponsibleEntry")
	assert.Equal(t, "another-team", responsiblesValue[0].TeamName)
}

func TestService_SetResponsiblesLabel_AndSecurityScanLabel(t *testing.T) {
	service := componentconstructor.NewService()

	// given
	constructor := component.NewConstructor(testModuleName, testModuleVersion)

	// when
	service.SetResponsiblesLabel(constructor, "test-team")
	service.SetComponentLabel(constructor, common.SecurityScanLabelKey, common.SecurityScanEnabledValue)

	// then
	require.Len(t, constructor.Components[0].Labels, 2)

	// Verify responsibles label
	responsiblesLabel := constructor.Components[0].Labels[0]
	assert.Equal(t, common.ResponsiblesLabelKey, responsiblesLabel.Name)
	responsiblesValue, ok := responsiblesLabel.Value.([]component.ResponsibleEntry)
	require.True(t, ok, "label value should be []component.ResponsibleEntry")
	assert.Equal(t, "test-team", responsiblesValue[0].TeamName)

	// Verify security scan label
	securityLabel := constructor.Components[0].Labels[1]
	assert.Equal(t, common.SecurityScanLabelKey, securityLabel.Name)
	assert.Equal(t, common.SecurityScanEnabledValue, securityLabel.Value)
}
