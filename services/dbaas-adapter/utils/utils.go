// Copyright 2024-2025 NetCracker Technology Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/utils"
	"github.com/gofiber/fiber/v2"
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

type MongodbUpdateSettingsRequest struct {
	CurrentSettings map[string]interface{} `json:"currentSettings"`
	NewSettings     map[string]interface{} `json:"newSettings"`
}

type ShardingSettings struct {
	CollectionName string `json:"collectionName"`
	ShardKeys      []ShardKeyField
}

type ShardKeyField struct {
	Field  string `json:"field"`
	Hashed bool   `json:"hashed"`
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

func ValidateDbIdentifierParam(ctx context.Context, paramName string, paramValue string, pattern string) bool {
	logger := utils.GetLogger(GetEnvBool("LOG_DEBUG", false))
	if paramValue != "" {
		matched, err := regexp.MatchString(pattern, paramValue)
		if err != nil {
			logger.Error(fmt.Sprintf("Error during check %s", paramName), zap.Error(err))
			return false
		}

		if !matched {
			logger.Info(fmt.Sprintf("Provided %s does not meet the requirements", paramName))
		}

		return matched
	}
	return true
}

func SendInvalidParameterResponse(c *fiber.Ctx, paramName string, paramValue string, pattern string) error {
	return c.Status(400).SendString(fmt.Sprintf("Invalid '%s' param provided: %s. '%s' param must comply to the pattern %s", paramName, paramValue, paramName, pattern))
}
