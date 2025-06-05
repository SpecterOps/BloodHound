package config

import (
	"fmt"
	"os"
)

// Edit the following section to match your configuration
// To generate the API keys follow this guide : https://.....
var Server = ServerConfig{
	URL:    "http://localhost:8080",
	APIKey: "wTmWGMsdqWj8HeULYhRXia+0QosnImF2wo2rTeHqSKQRfPo/MdkDzA==", // TODO: Insert API key
	APIID:  "8229c6eb-630f-4dc7-86a2-7626a8acbcf3",                     // TODO: Insert API ID
}

var Generator = GeneratorConfig{
	User: GeneratorUserFiles{
		First: "resources/firstname.yml",
		Last:  "resources/lastname.yml",
	},
	Percentages: GeneratorPercentages{
		Server:      15,
		Workstation: 85,
	},
}

type ServerConfig struct {
	URL    string
	APIKey string
	APIID  string
}

type GeneratorUserFiles struct {
	First string
	Last  string
}

type GeneratorPercentages struct {
	Server      int
	Workstation int
}

type GeneratorConfig struct {
	User        GeneratorUserFiles
	Percentages GeneratorPercentages
}

// Validate ensures required config fields are set
func Validate() {
	if Server.APIKey == "" || Server.APIID == "" {
		fmt.Println("❌ API key or ID not set in config.go.")
		fmt.Println("ℹ️  Please generate API credentials and update the Server.APIKey and Server.APIID values in config.go.")
		os.Exit(1)
	}
}
