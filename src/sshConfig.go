package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func copyFile(src, dst string) error {
	usr, _ := user.Current()
	dir := usr.HomeDir

	if strings.HasPrefix(src, "~/") {
		src = filepath.Join(dir, src[2:])
	}

	if strings.HasPrefix(dst, "~/") {
		dst = filepath.Join(dir, dst[2:])
	}

	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

var sshConfigCommand = &cobra.Command{
	Use:                "ssh-config",
	Short:              "Configure local SSH",
	Long:               "Configure local SSH",
	DisableFlagParsing: false,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose > 0 {
			log.Infof("buildVersion: %s buildDate: %s buildHash: %s buildOS: %s", buildVersion, buildDate, buildHash, buildOS)
		}

		storeSSHConfigFromGCS()
	},
}

func init() {
	usr, _ := user.Current()
	sshFile := usr.HomeDir + "/.ssh/config"

	sshConfigCommand.Flags().StringVarP(&sshConfigFile, "ssh-config-file", "c", sshFile, "SSH config file")
}
