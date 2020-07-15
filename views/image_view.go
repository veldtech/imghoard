package imghoard

import (
	"database/sql"
	"errors"
	"fmt"
	snowflake2 "github.com/mikibot/imghoard/services/snowflake"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/snowflake"
	jsoniter "github.com/json-iterator/go"
	models "github.com/mikibot/imghoard/models"
	spaces "github.com/mikibot/imghoard/services/spaces"
	"github.com/savsgio/atreugo/v7"
)

var json = jsoniter.ConfigFastest

// ImageView is the dataset for the image controller
type ImageView struct {
	BaseURL      string
	SpacesClient *spaces.SpacesAPIClient
	Db *sql.DB
	Generator snowflake2.IdGenerator
}

// GetImage gets a random image with optional tags
// GET /images?page=1[tags={...}]
func (i ImageView) GetImage(ctx *atreugo.RequestCtx) error {
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

	var images []models.ImageResult
	if args.Has("tags") {
		tags := strings.Split(string(args.Peek("tags")), " ")
		i, err := models.GetTags(i.Db, i.BaseURL, 100, page*100, tags)
		if err != nil {
			return ctx.JSONResponse(models.New(err.Error()), 500)
		}
		images = i
	} else {
		i, err := models.Get(i.Db, i.BaseURL, 100, page*100)
		if err != nil {
			return ctx.JSONResponse(models.New(err.Error()), 500)
		}
		images = i
	}

	if images == nil ||
		len(images) == 0 {
		return ctx.JSONResponse(models.New("not found"), 404)
	}

	return ctx.JSONResponse(images)
}

// GetImageByID gets a specific image by ID
// GET /images/:id
func (i ImageView) GetImageByID(ctx *atreugo.RequestCtx) error {
	return nil
}

// PostImage allows you to upload an image and set the tags
// POST /images
// - Requires authentication
func (i ImageView) PostImage(ctx *atreugo.RequestCtx) error {
	if strings.Contains(string(ctx.Request.Header.ContentType()), "multipart/form-data") {
		return i.uploadImageFromMultipart(ctx)
	}

	var newPost = spaces.ImageSubmission{}
	err := json.Unmarshal(ctx.PostBody(), &newPost)
	if err != nil {
		log.Println("error: " + err.Error())
		return errors.New("invalid content: could not read data")
	}

	if len(newPost.Data) == 0 {
		return errors.New("invalid content: image.data is empty")
	}

	image, err := i.SpacesClient.UploadData(newPost.Data)
	if err != nil {
		return ctx.JSONResponse(atreugo.JSON{
			"error": err.Error(),
		}, 500)
	}

	image.Tags = newPost.Tags
	err = image.Insert(i.Db, i.Generator)
	if err != nil {
		return ctx.JSONResponse(atreugo.JSON{
			"error": err.Error(),
		})
	}

	return ctx.JSONResponse(atreugo.JSON{
		"file": fmt.Sprintf("%s%s.%s",
			i.BaseURL,
			snowflake.ID(image.ID).Base32(),
			image.Extension()),
	})
}

func (i ImageView) uploadImageFromMultipart(ctx *atreugo.RequestCtx) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		return err
	}

	if form.File["file"] == nil {
		return errors.New("cannot find file attached to form")
	}
	stream, err := form.File["file"][0].Open()
	if err != nil {
		return err
	}

	imageBytes, err := ioutil.ReadAll(io.Reader(stream))
	if err != nil {
		return err
	}

	imageContentType := form.Value["filetype"]
	if len(imageContentType) == 0 || imageContentType[0] == "" {
		return errors.New("validation['filetype']: is empty")
	}

	tags := strings.Split(strings.ToLower(strings.Join(form.Value["tags"], ",")), ",")

	image, err := i.SpacesClient.UploadDataRaw(imageBytes, imageContentType[0])
	if err != nil {
		return err
	}

	image.Tags = tags
	err = image.Insert(i.Db, i.Generator)
	if err != nil {
		return err
	}

	return ctx.JSONResponse(atreugo.JSON{
		"file": fmt.Sprintf("%s%s.%s", i.BaseURL, snowflake.ID(image.ID).Base32(), image.Extension()),
	})
}

// GetTag gets a tag by ID and shows its metadata
// GET /tags/{tagName}
func (i ImageView) GetTag(ctx *atreugo.RequestCtx) error {
	return nil
}

// PatchTag updates a tag's metadata
// PATCH /tags/{tagName}
func (i ImageView) PatchTag(ctx *atreugo.RequestCtx) error {
	return nil
}
