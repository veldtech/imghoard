package imghoard

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/palantir/stacktrace"
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

type ImageSubmissionJSON struct {
	Data string
	Tags []string
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
			return models.Error(ctx, 500, stacktrace.Propagate(err, ""))
		}
		images = i
	} else {
		i, err := view.Handler.GetImages(100, page*100)
		if err != nil {
			return models.Error(ctx, 500, stacktrace.Propagate(err, ""))
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
		return models.Error(ctx, 400, stacktrace.Propagate(err, ""))
	}
	image, err := view.Handler.GetImage(uuid.Snowflake(id))
	if err != nil {
		return models.Error(ctx, 500, stacktrace.Propagate(err, ""))
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
	contentType := string(ctx.Request.Header.ContentType())

	var submission spaces.ImageSubmission
	if strings.HasPrefix(contentType, "multipart/form-data") {
		multiPartSubmission, err := parseMultipartImage(ctx)
		if err != nil {
			return models.Error(ctx, 400, stacktrace.Propagate(err, ""))
		}
		submission = multiPartSubmission
	} else {
		var jsonSubmission ImageSubmissionJSON
		err := json.Unmarshal(ctx.PostBody(), &jsonSubmission)
		if err != nil {
			return models.Error(ctx, 400, stacktrace.Propagate(err, ""))
		}

		if len(jsonSubmission.Data) == 0 {
			return models.ErrorStr(ctx, 400, "image.data is empty")
		}
		submission, err = parseHTTPImage(jsonSubmission.Data)
		if err != nil {
			return models.Error(ctx, 400, stacktrace.Propagate(err, ""))
		}
		submission.Tags = jsonSubmission.Tags
	}
	image, err := view.Handler.AddImage(submission)
	if err != nil {
		return models.Error(ctx, 500, stacktrace.Propagate(err, ""))
	}

	return models.JSON(ctx, atreugo.JSON{
		"file": fmt.Sprintf("%s%s.%s",
			view.BaseUrl,
			image.ID.ToBase64(),
			image.Extension()),	
	})
}

func parseMultipartImage(ctx *atreugo.RequestCtx) (spaces.ImageSubmission, error) {
	form, err := ctx.FormFile("data")
	if err != nil {
		return spaces.ImageSubmission{}, models.Error(ctx, 400, stacktrace.Propagate(err, "data"))
	}

	contentLength := form.Size
	if contentLength == 0 {
		return spaces.ImageSubmission{}, stacktrace.Propagate(errors.New("invalid content-length"), "")
	}

	open, err := form.Open()
	if err != nil {
		return spaces.ImageSubmission{}, stacktrace.Propagate(err, "")
	}

	buffer := make([]byte, contentLength)
	open.Read(buffer)

	if err != nil {
		return spaces.ImageSubmission{}, stacktrace.Propagate(err, "")
	}

	tags := strings.Split(string(ctx.FormValue("tags")), ",")
	dataType := string(ctx.FormValue("data-type"))

	return spaces.ImageSubmission{
		ContentType: dataType,
		Data: buffer,
		Tags: tags,
	}, nil
}

func parseHTTPImage(image string) (spaces.ImageSubmission, error) {
	segments := strings.Split(image, ",")
	if len(segments) != 2 {
		return spaces.ImageSubmission{}, errors.New("invalid image payload")
	}

	var metadata = bufio.NewReader(strings.NewReader(segments[0]))

	header, err := metadata.ReadBytes(':')
	if err != nil {
		return spaces.ImageSubmission{}, err
	}
	header = bytes.Trim(header, ":")

	if string(header) != "data" {
		return spaces.ImageSubmission{}, errors.New("header mismatch: Header does not start with 'data'")
	}

	contentType, err := metadata.ReadBytes(';')
	if err != nil {
		return spaces.ImageSubmission{}, err
	}
	contentType = bytes.TrimRight(contentType, ";")

	encoding, _, err := metadata.ReadLine()
	if err != nil {
		return spaces.ImageSubmission{}, err
	}

	if string(encoding) != "base64" {
		return spaces.ImageSubmission{}, fmt.Errorf("encoding format '%s' not supported", string(encoding))
	}

	extension := strings.Split(string(contentType), "/")
	if len(extension) != 2 {
		return spaces.ImageSubmission{}, errors.New("invalid ContentType")
	}

	decoded, err := base64.StdEncoding.DecodeString(segments[1])
	if err != nil {
		return spaces.ImageSubmission{}, err
	}

	return spaces.ImageSubmission{
		Data: decoded,
		ContentType: string(contentType),
		Tags: nil,
	}, nil
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
