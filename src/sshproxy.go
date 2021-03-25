package main

import (
	"os"
	"os/user"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var sshProxyCommand = &cobra.Command{
	Use:                "sshproxy",
	Short:              "Create ProxyCommand for SSH",
	Long:               "Create ProxyCommand for SSH",
	Args:               cobra.MinimumNArgs(2),
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose > 0 {
			log.Infof("buildVersion: %s buildDate: %s buildHash: %s buildOS: %s", buildVersion, buildDate, buildHash, buildOS)
		}

		fileName, err := getInventoryFilePath()
		if err != nil {
			log.Fatal(err)
		}

		var inventory Inventory
		inventory, err = loadInventoryFile(inventory, fileName)
		if err != nil {
			log.Fatal(err)
		}

		customInventoryFile := os.Getenv("COG_CUSTOM_INVENTORY")
		if customInventoryFile == "" {
			customInventoryFile = "~/.config/cog/custom-inventory.yaml"
		}
		if fileExists(customInventoryFile) {
			inventory, err = loadInventoryFile(inventory, customInventoryFile)
			if err != nil {
				log.Fatal(err)
			}
		}

		targetHost := strings.Split(args[1], ":")[0]
		hosts := []string{}
		hosts, certGroup, sshGroup = getHosts(inventory, []string{targetHost})
		if len(hosts) != 1 {
			log.Fatalf("Can't find %s in cog inventory.", targetHost)
		}

		bastionHost := getBastionHost(sshGroup + "." + certGroup)

		if bastionUser == "" {
			bastionUser = os.Getenv("COG_BASTION_USER")
			if bastionUser == "" {
				bastionUser = viper.GetString("bastion_user")
			}
			if bastionUser == "" {
				usr, err := user.Current()
				if err != nil {
					log.Fatal(err)
				}
				bastionUser = usr.Username
			}
		}
		if verbose > 0 {
			log.Info("bastion user: ", bastionUser)
			log.Info("ssh user: ", sshUser)
		}
		err = sshGo(bastionUser, certGroup, append([]string{bastionUser + "@" + bastionHost}, args...))
		if err != nil {
			log.Fatal(err)
		}
	},
}
