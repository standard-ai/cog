package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v1"
)

const defaultMaxInventoryFileAge = 86400

type Bastion struct {
	Bastions                []string `yaml:"bastions"`
	Globs                   []string `yaml:"globs"`
	Hosts                   []string `yaml:"hosts"`
	SSHCertificateAuthority string   `yaml:"ssh_ca"`
	SSHGroup                string   `yaml:"ssh_group"`
}

var forceInventory bool

type Inventory struct {
	BastionMap []Bastion           `yaml:"map"`
	Aliases    map[string][]string `yaml:"aliases"`
}

func fileExists(fileName string) bool {
	if strings.HasPrefix(fileName, "~/") {
		usr, _ := user.Current()
		dir := usr.HomeDir
		fileName = filepath.Join(dir, fileName[2:])
	}

	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func saveInventoryToPath(inventory Inventory, inventoryFile string) (err error) {
	data, err := yaml.Marshal(inventory)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(inventoryFile, data, 0600)
	return err
}

func fileTooOld(inventoryFile string, age uint) bool {
	file, err := os.Stat(inventoryFile)
	if err != nil {
		log.Fatal(err)
	}

	expireTime := file.ModTime().Add(time.Second * time.Duration(age))
	now := time.Now()
	if expireTime.Before(now) {
		return true
	}
	return false
}

func getInventoryFilePath() (inventoryFile string, err error) {
	inventoryFile = os.Getenv("COG_INVENTORY")
	if inventoryFile != "" {
		return inventoryFile, nil
	}

	usr, _ := user.Current()
	dir := usr.HomeDir

	files := [6]string{"./inventory.yml", "./inventory.yaml", "~/inventory.yml", "~/inventory.yaml", "~/.config/cog/inventory.yml", "~/.config/cog/inventory.yaml"}
	for _, file := range files {
		if strings.HasPrefix(file, "~/") {
			file = filepath.Join(dir, file[2:])
		}

		if fileExists(file) {
			return file, nil
		}
	}
	return inventoryFile, fmt.Errorf("Unable to discover a valid inventory file and COG_INVENTORY is not set.")
}

func loadInventoryFile(inventory Inventory, fileName string) (Inventory, error) {
	var data Inventory

	if strings.HasPrefix(fileName, "~/") {
		usr, _ := user.Current()
		dir := usr.HomeDir
		fileName = filepath.Join(dir, fileName[2:])
	}

	if verbose > 0 {
		log.Info("Reading ", fileName)
	}
	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		return data, err
	}
	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		return data, err
	}
	// Now add data to inventory
	for k, v := range data.Aliases {
		if inventory.Aliases == nil {
			inventory.Aliases = make(map[string][]string)
		}
		inventory.Aliases[k] = v
	}
	for _, group := range data.BastionMap {
		inventory.BastionMap = append(inventory.BastionMap, group)
	}
	return inventory, nil
}

func getInventoryFromGCS(bucket string, fileName string) (inventory Inventory, err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return inventory, err
	}
	rc, err := client.Bucket(bucket).Object(fileName).NewReader(ctx)
	if err != nil {
		return inventory, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return inventory, err
	}

	err = yaml.Unmarshal(data, &inventory)
	if err != nil {
		return inventory, err
	}
	return inventory, nil
}

func cacheInventory(forceUpgrade bool) {
	var maxInventoryFileAge uint
	if envInventoryAge := os.Getenv("COG_MAX_INVENTORY_AGE"); envInventoryAge != "" {
		x, err := strconv.ParseUint(envInventoryAge, 10, 32)
		if err != nil {
			log.Fatal(err)
		}
		maxInventoryFileAge = uint(x)
	} else {
		maxInventoryFileAge = defaultMaxInventoryFileAge
	}

	inventoryFile, err := getInventoryFilePath()
	if err != nil {
		usr, _ := user.Current()
		dir := usr.HomeDir
		inventoryFile = dir + "/.config/cog/inventory.yaml"
	}
	if forceUpgrade || verbose > 0 {
		log.Infof("Using inventory: %s", inventoryFile)
	}

	if !fileExists(inventoryFile) || fileTooOld(inventoryFile, maxInventoryFileAge) || forceUpgrade {
		if !fileExists(inventoryFile) {
			log.Infof("%s does not exist", inventoryFile)
		} else {
			if forceUpgrade || verbose > 0 {
				log.Infof("%s is too old", inventoryFile)
			}
		}
		gcsBucket := viper.GetString("gcs_bucket")
		gcsFilename := viper.GetString("gcs_filename")
		if gcsBucket == "" || gcsFilename == "" {
			if verbose > 0 {
				log.Info("No inventory GCS sources set (gcs_bucket and gcs_filename)")
			}
			return
		}
		if forceUpgrade || verbose > 0 {
			log.Infof("Downloading file from gs://%s/%s", gcsBucket, gcsFilename)
		}
		inventory, err := getInventoryFromGCS(gcsBucket, gcsFilename)
		if err != nil {
			log.Fatal(err, "\nMake sure the gcloud SDK is installed and you have run `gcloud init` and `gcloud auth application-default login`")
		}
		if forceUpgrade || verbose > 0 {
			log.Info("Saving inventory to ", inventoryFile)
		}
		err = saveInventoryToPath(inventory, inventoryFile)
		if err != nil {
			log.Fatal(err)
		}
	}
}

var cacheInventoryCommand = &cobra.Command{
	Use:   "update-inventory",
	Short: "Download latest inventory file",
	Long:  "Download latest inventory file from GCS",
	Run: func(cmd *cobra.Command, args []string) {
		if verbose > 0 {
			log.Infof("buildVersion: %s buildDate: %s buildHash: %s buildOS: %s", buildVersion, buildDate, buildHash, buildOS)
		}
		cacheInventory(forceInventory)
	},
}

func init() {
	cacheInventoryCommand.Flags().BoolVarP(&forceInventory, "force", "f", false, "Force update")
}
