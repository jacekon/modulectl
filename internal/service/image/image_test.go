package image_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/internal/service/image"
)

func TestIsImageReferenceCandidate_ValidImagesWithTagsAndDigests(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid image with tag",
			input:    "nginx:1.20",
			expected: true,
		},
		{
			name:     "valid image with digest",
			input:    "nginx@sha256:abc123",
			expected: true,
		},
		{
			name:     "valid image with registry and tag",
			input:    "registry.io/nginx:1.20",
			expected: true,
		},
		{
			name:     "minimum length with tag",
			input:    "a:b",
			expected: true,
		},
		{
			name:     "minimum length with digest",
			input:    "a@b",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := image.IsImageReferenceCandidate(tt.input)
			require.Equal(
				t,
				tt.expected,
				result,
				"IsImageReferenceCandidate(%q) = %v, want %v",
				tt.input,
				result,
				tt.expected,
			)
		})
	}
}

func TestIsImageReferenceCandidate_InvalidImages(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "too short",
			input:    "ab",
			expected: false,
		},
		{
			name:     "too long",
			input:    strings.Repeat("a", 257),
			expected: false,
		},
		{
			name:     "no tag or digest",
			input:    "nginx",
			expected: false,
		},
		{
			name:     "contains space",
			input:    "nginx :1.20",
			expected: false,
		},
		{
			name:     "contains tab",
			input:    "nginx\t:1.20",
			expected: false,
		},
		{
			name:     "contains newline",
			input:    "nginx\n:1.20",
			expected: false,
		},
		{
			name:     "contains carriage return",
			input:    "nginx\r:1.20",
			expected: false,
		},
		{
			name:     "contains protocol part",
			input:    "http://api-server.myapp.svc.cluster.local:8080",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := image.IsImageReferenceCandidate(tt.input)
			require.Equal(
				t,
				tt.expected,
				result,
				"IsImageReferenceCandidate(%q) = %v, want %v",
				tt.input,
				result,
				tt.expected,
			)
		})
	}
}

func TestParseImageInfo_ValidImagesWithTags(t *testing.T) {
	tests := []struct {
		name           string
		imageURL       string
		expectedName   string
		expectedTag    string
		expectedDigest string
	}{
		{
			name:           "simple image with tag",
			imageURL:       "nginx:1.20",
			expectedName:   "nginx",
			expectedTag:    "1.20",
			expectedDigest: "",
		},
		{
			name:           "registry with image and tag",
			imageURL:       "registry.io/nginx:1.20",
			expectedName:   "nginx",
			expectedTag:    "1.20",
			expectedDigest: "",
		},
		{
			name:           "nested registry path",
			imageURL:       "registry.io/team/nginx:1.20",
			expectedName:   "nginx",
			expectedTag:    "1.20",
			expectedDigest: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := image.ParseImageInfo(tt.imageURL)
			require.NoError(t, err, "ParseImageInfo(%q) returned unexpected error", tt.imageURL)

			assertImageInfo(t, info, tt.expectedName, tt.expectedTag, tt.expectedDigest, tt.imageURL)
		})
	}
}

func TestParseImageInfo_ValidImagesWithDigests(t *testing.T) {
	tests := []struct {
		name           string
		imageURL       string
		expectedName   string
		expectedDigest string
	}{
		{
			name:           "simple image with digest",
			imageURL:       "nginx@sha256:abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			expectedName:   "nginx",
			expectedDigest: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
		{
			name:           "registry with image and digest",
			imageURL:       "registry.io/nginx@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expectedName:   "nginx",
			expectedDigest: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := image.ParseImageInfo(tt.imageURL)
			require.NoError(t, err, "ParseImageInfo(%q) returned unexpected error", tt.imageURL)

			assertImageInfo(t, info, tt.expectedName, "", tt.expectedDigest, tt.imageURL)
		})
	}
}

func TestParseImageInfo_ValidImagesWithTagsAndDigests(t *testing.T) {
	imageURL := "nginx:1.20@sha256:abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	expectedName := "nginx"
	expectedTag := "1.20"
	expectedDigest := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

	info, err := image.ParseImageInfo(imageURL)
	require.NoError(t, err, "ParseImageInfo(%q) returned unexpected error", imageURL)

	assertImageInfo(t, info, expectedName, expectedTag, expectedDigest, imageURL)
}

func TestParseImageInfo_InvalidImages(t *testing.T) {
	tests := []struct {
		name          string
		imageURL      string
		expectedError error
	}{
		{
			name:          "empty image URL",
			imageURL:      "",
			expectedError: image.ErrEmptyImageURL,
		},
		{
			name:          "invalid image reference",
			imageURL:      "invalid:::reference",
			expectedError: nil,
		},
		{
			name:          "no tag or digest",
			imageURL:      "nginx",
			expectedError: image.ErrNoTagOrDigest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := image.ParseImageInfo(tt.imageURL)
			require.Error(t, err, "ParseImageInfo(%q) expected error but got info: %+v", tt.imageURL, info)

			if tt.expectedError != nil {
				require.ErrorIs(
					t,
					err,
					tt.expectedError,
					"ParseImageInfo(%q) error = %v, want %v",
					tt.imageURL,
					err,
					tt.expectedError,
				)
			}
		})
	}
}

func TestValidateAndParseImageInfo_ValidImagesWithAllowedTags(t *testing.T) {
	tests := []struct {
		name           string
		imageURL       string
		expectedName   string
		expectedTag    string
		expectedDigest string
	}{
		{
			name:           "valid tag",
			imageURL:       "nginx:1.20",
			expectedName:   "nginx",
			expectedTag:    "1.20",
			expectedDigest: "",
		},
		{
			name:           "valid digest only",
			imageURL:       "nginx@sha256:abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			expectedName:   "nginx",
			expectedTag:    "",
			expectedDigest: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
		{
			name:           "valid tag and digest",
			imageURL:       "nginx:1.20@sha256:abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			expectedName:   "nginx",
			expectedTag:    "1.20",
			expectedDigest: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := image.ValidateAndParseImageInfo(tt.imageURL)
			require.NoError(t, err, "ValidateAndParseImageInfo(%q) returned unexpected error", tt.imageURL)

			assertImageInfo(t, info, tt.expectedName, tt.expectedTag, tt.expectedDigest, tt.imageURL)
		})
	}
}

func TestValidateAndParseImageInfo_InvalidImagesWithDisallowedTags(t *testing.T) {
	tests := []struct {
		name          string
		imageURL      string
		expectedError error
	}{
		{
			name:          "latest tag",
			imageURL:      "nginx:latest",
			expectedError: image.ErrDisallowedTag,
		},
		{
			name:          "main tag",
			imageURL:      "nginx:main",
			expectedError: image.ErrDisallowedTag,
		},
		{
			name:          "latest tag uppercase",
			imageURL:      "nginx:LATEST",
			expectedError: image.ErrDisallowedTag,
		},
		{
			name:          "main tag mixed case",
			imageURL:      "nginx:Main",
			expectedError: image.ErrDisallowedTag,
		},
		{
			name:          "missing tag and digest - returns parse error first",
			imageURL:      "nginx",
			expectedError: image.ErrNoTagOrDigest,
		},
		{
			name:          "empty image URL",
			imageURL:      "",
			expectedError: image.ErrEmptyImageURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := image.ValidateAndParseImageInfo(tt.imageURL)
			require.Error(t, err, "ValidateAndParseImageInfo(%q) expected error but got info: %+v", tt.imageURL, info)
			require.ErrorIs(
				t,
				err,
				tt.expectedError,
				"ValidateAndParseImageInfo(%q) error = %v, want %v",
				tt.imageURL,
				err,
				tt.expectedError,
			)
		})
	}
}

func TestIsMainOrLatestTag_ValidCases(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected bool
	}{
		{
			name:     "latest lowercase",
			tag:      "latest",
			expected: true,
		},
		{
			name:     "latest uppercase",
			tag:      "LATEST",
			expected: true,
		},
		{
			name:     "main lowercase",
			tag:      "main",
			expected: true,
		},
		{
			name:     "main uppercase",
			tag:      "MAIN",
			expected: true,
		},
		{
			name:     "main mixed case",
			tag:      "Main",
			expected: true,
		},
		{
			name:     "other tag",
			tag:      "v1.20",
			expected: false,
		},
		{
			name:     "empty tag",
			tag:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expected {
				_, err := image.ValidateAndParseImageInfo("nginx:" + tt.tag)
				require.ErrorIs(t, err, image.ErrDisallowedTag, "Expected disallowed tag error for %q", tt.tag)
			} else if tt.tag != "" {
				info, err := image.ValidateAndParseImageInfo("nginx:" + tt.tag)
				require.NoError(t, err, "Expected no error for allowed tag %q", tt.tag)
				require.Equal(t, tt.tag, info.Tag)
			}
		})
	}
}

// Test helper functions.
func assertImageInfo(
	t *testing.T,
	info *image.ImageInfo,
	expectedName, expectedTag, expectedDigest, expectedFullURL string,
) {
	t.Helper()

	require.Equal(t, expectedName, info.Name, "Name mismatch")
	require.Equal(t, expectedTag, info.Tag, "Tag mismatch")
	require.Equal(t, expectedDigest, info.Digest, "Digest mismatch")
	require.Equal(t, expectedFullURL, info.FullURL, "FullURL mismatch")
}
