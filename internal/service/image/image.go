package image

import (
	"errors"
	"fmt"
	"strings"

	"github.com/distribution/reference"
)

const (
	latestTag         = "latest"
	mainTag           = "main"
	minImageRefLength = 3
	maxImageRefLength = 256
)

var (
	ErrEmptyImageURL       = errors.New("empty image URL")
	ErrImageNameExtraction = errors.New("could not extract image name")
	ErrNoTagOrDigest       = errors.New("no tag or digest found")
	ErrMissingImageTag     = errors.New("image is missing a tag")
	ErrDisallowedTag       = errors.New("image tag is disallowed (latest/main)")
)

type ImageInfo struct {
	Name    string
	Tag     string
	Digest  string
	FullURL string
}

// IsImageReferenceCandidate checks if the provided string is a valid candidate for an image reference.
func IsImageReferenceCandidate(value string) bool {
	if len(value) < minImageRefLength || len(value) > maxImageRefLength {
		return false
	}

	// Exclude values that look like URLs (e.g., contain "://") as docker image references do not include schemes.
	if strings.Contains(value, "://") {
		return false
	}

	hasTagOrDigest := false
	for _, c := range value {
		switch c {
		case ':', '@':
			hasTagOrDigest = true
		case ' ', '\t', '\n', '\r':
			return false
		}
	}

	return hasTagOrDigest
}

// ValidateAndParseImageInfo validates the image URL and parses it into ImageInfo.
func ValidateAndParseImageInfo(imageURL string) (*ImageInfo, error) {
	info, err := ParseImageInfo(imageURL)
	if err != nil {
		return nil, err
	}

	if info.Tag == "" && info.Digest == "" {
		return nil, fmt.Errorf("%w: %q", ErrMissingImageTag, imageURL)
	}

	if info.Tag != "" && isMainOrLatestTag(info.Tag) {
		return nil, fmt.Errorf("%w: %q", ErrDisallowedTag, info.Tag)
	}

	return info, nil
}

// ParseImageInfo parses image reference and extracts all components.
func ParseImageInfo(imageURL string) (*ImageInfo, error) {
	if imageURL == "" {
		return nil, fmt.Errorf("failed to parse image reference: %w", ErrEmptyImageURL)
	}

	ref, err := reference.ParseAnyReference(imageURL)
	if err != nil {
		return nil, fmt.Errorf("invalid image reference: %w", err)
	}

	var imageName string
	if named, ok := ref.(reference.Named); ok {
		parts := strings.Split(named.Name(), "/")
		imageName = parts[len(parts)-1]
	} else {
		return nil, fmt.Errorf("failed to extract image name from %s: %w", imageURL, ErrImageNameExtraction)
	}

	info := &ImageInfo{
		Name:    imageName,
		FullURL: imageURL,
	}

	switch refType := ref.(type) {
	case reference.Tagged:
		info.Tag = refType.Tag()
		if digested, ok := refType.(reference.Digested); ok {
			info.Digest = strings.TrimPrefix(digested.Digest().String(), "sha256:")
		}
	case reference.Digested:
		info.Digest = strings.TrimPrefix(refType.Digest().String(), "sha256:")
	default:
		return nil, fmt.Errorf("no tag or digest found in %s: %w", imageURL, ErrNoTagOrDigest)
	}

	return info, nil
}

func isMainOrLatestTag(tag string) bool {
	switch strings.ToLower(tag) {
	case latestTag, mainTag:
		return true
	default:
		return false
	}
}
