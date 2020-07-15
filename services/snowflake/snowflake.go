package snowflake

import (
	"encoding/binary"
	"errors"
	"github.com/bwmarrin/snowflake"
	"log"
	"net"
	"strconv"
	"strings"
)

// InitSnowflake generates the initial snowflakeGenerator
func InitSnowflake() IdGenerator {
	ip, err := fetchLocalIPAddressBytes()

	gen, err := snowflake.NewNode(int64(binary.BigEndian.Uint32(ip)) % 1023)
	if err != nil {
		log.Panic(err)
	}

	return gen
}

type IdGenerator interface {
	Generate() snowflake.ID
}

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