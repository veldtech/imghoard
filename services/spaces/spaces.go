package imghoard

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"

	models "github.com/mikibot/imghoard/models"
	uuid "github.com/mikibot/imghoard/services/snowflake"
	"github.com/minio/minio-go/v6"
)

// ApiClient contains data for S3 bucket calls
type ApiClient struct {
	config  Config
	s3      *minio.Client
	uuidGen *uuid.SnowflakeService
}

type Config struct {
	AccessKey string
	SecretKey string
	Endpoint  string
	Bucket    string
	Folder    string
}

// New creates and saves a DigitalOcean CDN API client.
// TODO: maybe not wrap the minioClient now that it is changed from aws-s3
func New(config Config, idGenerator *uuid.SnowflakeService) *ApiClient {
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

	_, err := c.s3.PutObject(
		c.config.Bucket,
		c.config.Folder+filePath,
		bytes.NewReader(image.Data),
		int64(len(image.Data)),
		minio.PutObjectOptions{
			UserMetadata: map[string]string{"x-amz-acl": "public-read"},
			ContentType:  image.ContentType,
		})
	if err != nil {
		log.Fatalln(err)
	}

	x := models.Image{
		ID:          id,
		ContentType: image.ContentType,
	}
	fmt.Print(x)
	return x, nil
}
