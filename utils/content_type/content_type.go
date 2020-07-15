package content_type

import (
	"errors"
	"strings"
)

type ContentType struct {
	Type string
	Extension string
}

func FromString(value string) (ContentType, error) {
	valueSplit := strings.Split(value, "/")
	if len(valueSplit) != 2 {
		return ContentType{}, errors.New("could not parse content type")
	}

	return ContentType{
		Type: valueSplit[0],
		Extension: valueSplit[1],
	}, nil
}

func (contentType ContentType) ToString() string {
	return contentType.Type + "/" + contentType.Extension
}