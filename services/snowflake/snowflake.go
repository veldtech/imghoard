package imghoard

import (
	"encoding/binary"
	"log"

	"github.com/btcsuite/btcutil/base58"
	"github.com/bwmarrin/snowflake"
)

type Snowflake snowflake.ID

func (sf Snowflake) ToBase64() string {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(snowflake.ID(sf).Int64()))
	return base58.Encode(bytes)
}

type SnowflakeService struct {
	node *snowflake.Node
}

// InitSnowflake generates the initial snowflakeGenerator
func InitSnowflake() *SnowflakeService {
	gen, err := snowflake.NewNode(1)
	if err != nil {
		log.Panicf("Could not create snowflake generator: %s", err)
	}
	return &SnowflakeService{
		node: gen,
	}
}

// GenerateID creates a new unique id
func (service *SnowflakeService) GenerateID() Snowflake {
	return Snowflake(service.node.Generate())
}
