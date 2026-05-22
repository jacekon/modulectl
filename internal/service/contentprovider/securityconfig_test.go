package contentprovider_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

func Test_SecurityConfig_NewSecurityConfig_ReturnsError_WhenYamlConverterIsNil(t *testing.T) {
	_, err := contentprovider.NewSecurityConfig(nil)

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	require.Contains(t, err.Error(), "yamlConverter")
}

func Test_SecurityConfig_GetDefaultContent_ReturnsError_WhenArgsIsNil(t *testing.T) {
	svc, _ := contentprovider.NewSecurityConfig(&objectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(nil)

	require.ErrorIs(t, err, contentprovider.ErrInvalidArg)
	require.Empty(t, result)
	require.Contains(t, err.Error(), "args")
}

func Test_SecurityConfig_GetDefaultContent_ReturnsError_WhenModuleNameArgMissing(t *testing.T) {
	svc, _ := contentprovider.NewSecurityConfig(&objectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(types.KeyValueArgs{})

	require.ErrorIs(t, err, contentprovider.ErrMissingArg)
	require.Empty(t, result)
	require.Contains(t, err.Error(), "moduleName")
}

func Test_SecurityConfig_GetDefaultContent_ReturnsError_WhenModuleNameArgIsEmpty(t *testing.T) {
	svc, _ := contentprovider.NewSecurityConfig(&objectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(types.KeyValueArgs{contentprovider.ArgModuleName: ""})

	require.ErrorIs(t, err, contentprovider.ErrInvalidArg)
	require.Empty(t, result)
	require.Contains(t, err.Error(), "moduleName")
}

func Test_SecurityConfig_GetDefaultContent_ReturnsConvertedContent(t *testing.T) {
	svc, _ := contentprovider.NewSecurityConfig(&objectToYAMLConverterStub{})

	result, err := svc.GetDefaultContent(types.KeyValueArgs{contentprovider.ArgModuleName: "module-name"})

	require.NoError(t, err)
	require.Equal(t, convertedContent, result)
}

func Test_SecurityConfig_GetSecurityConfig_Structure(t *testing.T) {
	svc, _ := contentprovider.NewSecurityConfig(&objectToYAMLConverterCapture{})

	_, err := svc.GetDefaultContent(types.KeyValueArgs{contentprovider.ArgModuleName: "test-module"})

	require.NoError(t, err)
}

// Test Stubs

type objectToYAMLConverterStub struct{}

const convertedContent = "content"

func (o *objectToYAMLConverterStub) ConvertToYaml(_ any) string {
	return convertedContent
}

type objectToYAMLConverterCapture struct {
	capturedConfig any
}

func (o *objectToYAMLConverterCapture) ConvertToYaml(obj any) string {
	o.capturedConfig = obj
	// Verify the structure
	config, ok := obj.(contentprovider.SecurityScanConfig)
	if !ok {
		return "error: not a SecurityScanConfig"
	}

	if config.ModuleName == "" {
		return "error: empty module name"
	}

	if len(config.BDBA) != 2 {
		return "error: expected 2 BDBA images"
	}

	return "valid-config"
}
