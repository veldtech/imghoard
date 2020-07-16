package imghoard

import (
	"bytes"
	"github.com/mikibot/imghoard/config"
	models "github.com/mikibot/imghoard/models"
	uuid "github.com/mikibot/imghoard/services/snowflake"
	"github.com/mikibot/imghoard/utils/content_type"
	"github.com/minio/minio-go/v6"
	"github.com/palantir/stacktrace"
	"log"
)

// ApiClient contains data for S3 bucket calls
type ApiClient struct {
	config  config.Config
	s3      *minio.Client
	uuidGen uuid.IdGenerator
}

// New creates and saves a DigitalOcean CDN API client.
// TODO: maybe not wrap the minioClient now that it is changed from aws-s3
func New(config config.Config, idGenerator uuid.IdGenerator) *ApiClient {
	minioClient, err := minio.New(config.S3Endpoint, config.S3AccessKey, config.S3SecretKey, false)
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
	id := c.uuidGen.Generate()

	contentType, err := content_type.FromString(image.ContentType)
	if err != nil {
		return models.Image{}, stacktrace.Propagate(err, "")
	}

	filePath := id.ToBase64() + "." + contentType.Extension

	_, err = c.s3.PutObject(
		c.config.S3Bucket,
		c.config.S3Folder + filePath,
		bytes.NewReader(image.Data),
		int64(len(image.Data)),
		minio.PutObjectOptions{
			UserMetadata: map[string]string{"x-amz-acl": "public-read"},
			ContentType:  image.ContentType,
		})
	if err != nil {
		log.Fatalln(err)
	}

	return  models.Image{
		ID:          id,
		ContentType: image.ContentType,
	}, nil
}

func (c *ApiClient) UploadDataRaw(image []byte, filename string) (models.Image, error) {
	contentType, err := content_type.FromString(filename)
	if err != nil {
		return models.Image{}, err
	}

	id := c.uuidGen.Generate()
	filePath := id.ToBase64() + "." + contentType.Extension

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
		ID:          id,
		ContentType: contentType.ToString(),
	}, nil
}

