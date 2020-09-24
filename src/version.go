package main

import (
	"fmt"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Long:  "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		showVersion(buildVersion, buildDate, buildHash, buildOS)
	},
}

func getCurrentVersion() (version string, date string) {
	path := viper.GetString("gcs_binary_path")
	binaryGCSBucket := viper.GetString("gcs_binary_bucket")

	if path == "" || binaryGCSBucket == "" {
		if verbose > 0 {
			log.Info("gcs_binary_bucket or gcs_binary_path is not defined, cannot retrieve binary")
		}
		return
	}

	if runtime.GOOS == "linux" {
		path = path + "linux/cog.version"
	} else {
		path = path + "darwin/cog.version"
	}

	versionDataBytes, err := getFileFromGCS(binaryGCSBucket, path)
	versionData := strings.Split(string(versionDataBytes), "\n")
	if err != nil {
		log.Fatal(err)
	}
	version = strings.Split(versionData[0], ": ")[1]
	date = strings.Split(versionData[1], ": ")[1]
	return
}

func showVersion(buildVersion string, buildDate string, buildHash string, buildOS string) {
	currentBuildVersion, currentBuildDate := getCurrentVersion()
	if currentBuildDate > buildDate {
		fmt.Printf("A new version (%s) exists. Upgrade with `cog upgrade`.\n\n", currentBuildVersion)
	}
	fmt.Printf(`Version: %s
Build Date: %s
Git Hash: %s
Build OS: %s
Install Method: %s
`, buildVersion, buildDate, buildHash, buildOS, buildInstallMethod)
}
