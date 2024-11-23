package server

import (
	"log"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"gopkg.in/ini.v1"
)

func LoadCorsConfig() cors.Config {
	f, _ := os.Getwd()
	log.Println(f)
	cfg, err := ini.Load(f + "/server/config.cfg")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	corsSection := cfg.Section("cors")
	allowOrigins := strings.Split(corsSection.Key("allow_origins").String(), ",")
	for i := range allowOrigins {
		allowOrigins[i] = strings.TrimSpace(allowOrigins[i])
	}

	allowMethods := strings.Split(corsSection.Key("allow_methods").String(), ",")
	for i := range allowMethods {
		allowMethods[i] = strings.TrimSpace(allowMethods[i])
	}

	allowHeaders := strings.Split(corsSection.Key("allow_headers").String(), ",")
	for i := range allowHeaders {
		allowHeaders[i] = strings.TrimSpace(allowHeaders[i])
	}

	exposeHeaders := strings.Split(corsSection.Key("expose_headers").String(), ",")
	for i := range exposeHeaders {
		exposeHeaders[i] = strings.TrimSpace(exposeHeaders[i])
	}

	allowCredentials, err := corsSection.Key("allow_credentials").Bool()
	if err != nil {
		log.Fatalf("Failed to parse allow_credentials: %v", err)
	}

	return cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		ExposeHeaders:    exposeHeaders,
		AllowCredentials: allowCredentials,
	}
}