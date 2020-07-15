package imghoard

import (
	"database/sql"
	"strings"

	"github.com/bwmarrin/snowflake"
	"github.com/lib/pq"

	sf "github.com/mikibot/imghoard/services/snowflake"
)

// Image is the model that contains all the data for submitted images.
type Image struct {
	ID          int64
	ContentType string
	Tags        []string
}

// ImageResult is the result for images
type ImageResult struct {
	ID   int64
	Tags []string
	URL  string
}

// Extension gets the file's extension
func (img Image) Extension() string {
	return extension(img.ContentType)
}

// Get n amount of images from the
func Get(db *sql.DB, baseURL string, amount int, offset int) ([]ImageResult, error) {
	rows, err := db.Query(`select f.id, f.contentType, array(select t.name from image_tags 
		it join tag t on it.tag_id = t.id where it.image_id = f.id) as tags	 
		from image f offset $1 limit $2;`, offset, amount)
	if err != nil {
		return nil, err
	}

	var result []ImageResult
	for rows.Next() {
		var id int64
		var contentType string
		var tags []string

		rows.Scan(&id, &contentType, pq.Array(&tags))

		result = append(result, ImageResult{
			ID:   id,
			URL:  imageURL(snowflake.ID(id), contentType, baseURL),
			Tags: tags,
		})
	}
	return result, nil
}

// GetImageByID returns a specific image selected from the ID
func GetImageByID(id int64) Image {
	return Image{
		ID: 0,
	}
}

// GetTags gets with
func GetTags(db *sql.DB, baseURL string, amount int, offset int, tags []string) ([]ImageResult, error) {
	rows, err := db.Query(`select f.id, f.contentType, array(select t.name from image_tags
		it join tag t on it.tag_id = t.id where it.image_id = f.id) as tags from image
		f where array(select t.name from image_tags it join tag t on it.tag_id = t.id 
		where it.image_id = f.id) @> $1 offset $2 limit $3;`,
		pq.Array(tags), offset, amount)
	if err != nil {
		return nil, err
	}

	var result []ImageResult
	for rows.Next() {
		var id int64
		var contentType string
		var tags []string

		rows.Scan(&id, &contentType, pq.Array(&tags))

		result = append(result, ImageResult{
			ID:   id,
			URL:  imageURL(snowflake.ID(id), contentType, baseURL),
			Tags: tags,
		})
	}
	return result, nil
}

// ImageURL creates a valid URL from the current post metadata
func (img Image) ImageURL(baseURL string) string {
	return imageURL(snowflake.ID(img.ID), img.ContentType, baseURL)
}

// Insert inserts the metadata of the image to the database.
func (img Image) Insert(db *sql.DB, generator sf.IdGenerator) error {
	_, err := db.Query(
		"INSERT INTO image (id, contentType) VALUES ($1, $2);", img.ID, img.ContentType)
	if err != nil {
		return err
	}
	for _, tag := range img.Tags {
		id := generator.Generate()
		_, err := db.Query("INSERT INTO tag (id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING;",
			id,
			tag)
		if err != nil {
			return err
		}

		_, err = db.Query("INSERT INTO image_tags(tag_id, image_id) select id, $1 from tag where name = $2;",
			img.ID,
			tag)
		if err != nil {
			return err
		}
	}
	return nil
}

func extension(contentType string) string {
	split := strings.Split(contentType, "/")
	if len(split) != 2 {
		return "png"
	}
	return split[1]
}

func imageURL(id snowflake.ID, contentType string, baseURL string) string {
	return baseURL + id.Base32() + "." + extension(contentType)
}
 