package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"sort"
	"strings"

	"github.com/gobwas/glob"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var certGroup string
var sshGroup string
var sshUser, bastionUser string

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func getBastionHost(environment string) string {
	if verbose > 0 {
		log.Infof("Looking up bastion host for environment %s", environment)
	}

	bastionHost := ""
	bastionList := viper.GetStringMapString("bastion_map")
	if _, ok := bastionList[environment]; ok {
		bastionHost = bastionList[environment]
		if bastionHost == "" {
			log.Fatal("Unable to find bastion server")
		}
	} else {
		log.Fatalf("Unable to find bastion server for environment %s. Please run 'cog init'", environment)
	}

	return bastionHost
}

func getCertGroup(inventory Inventory, host string) (certGroup, sshGroup string) {
	if verbose > 0 {
		log.Info("Getting cert group and SSH group for ", host)
	}
	for _, group := range inventory.BastionMap {
		for _, s := range group.Hosts {
			if s == host {
				if verbose > 0 {
					log.Infof("Cert group: %s, SSH group: %s", group.SSHCertificateAuthority, group.SSHGroup)
				}
				return group.SSHCertificateAuthority, group.SSHGroup
			}
		}
	}

	for _, group := range inventory.BastionMap {
		for _, match := range group.Globs {
			g := glob.MustCompile(match, '.')
			if g.Match(host) {
				if verbose > 0 {
					log.Infof("Found glob match for %s: %s", host, match)
				}
				return group.SSHCertificateAuthority, group.SSHGroup
			}
		}
	}

	log.Fatal("Unable to discover SSH Group")
	return
}

func getHosts(inventory Inventory, args []string) ([]string, string, string) {
	var targetHosts []string
	for _, group := range inventory.BastionMap {

		for _, host := range group.Hosts {
			add := true
			for _, word := range args {
				if !strings.Contains(host, word) {
					add = false
					break
				}
			}
			if add {
				targetHosts = append(targetHosts, host)
			}
		}
	}

	for host, aliases := range inventory.Aliases {
		for _, alias := range aliases {
			add := true
			for _, word := range args {
				if !strings.Contains(alias, word) {
					add = false
					break
				}
			}
			if add && !stringInSlice(host, targetHosts) {
				targetHosts = append(targetHosts, host)
			}

		}
	}

	if len(targetHosts) == 1 {
		certGroup, sshGroup = getCertGroup(inventory, targetHosts[0])
		return targetHosts, certGroup, sshGroup
	}

	if len(targetHosts) == 0 && len(args) == 1 {
		found := false
		found, certGroup, sshGroup = checkGlob(inventory, args[0])
		if found {
			return args, certGroup, sshGroup
		}
	}
	return targetHosts, "", ""
}

func checkGlob(inventory Inventory, host string) (bool, string, string) {
	if verbose > 0 {
		log.Infof("Checking for glob match on %s", host)
	}
	for _, group := range inventory.BastionMap {
		for _, match := range group.Globs {
			g := glob.MustCompile(match, '.')
			if g.Match(host) {
				if verbose > 0 {
					log.Infof("Found glob match for %s: %s", host, match)
				}
				return true, group.SSHCertificateAuthority, group.SSHGroup
			}
		}
	}
	return false, "", ""
}

func sshGo(vaultUser, certGroup string, args []string) error {
	sshAuthSock, cleanup := getSock()
	if cleanup != nil {
		defer cleanup()
	}
	key, err := NewKey()
	if err != nil {
		return err
	}
	if err = vaultSignKey(certGroup, vaultUser, key); err != nil {
		// This is a [Sc] hack to line up with the Vault backend names
		if verbose > 0 {
			log.Info("Retrying signing with an ssh- prefix")
		}
		if err = vaultSignKey("ssh-"+certGroup, vaultUser, key); err != nil {
			return err
		}
	}
	err = addCertToAgent(key, sshAuthSock)
	defer removeCertFromAgent(key, sshAuthSock)
	if err != nil {
		return err
	}
	cmd := exec.Command("ssh", args...)
	if verbose > 0 {
		log.Info("Command: ", cmd)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		// 255 from ssh indicates something actually went wrong. Otherwise it's probably seeing a non
		// zero exit code from the last command run before closing the session
		if strings.Contains(err.Error(), "exit status") && !strings.Contains(err.Error(), "255") {
			return nil
		}
	}
	return nil
}

func init() {
	sshGoCommand.Flags().StringVarP(&sshUser, "user", "u", sshUser, "Username to SSH as (or COG_SSH_USER)")
	sshGoCommand.Flags().StringVarP(&bastionUser, "bastion-user", "b", bastionUser, "Username to SSH to the bastion as (OR COG_BASTION_USER)")
}

var sshGoCommand = &cobra.Command{
	Use:   "ssh HOST|FILTER [FILTER]*",
	Short: "SSH through a bastion",
	Long: `SSH through a bastion

While a specific host may be specified, you may also specify space separated
keywords. Filters will be applied, and if more than one server matches your
filters, the list will be displayed. If exactly one server matches your
filters, it will SSH into the desired target.
	`,
	DisableFlagParsing: false,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose > 0 {
			log.Infof("buildVersion: %s buildDate: %s buildHash: %s buildOS: %s", buildVersion, buildDate, buildHash, buildOS)
		}

		if len(args) > 0 && strings.Contains(args[0], "@") {
			argArray := strings.Split(args[0], "@")
			sshUser = argArray[0]
			args[0] = argArray[1]
			if verbose > 0 {
				log.Infof("Extracting ssh username (%s) from args", sshUser)
			}
		}

		if sshUser == "" {
			sshUser = os.Getenv("COG_SSH_USER")
			if sshUser == "" {
				sshUser = viper.GetString("ssh_user")
			}
		}

		cacheInventory(false)
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

		hosts, certGroup, sshGroup := getHosts(inventory, args)
		if len(hosts) == 0 {
			log.Fatal("No hosts found that match your search terms")
		}

		if len(hosts) > 1 {

			sort.Strings(hosts)

			for _, host := range hosts {
				fmt.Print(host)
				if len(inventory.Aliases[host]) > 0 {
					fmt.Print(" # ", strings.Join(inventory.Aliases[host], " "))
				}
				fmt.Println()
			}
			return
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

		if sshUser != "" {
			hosts[0] = sshUser + "@" + hosts[0]
		}
		if verbose > 1 {
			verboseString := "-"
			for i := 1; i < verbose; i++ {
				verboseString = verboseString + "v"
			}

			if stringInSlice(bastionHost, hosts) {
				err = sshGo(bastionUser, certGroup, append([]string{verboseString}, hosts...))
			} else {
				err = sshGo(bastionUser, certGroup, append([]string{verboseString, "-J", bastionUser + "@" + bastionHost}, hosts...))
			}

		} else {
			if stringInSlice(bastionHost, hosts) {
				err = sshGo(bastionUser, certGroup, hosts)
			} else {
				err = sshGo(bastionUser, certGroup, append([]string{"-J", bastionUser + "@" + bastionHost}, hosts...))
			}
		}
		if err != nil {
			log.Fatal(err)
		}
	},
}
