package imghoard

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/mikibot/imghoard/config"
	"github.com/mikibot/imghoard/utils/content_type"
	"log"
	"strings"

	models "github.com/mikibot/imghoard/models"
	"github.com/mikibot/imghoard/services/snowflake"
	"github.com/minio/minio-go/v6"
)

// SpacesAPIClient contains data for S3 bucket calls
type SpacesAPIClient struct {
	folder string
	bucket string
	s3 *minio.Client
	generator snowflake.IdGenerator
}

// New creates and saves a DigitalOcean CDN API client.
// TODO: maybe not wrap the minioClient now that it is changed from aws-s3
func New(config config.Config, generator snowflake.IdGenerator) *SpacesAPIClient {
	minioClient, err := minio.New(config.S3Endpoint, config.S3AccessKey, config.S3SecretKey, false)
	if err != nil {
		log.Fatalln(err)
	}

	return &SpacesAPIClient{
		folder: config.S3Folder,
		bucket: config.S3Bucket,
		s3:     minioClient,
		generator: generator,
	}
}

// ImageSubmission is in need of documentation
type ImageSubmission struct {
	Data string
	Tags []string
}

// UploadData uploads your data to the preferred do client
func (c *SpacesAPIClient) UploadData(image string) (models.Image, error) {
	segments := strings.Split(image, ",")
	if len(segments) != 2 {
		return models.Image{}, errors.New("Invalid image payload")
	}

	var metadata = bufio.NewReader(strings.NewReader(segments[0]))

	header, err := metadata.ReadBytes(':')
	if err != nil {
		return models.Image{}, err
	}
	header = bytes.Trim(header, ":")

	if string(header) != "data" {
		return models.Image{}, errors.New("header mismatch: Header does not start with 'data'")
	}

	contentType, err := metadata.ReadBytes(';')
	if err != nil {
		return models.Image{}, err
	}
	contentType = bytes.TrimRight(contentType, ";")

	encoding, _, err := metadata.ReadLine()
	if err != nil {
		return models.Image{}, err
	}

	if string(encoding) != "base64" {
		return models.Image{}, fmt.Errorf("encoding format '%s' not supported", string(encoding))
	}

	id := c.generator.Generate()

	content, err := content_type.FromString(string(contentType))
	if err != nil {
		return models.Image{}, err
	}
	var filePath = id.Base32() + "." + content.Extension

	decoded, err := base64.StdEncoding.DecodeString(segments[1])
	if err != nil {
		return models.Image{}, err
	}

	_, err = c.s3.PutObject(
		c.bucket,
		c.folder + filePath,
		bytes.NewReader(decoded),
		int64(len(decoded)),
		minio.PutObjectOptions{
			UserMetadata: map[string]string{"x-amz-acl": "public-read"},
			ContentType:  content.ToString(),
		})
	if err != nil {
		log.Fatalln(err)
	}

	return models.Image{
		ID:          int64(id),
		ContentType: content.ToString(),
	}, nil
}

func (c *SpacesAPIClient) UploadDataRaw(image []byte, filename string) (models.Image, error) {
	contentType, err := content_type.FromString(filename)
	if err != nil {
		return models.Image{}, err
	}

	id := c.generator.Generate()
	filePath := id.Base32() + "." + contentType.Extension

	_, err = c.s3.PutObject(
		c.bucket,
		c.folder + filePath,
		bytes.NewReader(image),
		int64(len(image)),
		minio.PutObjectOptions{
			UserMetadata: map[string]string{"x-amz-acl": "public-read"},
			ContentType:  contentType.ToString(),
		})
	if err != nil {
		log.Fatalln(err)
	}

	return models.Image{
		ID:          int64(id),
		ContentType: contentType.ToString(),
	}, nil
}

