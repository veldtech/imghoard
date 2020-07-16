package snowflake

import (
	"encoding/binary"
	"log"

	"github.com/btcsuite/btcutil/base58"
	"github.com/bwmarrin/snowflake"
	"log"
	"net"
	"strconv"
	"strings"
)

type Snowflake snowflake.ID

type IdGenerator interface {
	Generate() Snowflake
}

// InitSnowflake generates the initial snowflakeGenerator
func New() (IdGenerator, error) {
	ip, err := fetchLocalIPAddressBytes()
	if err != nil {
		return nil, err
	}

	gen, err := snowflake.NewNode(int64(binary.BigEndian.Uint32(ip)) % 1023)
	if err != nil {
		return nil, err
	}

	return gen, nil
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

func (sf Snowflake) ToBase64() string {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(snowflake.ID(sf).Int64()))
	return base58.Encode(bytes)
}