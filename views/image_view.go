package imghoard

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/palantir/stacktrace"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	imagehandler "github.com/mikibot/imghoard/services/imagehandler"
	uuid "github.com/mikibot/imghoard/services/snowflake"

	jsoniter "github.com/json-iterator/go"
	models "github.com/mikibot/imghoard/models"
	spaces "github.com/mikibot/imghoard/services/spaces"
	"github.com/savsgio/atreugo/v11"
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
	Handler imagehandler.ImageHandler
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
			return models.Error(ctx, 500, stacktrace.Propagate(err, "could not find images"))
		}
		images = i
	} else {
		i, err := view.Handler.GetImages(100, page*100)
		if err != nil {
			return models.Error(ctx, 500, stacktrace.Propagate(err, "could not get images"))
		}
		images = i
	}

	if images == nil || len(images) == 0 {
		return models.ErrorStr(ctx, 404, "no images found")
	}
	return models.JSON(ctx, view.toImageResult(images))
}

func (view ImageView) GetRandomImage(ctx *atreugo.RequestCtx) error {
	args := ctx.QueryArgs()

	var tags []string = nil;
	if args.Has("tags") {
		tags = strings.Split(string(args.Peek("tags")), " ")
	}

	image, err := view.Handler.GetRandomImage(tags)
	if err != nil {
		return stacktrace.Propagate(err, "no random image returned")
	}

	return models.JSON(ctx, view.toImageResponse(image))
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

	return models.JSON(ctx, view.toImageResponse(image))
}

// PostImage allows you to upload an image and set the tags
// POST /images
// - Requires authentication
func (view ImageView) PostImage(ctx *atreugo.RequestCtx) error {
	contentType := string(ctx.Request.Header.ContentType())

	var submission spaces.ImageSubmission

	if strings.HasPrefix(contentType, "multipart/form-data") {
		multipartSubmission, err := view.uploadImageFromMultipart(ctx)
		if err != nil {
			return models.Error(ctx, 400, stacktrace.Propagate(err, ""))
		}
		submission = multipartSubmission
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

func (i ImageView) uploadImageFromMultipart(ctx *atreugo.RequestCtx) (spaces.ImageSubmission, error) {
	form, err := ctx.MultipartForm()
	if err != nil {
		return spaces.ImageSubmission{}, err
	}

	if form.File["file"] == nil {
		return spaces.ImageSubmission{}, errors.New("cannot find file attached to form")
	}
	stream, err := form.File["file"][0].Open()
	if err != nil {
		return spaces.ImageSubmission{}, err
	}

	imageBytes, err := ioutil.ReadAll(io.Reader(stream))
	if err != nil {
		return spaces.ImageSubmission{}, err
	}

	imageContentType := form.Value["filetype"]
	if len(imageContentType) == 0 || imageContentType[0] == "" {
		return spaces.ImageSubmission{}, errors.New("validation['filetype']: is empty")
	}

	tags := strings.Split(strings.ToLower(strings.Join(form.Value["tags"], ",")), ",")

	return spaces.ImageSubmission{
		Data: imageBytes,
		Tags: tags,
		ContentType: imageContentType[0],
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
		result[i] = view.toImageResponse(images[i])
	}
	return result
}

func (view ImageView) toImageResponse(image models.Image) ImageResult {
	return ImageResult{
		Id:   image.ID,
		Tags: image.Tags,
		Url:  image.ImageURL(view.BaseUrl),
	}
}
