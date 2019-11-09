package imghoard

import (
	"fmt"
	"strconv"
	"strings"

	image "github.com/mikibot/imghoard/services/imagehandler"
	uuid "github.com/mikibot/imghoard/services/snowflake"

	jsoniter "github.com/json-iterator/go"
	models "github.com/mikibot/imghoard/models"
	spaces "github.com/mikibot/imghoard/services/spaces"
	"github.com/savsgio/atreugo/v9"
)

var json = jsoniter.ConfigFastest

// ImageResult is the user facing model for images.
type ImageResult struct {
	Id   uuid.Snowflake `json:"id"`
	Tags []string       `json:"tags"`
	Url  string         `json:"url"`
}

// ImageView is the dataset for the image controller
type ImageView struct {
	BaseUrl string
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
		} else {
			page = int(p)
		}
	}

	var images []models.Image
	if args.Has("tags") {
		tags := strings.Split(string(args.Peek("tags")), " ")
		i, err := view.Handler.FindImages(tags, 100, page*100)
		if err != nil {
			return models.Error(ctx, 500, err)
		}
		images = i
	} else {
		i, err := view.Handler.GetImages(100, page*100)
		if err != nil {
			return models.Error(ctx, 500, err)
		}
		images = i
	}

	if images == nil ||
		len(images) == 0 {
		return models.ErrorStr(ctx, 404, "no images found")
	}
	return models.JSON(ctx, view.toImageResult(images))
}

// GetImageByID gets a specific image by ID
// GET /images/:id
func (view ImageView) GetImageByID(ctx *atreugo.RequestCtx) error {
	idStr, ok := ctx.UserValue("id").(string)
	if !ok {
		return models.ErrorStr(ctx, 400, "no 'id' parameter was provided")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return models.Error(ctx, 400, err)
	}
	image, err := view.Handler.GetImage(uuid.Snowflake(id))
	if err != nil {
		return models.Error(ctx, 500, err)
	}

	if image.ID == 0 {
		return models.ErrorStr(ctx, 404, "no image was found")
	}

	return models.JSON(ctx, ImageResult{
		Id:   image.ID,
		Url:  image.ImageURL(view.BaseUrl),
		Tags: image.Tags,
	})
}

// PostImage allows you to upload an image and set the tags
// POST /images
// - Requires authentication
func (view ImageView) PostImage(ctx *atreugo.RequestCtx) error {
	fmt.Println("ok")

	var newPost = spaces.ImageSubmission{}
	err := json.Unmarshal(ctx.PostBody(), &newPost)
	if err != nil {
		return models.Error(ctx, 400, err)
	}

	if len(newPost.Data) == 0 {
		return models.ErrorStr(ctx, 400, "image.data is empty")
	}

	image, err := view.Handler.AddImage(newPost)
	if err != nil {
		return models.Error(ctx, 500, err)
	}

	return models.JSON(ctx, atreugo.JSON{
		"file": fmt.Sprintf("%s%s.%s",
			view.BaseUrl,
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
			Id:   images[i].ID,
			Tags: images[i].Tags,
			Url:  images[i].ImageURL(view.BaseUrl),
		}
	}
	return result
}
