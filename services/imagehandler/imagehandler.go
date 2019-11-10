package imghoard

import (
	"database/sql"
	"github.com/lib/pq"
	models "github.com/mikibot/imghoard/models"
	"github.com/mikibot/imghoard/services/postgres"
	uuid "github.com/mikibot/imghoard/services/snowflake"
	spaces "github.com/mikibot/imghoard/services/spaces"
	"github.com/palantir/stacktrace"
)

type ReadImageHandler interface {
	GetBaseURL() string
	FindImages(tags []string, amount int, offset int) ([]models.Image, error)
	GetImage(id uuid.Snowflake) (models.Image, error)
	GetImages(amount int, offset int) ([]models.Image, error)
}

type WriteImageHandler interface {
	AddImage(submission spaces.ImageSubmission) (models.Image, error)
}

type ImageHandler interface {
	ReadImageHandler
	WriteImageHandler
}

func New(baseURL string, spaces *spaces.ApiClient, database *postgres.Client) ImageHandler {
	var s ImageHandler
	s = &service{
		spacesClient: spaces,
		baseURL:      baseURL,
		database:     database,
	}
	return s
}

func NewMock(baseURL string, spaces *spaces.ApiClient, database *postgres.Client) ImageHandler {
	var s ImageHandler
	s = &mockService{
		ImageHandler: New(baseURL, spaces, database),
	}
	return s
}

type service struct {
	spacesClient *spaces.ApiClient
	database     *postgres.Client
	baseURL      string
}

type mockService struct {
	ImageHandler
}

func (handler *mockService) AddImage(submission spaces.ImageSubmission) (models.Image, error) {
	return models.Image{}, nil
}

func (handler *service) AddImage(submission spaces.ImageSubmission) (models.Image, error) {
	image, err := handler.spacesClient.UploadData(submission)
	if err != nil {
		return models.Image{}, stacktrace.Propagate(err, "")
	}

	tx, err := handler.database.Begin()
	if err != nil {
		return models.Image{}, stacktrace.Propagate(err, "")
	}

	_, err = tx.Exec(
		"INSERT INTO image (id, contentType) VALUES ($1, $2);", image.ID, image.ContentType)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return models.Image{}, stacktrace.Propagate(err, "")
		}
		return models.Image{}, stacktrace.Propagate(err, "")
	}

	for _, tag := range submission.Tags {
		_, err := tx.Exec(
			"INSERT INTO tag (id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING;",
			image.ID,
			tag)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return models.Image{}, stacktrace.Propagate(err, "")
			}
			return models.Image{}, stacktrace.Propagate(err, "")
		}

		_, err = tx.Exec(
			"INSERT INTO image_tags(tag_id, image_id) select id, $1 from tag where name = $2;",
			image.ID,
			tag)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return models.Image{}, stacktrace.Propagate(err, "")
			}
			return models.Image{}, stacktrace.Propagate(err, "")
		}
	}

	if err := tx.Commit(); err != nil {
		return models.Image{}, stacktrace.Propagate(err, "")
	}
	return image, nil
}

func (handler *service) GetBaseURL() string {
	return handler.baseURL
}

func (handler *service) GetImage(id uuid.Snowflake) (models.Image, error) {
	rows, err := handler.database.Query(`select f.id, f.contentType, array(select t.name from image_tags 
		it join tag t on it.tag_id = t.id where it.image_id = f.id) as tags	 
		from image f WHERE id = $1`, id)
	if err != nil {
		return models.Image{}, stacktrace.Propagate(err, "")
	}

	images, err := fetchImages(rows)
	if err != nil {
		return models.Image{}, stacktrace.Propagate(err, "")
	}

	if len(images) == 0 {
		return models.Image{}, nil
	}
	return images[0], nil
}

func (handler *service) GetImages(amount int, offset int) ([]models.Image, error) {
	rows, err := handler.database.Query(`select f.id, f.contentType, array(select t.name from image_tags 
		it join tag t on it.tag_id = t.id where it.image_id = f.id) as tags	 
		from image f offset $1 limit $2;`, offset, amount)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}

	images, err := fetchImages(rows)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	return images, nil}

func (handler *service) FindImages(tags []string, amount int, offset int) ([]models.Image, error) {
	rows, err := handler.database.Query(
		`select f.id, f.contentType, array(select t.name from image_tags it join tag t on it.tag_id = t.id where 
         it.image_id = f.id) as tags from image f where array(select t.name from image_tags it join tag t on 
    	 it.tag_id = t.id where it.image_id = f.id) @> $1 offset $2 limit $3;`,
		pq.Array(tags), offset, amount)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	images, err := fetchImages(rows)
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	return images, nil
}

func fetchImages(rows *sql.Rows) ([]models.Image, error) {
	var result []models.Image
	for rows.Next() {
		var id int64
		var contentType string
		var tags []string

		err := rows.Scan(&id, &contentType, pq.Array(&tags))
		if err != nil {
			return nil, stacktrace.Propagate(err, "")
		}

		result = append(result, models.Image{
			ID:          uuid.Snowflake(id),
			ContentType: contentType,
			Tags:        tags,
		})
	}
	return result, nil
}
