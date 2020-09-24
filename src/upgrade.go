package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var forceUpgrade bool

func init() {
	upgradeCommand.Flags().BoolVarP(&forceUpgrade, "force", "f", false, "Force upgrade")
}

var upgradeCommand = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade cog",
	Long:  "Upgrade cog",
	Run: func(cmd *cobra.Command, args []string) {
		if verbose > 0 {
			log.Infof("buildVersion: %s buildDate: %s buildHash: %s buildOS: %s", buildVersion, buildDate, buildHash, buildOS)
		}
		path := viper.GetString("gcs_binary_path")
		binaryGCSBucket := viper.GetString("gcs_binary_bucket")

		if path == "" || binaryGCSBucket == "" {
			log.Info("gcs_binary_bucket or gcs_binary_path is not defined, cannot retrieve binary")
			return
		}

		currentBuildVersion, currentBuildDate := getCurrentVersion()
		if forceUpgrade || currentBuildDate > buildDate {
			log.Printf("Upgrading to %s ...\n", currentBuildVersion)

			if runtime.GOOS == "linux" {
				path = path + "linux/cog"
			} else {
				path = path + "darwin/cog"
			}

			fullPath, err := exec.LookPath(os.Args[0])
			if err != nil {
				log.Fatal(err)
			}

			now := time.Now().Format("2006-01-02-15.04.05")
			tempFile := fullPath + "." + now
			log.Printf("Downloading cog to %s\n", tempFile)
			binaryBytes, err := getFileFromGCS(binaryGCSBucket, path)
			err = ioutil.WriteFile(tempFile, binaryBytes, 0700)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("Moving %s to %s ...\n", tempFile, fullPath)
			err = os.Rename(tempFile, fullPath)
			if err != nil {
				log.Fatal(err)
			}
			cacheInventory(true)
		} else {
			fmt.Printf("%s is the current version. No upgrade necessary.\n", buildVersion)
		}
	},
}
