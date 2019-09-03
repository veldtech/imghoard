package imghoard

import (
	"database/sql"
	"github.com/bwmarrin/snowflake"
	models "github.com/mikibot/imghoard/models"
	imghoard "github.com/mikibot/imghoard/services/spaces"
)

type ImageHandler interface {
	AddImage(submission imghoard.ImageSubmission)
	GetImage(id snowflake.ID) models.Image
	FindImages(tags []string, amount int, offset int) []models.Image
}

type InternalImageHandler struct {
	spacesClient *imghoard.SpacesAPIClient
	pgConn *sql.DB
}

func GetImageByTags(tags string) []models.Image {

}