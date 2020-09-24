package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func removeVaultToken() error {
	usr, _ := user.Current()
	dir := usr.HomeDir
	tokenPath := dir + "/.vault-token"
	if fileExists(tokenPath) {
		log.Infof("Removing %s\n", tokenPath)
		err := os.Remove(tokenPath)
		if err != nil {
			log.Fatal(err)
		}

	} else {
		log.Infof("No %s found\n", tokenPath)
	}

	return nil
}

func closeSSHControlPathSockets() error {
	usr, _ := user.Current()
	dir := usr.HomeDir
	fileGlob := dir + "/.ssh/a-*"
	files, err := filepath.Glob(fileGlob)
	if err != nil {
		log.Fatal(err)
	}
	if len(files) == 0 {
		log.Infof("No SSH ControlMaster files found in %s", dir+"/.ssh")
	} else {
		for _, f := range files {
			// This explains that we need to pass a bogus argument to kill the ControlMaster connection. Weird.
			// https://unix.stackexchange.com/questions/24005/how-to-close-kill-ssh-controlmaster-connections-manually
			controlMasterArguments := strings.Split(fmt.Sprintf("-o ControlPath=%s -O exit BOGUSARG", f), " ")

			cmd := exec.Command("ssh", controlMasterArguments...)
			log.Info("Command: ", cmd)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			err = cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return nil
}

var logoutCommand = &cobra.Command{
	Use:   "logout",
	Short: "Remove vault tokens and all SSH ControlMaster connections",
	Long:  "Remove vault tokens and all SSH ControlMaster connections",
	Run: func(cmd *cobra.Command, args []string) {
		if verbose > 0 {
			log.Infof("buildVersion: %s buildDate: %s buildHash: %s buildOS: %s", buildVersion, buildDate, buildHash, buildOS)
		}

		err := removeVaultToken()
		if err != nil {
			log.Fatal(err)
		}

		err = closeSSHControlPathSockets()
		if err != nil {
			log.Fatal(err)
		}

	},
}
