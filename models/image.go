package imghoard

import (
	"strings"

	"github.com/mikibot/imghoard/services/snowflake"
)

// Image is the model that contains all the data for submitted images.
type Image struct {
	ID          snowflake.Snowflake
	ContentType string
	Tags        []string
}

// ImageURL creates a valid URL from the current post metadata
func (img Image) ImageURL(baseURL string) string {
	return imageURL(img.ID, img.ContentType, baseURL)
}

// Extension gets the file's extension
func (img Image) Extension() string {
	return extension(img.ContentType)
}

func extension(contentType string) string {
	split := strings.Split(contentType, "/")
	if len(split) != 2 {
		return "png"
	}
	return split[1]
}

func imageURL(id snowflake.Snowflake, contentType string, baseURL string) string {
	return baseURL + id.ToBase64() + "." + extension(contentType)
}
