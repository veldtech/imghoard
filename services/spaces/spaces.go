package imghoard

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	models "github.com/mikibot/imghoard/models"
	snowflake "github.com/mikibot/imghoard/services/snowflake"
	"github.com/minio/minio-go/v6"
)

// SpacesAPIClient contains data for S3 bucket calls
type SpacesAPIClient struct {
	folder string
	bucket string
	s3     *minio.Client
}

// New creates and saves a DigitalOcean CDN API client.
func New() *SpacesAPIClient {
	spacesKey, valid := os.LookupEnv("DO_ACCESS")
	if !valid {
		log.Fatalln("Could not create DigitalOcean spaces api instance, missing DO_ACCESS.")
	}

	spacesSecret, valid := os.LookupEnv("DO_SECRET")
	if !valid {
		log.Fatalln("Could not create DigitalOcean spaces api instance, missing DO_ACCESS.")
	}

	endpoint, valid := os.LookupEnv("DO_ENDPOINT")
	if !valid {
		log.Fatalln("Could not create DigitalOcean space API instance, missing DO_ENDPOINT.")
	}

	spacesBucket, valid := os.LookupEnv("DO_BUCKET")
	if !valid {
		log.Fatalln("Could not create DigitalOcean space API instance, missing DO_BUCKET.")
	}

	spacesFolder, valid := os.LookupEnv("DO_FOLDER")
	if !valid {
		spacesFolder = ""
	}

	minioClient, err := minio.New(endpoint, spacesKey, spacesSecret, false)
	if err != nil {
		log.Fatalln(err)
	}

	return &SpacesAPIClient{
		folder: spacesFolder,
		bucket: spacesBucket,
		s3:     minioClient,
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

	id := snowflake.GenerateID()

	extension := strings.Split(string(contentType), "/")
	if len(extension) != 2 {
		return models.Image{}, errors.New("invalid ContentType")
	}
	var filePath = id.Base32() + "." + string(extension[1])

	decoded, err := base64.StdEncoding.DecodeString(segments[1])
	if err != nil {
		return models.Image{}, err
	}

	_, err = c.s3.PutObject(
		c.bucket,
		c.folder+filePath,
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
		ID:          int64(id),
		ContentType: string(contentType),
	}, nil
}
