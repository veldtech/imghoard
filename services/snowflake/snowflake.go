package snowflake

import (
	"encoding/binary"
	"github.com/btcsuite/btcutil/base58"
	"github.com/bwmarrin/snowflake"
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

	return Generator{
		Node: gen,
	}, nil
}

type Generator struct {
	IdGenerator
	*snowflake.Node
}

func (gen Generator) Generate() Snowflake {
	return Snowflake(gen.Node.Generate())
}

func (sf Snowflake) ToBase64() string {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(snowflake.ID(sf).Int64()))
	return base58.Encode(bytes)
}