package uuid

import (
	"github.com/donetkit/contrib/utils/snowflake"
	"log"
)

var snowFlakeNode *snowflake.Node

func init() {
	var err error
	snowFlakeNode, err = snowflake.NewNode(0)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func GenerateSnowFlakeId() int64 {
	return snowFlakeNode.Generate().Int64()
}
