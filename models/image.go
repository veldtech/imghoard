package imghoard

import (
	sf "github.com/mikibot/imghoard/services/snowflake"
	"strings"
)

// Image is the model that contains all the data for submitted images.
type Image struct {
	ID          sf.Snowflake
	ContentType string
	Tags        []string
}

// ImageURL creates a valid URL from the current post metadata
func (img Image) ImageURL(baseURL string) string {
	return imageURL(sf.Snowflake(img.ID), img.ContentType, baseURL)
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

func imageURL(id sf.Snowflake, contentType string, baseURL string) string {
	return baseURL + id.ToBase64() + "." + extension(contentType)
}

