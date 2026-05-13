package utils

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

var (
	log            *zap.Logger
	isDebugEnabled = *flag.Bool("log_debug", GetEnvBool("LOG_DEBUG", false), "If debug logs is enabled, env: LOG_DEBUG")
)

type Settings struct {
	ShardingSettings []ShardingSettings     `json:"shardingSettings,omitempty"`
	TargetShard      string                 `json:"targetShard"`
	Extra            map[string]interface{} `json:"extra,omitempty"`
}

type ShardingSettings struct {
	CollectionName string `json:"collectionName"`
	ShardKey       string `json:"shardKey"`
	Strategy       string `json:"strategy"` // "Hashed" or "Ranged"
}

func IsTLSEnabled() bool {
	if tlsEnabled, err := strconv.ParseBool(os.Getenv("TLS_ENABLED")); err == nil && tlsEnabled {
		return true
	}
	return false
}

func GetCACert() string {
	cert, _ := os.LookupEnv("TLS_ROOTCERT")
	return cert
}

func GetEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		bvalue, err := strconv.ParseBool(value)
		if err != nil {
			log.Error(fmt.Sprintf("Can't parse %s boolean variable", key), zap.Error(err))
			panic(err)
		}
		return bvalue
	}
	return fallback
}

func GetSecret(path string, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}

	val := strings.TrimSpace(string(data))
	if val == "" {
		return fallback
	}

	return val
}
