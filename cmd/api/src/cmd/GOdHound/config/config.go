package config

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

// Edit the following section to match your configuration
// To generate the API keys follow this guide : https://.....

// BHCE-Edge
var Server = ServerConfig{
	URL:    "http://localhost:8585",
	APIKey: "9+R0guXYZbOGXgWe65rqvk16Gq/dFCgP5kmp8kT1tT2G7xkm8ibp1A==",
	APIID:  "d236ef95-a872-4461-94fe-c841b7574df9",
}

// Main Client
/*
var Server = ServerConfig{
	URL:    "http://localhost:8080",
	APIKey: "wTmWGMsdqWj8HeULYhRXia+0QosnImF2wo2rTeHqSKQRfPo/MdkDzA==", // TODO: Insert API key
	APIID:  "8229c6eb-630f-4dc7-86a2-7626a8acbcf3",                     // TODO: Insert API ID
}
*/

// Edit the following to create the domain as you like

var Domain = DomainCongig{
	Name:    "TESTLAB.LOCAL",
	SID:     "",
	NbUsers: 500,
	NbComps: 500,
}

var Type = "AD" // Currently only AD is supported, but Entra is coming

var Directory = DirectoryConfig{
	BaseDir: "./output/",
	RunDir:  "", // this could be a prefix for your file. GOdHound will also prepend data & time of the run to the folder
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

type DomainCongig struct {
	Name    string
	SID     string
	NbUsers int
	NbComps int
}

type DirectoryConfig struct {
	BaseDir string
	RunDir  string
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
	if Type != "AD" {
		fmt.Println("❌ Currently only AD is supported, please set 'Type' to 'AD' in config.go.")
		fmt.Println("ℹ️  If you want to support new data creation, you can modify this check in config.go")
		os.Exit(1)
	}
}

func ConfDomain() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	subAuth1 := r.Uint32()
	subAuth2 := r.Uint32()
	subAuth3 := r.Uint32()

	Domain.SID = fmt.Sprintf("S-1-5-21-%d-%d-%d", subAuth1, subAuth2, subAuth3)
}
