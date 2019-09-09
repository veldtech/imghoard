package imghoard

import (
	"bufio"
	"bytes"
	"encoding/base64"
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
	Data string
	Tags []string
}

// UploadData uploads your data to the preferred do client
func (c *ApiClient) UploadData(image string) (models.Image, error) {
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

	id := c.uuidGen.GenerateID()

	extension := strings.Split(string(contentType), "/")
	if len(extension) != 2 {
		return models.Image{}, errors.New("invalid ContentType")
	}
	var filePath = id.ToBase64() + "." + string(extension[1])

	decoded, err := base64.StdEncoding.DecodeString(segments[1])
	if err != nil {
		return models.Image{}, err
	}

	_, err = c.s3.PutObject(
		c.config.Bucket,
		c.config.Folder+filePath,
		bytes.NewReader(decoded),
		int64(len(decoded)),
		minio.PutObjectOptions{
			UserMetadata: map[string]string{"x-amz-acl": "public-read"},
			ContentType:  string(contentType),
		})
	if err != nil {
		log.Fatalln(err)
	}

	return models.Image{
		ID:          id,
		ContentType: string(contentType),
	}, nil
}
