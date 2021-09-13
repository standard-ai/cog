package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cloud.google.com/go/storage"
)

var sshConfigFile string

var vaultAddress string
var vaultProxyHost string
var vaultIAPServiceAccount string
var vaultIAPClientID string

var gcsBucket string
var gcsFilename string
var binaryGCSBucket string
var binaryGCSPath string

var default_vaultAddress string
var default_vaultProxyHost string
var default_vaultIAPServiceAccount string
var default_vaultIAPClientID string

var default_gcsBucket string
var default_gcsFilename string

var default_binaryGCSBucket string
var default_binaryGCSPath string

var promptUser bool
var initVaultOnly bool

func askUser(text string, defaultValue string) string {
	fmt.Printf("%s (%s) > ", text, defaultValue)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.Replace(input, "\n", "", -1)
	if input == "" {
		input = defaultValue
	}
	return input
}

var initCommand = &cobra.Command{
	Use:   "init",
	Short: "Initialize cog",
	Long:  "Initialize cog",
	Run: func(cmd *cobra.Command, args []string) {
		if verbose > 0 {
			log.Infof("buildVersion: %s buildDate: %s buildHash: %s buildOS: %s", buildVersion, buildDate, buildHash, buildOS)
		}
		log.Info("Initializing cog")
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				usr, _ := user.Current()
				dir := usr.HomeDir + "/.config/cog/"
				log.Infof("Config file not found. Creating %s", dir)
				err := os.MkdirAll(dir, 0700)
				if err != nil {
					log.Fatal(err)
				}
				emptyFile, err := os.Create(dir + "cog.yaml")
				if err != nil {
					log.Fatal(err)
				}
				emptyFile.Close()
			} else {
				// Config file was found but another error was produced
				log.Fatal(err)
			}
		}

		if vaultAddress == "" {
			vaultAddress = default_vaultAddress
		}
		log.Info("Saving vault address: ", vaultAddress)
		viper.Set("vault_address", vaultAddress)

		if vaultProxyHost == "" {
			vaultProxyHost = default_vaultProxyHost
		}
		log.Info("Saving vault address: ", vaultProxyHost)
		viper.Set("vault_proxy_host", vaultProxyHost)

		if vaultIAPServiceAccount == "" {
			vaultIAPServiceAccount = default_vaultIAPServiceAccount
		}
		log.Info("Saving vault IAP service account: ", vaultIAPServiceAccount)
		viper.Set("vault_iap_service_account", vaultIAPServiceAccount)

		if vaultIAPClientID == "" {
			vaultIAPClientID = default_vaultIAPClientID
		}
		log.Info("Saving vault IAP client ID: ", vaultIAPClientID)
		viper.Set("vault_iap_client_id", vaultIAPClientID)

		gcsBucket = default_gcsBucket
		log.Info("Saving GCS Bucket: ", gcsBucket)
		viper.Set("gcs_bucket", gcsBucket)

		gcsFilename = default_gcsFilename
		log.Info("Saving GCS Filename: ", gcsFilename)
		viper.Set("gcs_filename", gcsFilename)

		binaryGCSBucket = default_binaryGCSBucket
		log.Info("Saving GCS Binary Bucket: ", binaryGCSBucket)
		viper.Set("gcs_binary_bucket", binaryGCSBucket)

		// Remove the suffix in case it is added in the configuration file
		binaryGCSPath = strings.TrimSuffix(default_binaryGCSPath, "/")
		log.Info("Saving GCS Binary Path: ", binaryGCSPath)

		viper.Set("gcs_binary_path", binaryGCSPath)

		err := viper.WriteConfig()
		if err != nil {
			log.Fatal(err)
		}
		if !initVaultOnly {
			storeSSHConfigFromGCS()
			storeKnownHostsFromGCS()
			cacheInventory(true)

			executePing(false)
		}
		// Set the user based on detected values

		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		bastionUser := viper.GetString("bastion_user")
		if bastionUser == "" {
			bastionUser = usr.Username
		}
		sshUser := viper.GetString("ssh_user")
		if sshUser == "" {
			sshUser = usr.Username
		}
		if promptUser {
			bastionUser = askUser("Default bastion ssh user", bastionUser)
			sshUser = askUser("Default non-bastion ssh user", sshUser)
		}

		viper.Set("bastion_user", bastionUser)
		viper.Set("ssh_user", sshUser)

		// Write config
		log.Info("Saving config")
		err = viper.WriteConfig()
		if err != nil {
			log.Fatal(err)
		}

	},
}

func getFileFromGCS(bucket string, fileName string) (file []byte, err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	rc, err := client.Bucket(bucket).Object(fileName).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func storeSSHConfigFromGCS() {
	gcsBucket := viper.GetString("gcs_bucket")
	gcsFilename := "ssh_config"
	if verbose > 0 {
		log.Infof("Downloading file from gs://%s/%s", gcsBucket, gcsFilename)
	}
	config, err := getFileFromGCS(gcsBucket, gcsFilename)
	if err != nil {
		log.Fatalf("%v", err)
	}

	_, err = os.Stat(sshConfigFile)
	if os.IsNotExist(err) {
		emptyFile, err := os.Create(sshConfigFile)
		if err != nil {
			log.Fatal(err)
		}
		emptyFile.Close()
	}
	file, err := os.Open(sshConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	newContents := ""
	oldContents := ""
	scanner := bufio.NewScanner(file)
	remove := false
	addedContents := false
	for scanner.Scan() {
		oldContents = oldContents + scanner.Text() + "\n"
		if scanner.Text() == "## BEGIN COG CONFIGURATION" {
			newContents = newContents + string(config)
			remove = true
			addedContents = true
		} else if scanner.Text() == "## END COG CONFIGURATION" {
			remove = false
		} else if !remove {
			newContents = newContents + scanner.Text() + "\n"
		}
	}
	if !addedContents {
		newContents = newContents + "\n" + string(config)
	}

	if newContents != oldContents {
		if verbose > 0 {
			log.Infof("Modifying %s", sshConfigFile)
		}
		if err := ioutil.WriteFile(sshConfigFile, []byte(newContents), 0600); err != nil {
			log.Fatalf("Error writing to %s: %v", sshConfigFile, err)
		}

	}
}

func storeKnownHostsFromGCS() {
	usr, _ := user.Current()
	dir := usr.HomeDir + "/.config/cog/"

	gcsBucket := viper.GetString("gcs_bucket")
	gcsFilename := "known_hosts"
	log.Infof("Downloading file from gs://%s/%s", gcsBucket, gcsFilename)
	hosts, err := getFileFromGCS(gcsBucket, gcsFilename)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if err := ioutil.WriteFile(dir+gcsFilename, hosts, 0600); err != nil {
		log.Fatalf("Error writing to %s: %v", dir+gcsFilename, err)
	}
}

func init() {
	usr, _ := user.Current()
	sshFile := usr.HomeDir + "/.ssh/config"

	initCommand.Flags().StringVarP(&sshConfigFile, "ssh-config-file", "c", sshFile, "SSH config file")
	initCommand.Flags().StringVarP(&vaultAddress, "vault-addr", "", "", "Vault address")
	initCommand.Flags().StringVarP(&vaultIAPServiceAccount, "vault-iap-service-account", "", "", "Vault IAP access service account")
	initCommand.Flags().StringVarP(&vaultIAPClientID, "vault-iap-client-id", "", "", "Vault IAP access client ID")
	initCommand.Flags().BoolVarP(&initVaultOnly, "vault-only", "", false, "Initialize cog for vault only (skips inventory, known_hosts, ping, et cetera")
	initCommand.Flags().BoolVarP(&promptUser, "prompt", "p", false, "Prompt user for input")
}
