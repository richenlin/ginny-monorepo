package snowflake

import (
	"crypto/rand"
	"hash/fnv"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/snowflake"
)

var snowflakeNode *snowflake.Node

func init() {
	var err error
	snowflakeNode, err = snowflake.NewNode(getNodeNumber())
	if err != nil {
		panic("init snowflake error " + err.Error())
	}
}

// ID new simple snowflake id
func ID() uint64 {
	return uint64(snowflakeNode.Generate())
}

// getNodeNumber 获取节点ID，先尝试从hostname中获取，如果无法获取，则随机产生
func getNodeNumber() int64 {
	// 先尝试从hostname
	hostname, err := os.Hostname()
	if err != nil {
		return getHostHashNumber("")
	}
	index := strings.LastIndex(hostname, "-")
	if index <= 0 {
		return getHostHashNumber(hostname)
	}
	hostSuffix := hostname[index+1:]
	if hostSuffix == "" {
		return getHostHashNumber(hostname)
	}
	number, errInt := strconv.Atoi(hostSuffix)
	if errInt != nil {
		return getHostHashNumber(hostname)
	}
	if number > 1023 || number < 0 {
		return getHostHashNumber(hostname)
	}
	return int64(number)
}

func getHostHashNumber(hostname string) int64 {
	if hostname == "" {
		randNumber, err := rand.Int(rand.Reader, big.NewInt(1023))
		if err != nil {
			return 0
		}
		return randNumber.Int64()
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(hostname))
	return int64(h.Sum32() % 1023)
}
