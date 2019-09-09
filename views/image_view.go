package imghoard

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	image "github.com/mikibot/imghoard/services/imagehandler"
	uuid "github.com/mikibot/imghoard/services/snowflake"

	jsoniter "github.com/json-iterator/go"
	models "github.com/mikibot/imghoard/models"
	spaces "github.com/mikibot/imghoard/services/spaces"
	"github.com/savsgio/atreugo/v7"
)

var json = jsoniter.ConfigFastest

// ImageResult is the user facing model for images.
type ImageResult struct {
	ID   uuid.Snowflake `json:"id"`
	Tags []string       `json:"tags"`
	URL  string         `json:"url"`
}

// ImageView is the dataset for the image controller
type ImageView struct {
	BaseURL string
	Handler image.ImageHandler
}

// GetImage gets a random image with optional tags
// GET /images?page=1[tags={...}]
func (view ImageView) GetImage(ctx *atreugo.RequestCtx) error {
	page := 0
	args := ctx.QueryArgs()
	if args.Has("page") {
		p, err := strconv.ParseInt(string(args.Peek("page")), 0, 16)
		if err != nil {
			page = 0
			log.Print(err)
		} else {
			page = int(p)
		}
	}

	var images []models.Image
	if args.Has("tags") {
		tags := strings.Split(string(args.Peek("tags")), " ")
		i, err := view.Handler.FindImages(tags, 100, page*100)
		if err != nil {
			return ctx.JSONResponse(errors.New(err.Error()), 500)
		}
		images = i
	} else {
		i, err := view.Handler.GetImages(100, page*100)
		if err != nil {
			return ctx.JSONResponse(errors.New(err.Error()), 500)
		}
		images = i
	}

	if images == nil ||
		len(images) == 0 {
		return ctx.JSONResponse(errors.New("not found"), 404)
	}
	return ctx.JSONResponse(view.toImageResult(images))
}

// GetImageByID gets a specific image by ID
// GET /images/:id
func (view ImageView) GetImageByID(ctx *atreugo.RequestCtx) error {
	idStr, ok := ctx.UserValue("id").(string)
	if !ok {
		return models.InternalServerError(ctx)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return models.BadRequest(ctx, "Invalid ID provided")
	}
	image, err := view.Handler.GetImage(uuid.Snowflake(id))
	if err != nil {
		return models.InternalServerError(ctx)
	}

	if image.ID == 0 {
		return models.NotFound(ctx)
	}

	return ctx.JSONResponse(ImageResult{
		ID:   image.ID,
		URL:  image.ImageURL(view.BaseURL),
		Tags: image.Tags,
	})
}

// PostImage allows you to upload an image and set the tags
// POST /images
// - Requires authentication
func (view ImageView) PostImage(ctx *atreugo.RequestCtx) error {
	var newPost = spaces.ImageSubmission{}
	json.Unmarshal(ctx.PostBody(), &newPost)
	if len(newPost.Data) == 0 {
		return errors.New("invalid content: image.data is empty")
	}

	image, err := view.Handler.AddImage(newPost)
	if err != nil {
		return ctx.JSONResponse(atreugo.JSON{
			"error": err.Error(),
		}, 500)
	}

	return ctx.JSONResponse(atreugo.JSON{
		"file": fmt.Sprintf("%s%s.%s",
			view.BaseURL,
			image.ID.ToBase64(),
			image.Extension()),
	})
}

// GetTag gets a tag by ID and shows its metadata
// GET /tags/{tagName}
func (view ImageView) GetTag(ctx *atreugo.RequestCtx) error {
	return nil
}

// PatchTag updates a tag's metadata
// PATCH /tags/{tagName}
func (view ImageView) PatchTag(ctx *atreugo.RequestCtx) error {
	return nil
}

func (view ImageView) toImageResult(images []models.Image) []ImageResult {
	result := make([]ImageResult, len(images))
	for i := 0; i < len(images); i++ {
		result[i] = ImageResult{
			ID:   images[i].ID,
			Tags: images[i].Tags,
			URL:  images[i].ImageURL(view.BaseURL),
		}
	}
	return result
}
