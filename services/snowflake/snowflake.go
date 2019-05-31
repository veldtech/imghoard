package imghoard

import (
	"log"
	"github.com/bwmarrin/snowflake"
)

var snowflakeGenerator *snowflake.Node

// InitSnowflake generates the initial snowflakeGenerator
func InitSnowflake() {
	gen, err := snowflake.NewNode(1)
	if(err != nil) {
		log.Panicf("Could not create snowflake generator: %s", err)
	}
	snowflakeGenerator = gen
}

// GenerateID creates a new unique id
func GenerateID() snowflake.ID {
	return snowflakeGenerator.Generate()
}