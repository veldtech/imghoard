package snowflake

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

func fetchLocalIPAddressBytes() ([]byte, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addresses {
		segments := strings.Split(addr.String(), "/")
		if len(segments) != 2 {
			continue
		}

		ipLen, err := strconv.Atoi(segments[1])
		if err != nil || ipLen > 24 {
			continue
		}

		ipBytes := strings.Split(segments[0], ".")
		byteArray := make([]byte, 4)
		for i, str := range ipBytes {
			ipSegment, err := strconv.Atoi(str)
			if err != nil {
				continue
			}
			byteArray[i] = byte(ipSegment)
		}

		return byteArray, nil
	}

	return nil, errors.New("could not find any IP address")
}
