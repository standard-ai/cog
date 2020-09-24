package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var buildDate string
var buildHash string
var buildVersion string
var buildOS string
var buildInstallMethod string
var verbose int

var rootCmd = &cobra.Command{
	Use:              "cog",
	TraverseChildren: true,
}

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	go func() {
		for range c {
			if verbose > 0 {
				log.Info("Caught SIGHUP")
			}
		}
	}()

	viper.SetConfigName("cog")                // name of config file (without extension)
	viper.SetConfigType("yaml")               // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("$HOME/.config/cog/") // call multiple times to add many search paths
	viper.AddConfigPath(".")                  // optionally look for config in the working directory

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if len(os.Args) > 1 && os.Args[1] != "init" {
				fmt.Println("cog is not initialized. Please run 'cog init'")
			}
		} else {
			log.Fatal(err)
			// Config file was found but another error was produced
		}
	}

	rootCmd.Flags().CountVarP(&verbose, "verbose", "v", "verbose output (use -vv to increase verbosity)")
	rootCmd.AddCommand(versionCommand)
	rootCmd.AddCommand(sshGoCommand)
	rootCmd.AddCommand(loginCommand)
	rootCmd.AddCommand(logoutCommand)

	rootCmd.AddCommand(cacheInventoryCommand)
	rootCmd.AddCommand(pingCommand)
	rootCmd.AddCommand(initCommand)
	rootCmd.AddCommand(sshProxyCommand)
	rootCmd.AddCommand(upgradeCommand)
	rootCmd.AddCommand(sshConfigCommand)
	rootCmd.AddCommand(sshUnconfigCommand)
	rootCmd.AddCommand(vaultProxyCommand)
	rootCmd.AddCommand(completionCommand)
	rootCmd.Execute()
}
