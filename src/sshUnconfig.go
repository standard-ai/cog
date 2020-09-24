package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"os/user"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var sshUnconfigCommand = &cobra.Command{
	Use:                "ssh-unconfig",
	Short:              "Revert local SSH configuration",
	Long:               "Revert local SSH configuration",
	DisableFlagParsing: false,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose > 0 {
			log.Infof("buildVersion: %s buildDate: %s buildHash: %s buildOS: %s", buildVersion, buildDate, buildHash, buildOS)
		}

		file, err := os.Open(sshConfigFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		newContents := ""
		scanner := bufio.NewScanner(file)
		remove := false
		removedContents := false
		for scanner.Scan() {
			if scanner.Text() == "## BEGIN COG CONFIGURATION" {
				remove = true
				removedContents = true
			} else if scanner.Text() == "## END COG CONFIGURATION" {
				remove = false
			} else if !remove {
				newContents = newContents + scanner.Text() + "\n"
			}
		}
		if removedContents {
			if verbose > 0 {
				log.Infof("Modifying %s", sshConfigFile)
			}
			if err := ioutil.WriteFile(sshConfigFile, []byte(newContents), 0600); err != nil {
				log.Fatalf("Error writing to %s: %v", sshConfigFile, err)
			}
		}
	},
}

func init() {
	usr, _ := user.Current()
	sshFile := usr.HomeDir + "/.ssh/config"

	sshUnconfigCommand.Flags().StringVarP(&sshConfigFile, "ssh-config-file", "c", sshFile, "SSH config file")
}
