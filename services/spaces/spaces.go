package imghoard

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mikibot/imghoard/config"
	"github.com/mikibot/imghoard/utils/content_type"
	"log"
	"strings"

	models "github.com/mikibot/imghoard/models"
	uuid "github.com/mikibot/imghoard/services/snowflake"
	"github.com/minio/minio-go/v6"
)

// ApiClient contains data for S3 bucket calls
type ApiClient struct {
	config  config.Config
	s3      *minio.Client
	uuidGen uuid.IdGenerator
}

// New creates and saves a DigitalOcean CDN API client.
// TODO: maybe not wrap the minioClient now that it is changed from aws-s3
func New(config config.Config, idGenerator *uuid.SnowflakeService) *ApiClient {
	minioClient, err := minio.New(config.Endpoint, config.AccessKey, config.SecretKey, false)
	if err != nil {
		log.Fatalln(err)
	}


	return &ApiClient{
		config:  config,
		s3:      minioClient,
		uuidGen: idGenerator,
	}
}

// ImageSubmission is in need of documentation
type ImageSubmission struct {
	Data []byte
	Tags []string
	ContentType string
}

// UploadData uploads your data to the preferred do client
func (c *ApiClient) UploadData(image ImageSubmission) (models.Image, error) {
	id := c.uuidGen.GenerateID()

	contentType := strings.Split(image.ContentType, "/")
	if len(contentType) < 2 {
		return models.Image{}, errors.New("invalid content type")
	}

	filePath := fmt.Sprintf("%s.%s", id.ToBase64(), contentType[1])

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
		c.config.S3Bucket,
		c.config.S3Folder + filePath,
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

