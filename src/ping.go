package main

import (
	"os"
	"runtime"
	"strings"

	"os/exec"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func executePing(saveConfig bool) {
	type BestServer struct {
		environment string
		server      string
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

	if len(inventory.BastionMap) == 0 {
		log.Fatal("No bastion inventory found. Please run 'cog init'")
	}

	bastionList := make(map[string]string)
	best := make(chan BestServer)

	for _, group := range inventory.BastionMap {
		go func(environment string, servers []string) {
			var t BestServer
			t.environment = environment
			t.server = getBestServers(environment, servers)
			best <- t
		}(group.SSHGroup+"."+group.SSHCertificateAuthority, group.Bastions)

	}

	for range inventory.BastionMap {
		t := <-best

		log.Infof("Best bastion server for %s: %s", t.environment, t.server)
		bastionList[t.environment] = t.server
	}
	viper.Set("bastion_map", bastionList)
	if saveConfig {
		log.Info("Saving config")
		err = viper.WriteConfig()
		if err != nil {
			log.Fatal(err)
		}
	}

}

var pingCommand = &cobra.Command{
	Use:   "ping",
	Short: "Calculate and cache the best bastion server to use",
	Long:  "Ping available bastion servers and store the 'best' one in you configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		if verbose > 0 {
			log.Infof("buildVersion: %s buildDate: %s buildHash: %s buildOS: %s", buildVersion, buildDate, buildHash, buildOS)
		}

		executePing(true)
	},
}

func getBestServers(environment string, servers []string) string {

	type Ping struct {
		server string
		time   int
		output string
	}

	const NUM_PINGS = 10
	const PING_INTERVAL = 1

	if len(servers) == 0 {
		log.Info("No servers to ping")
		return ""
	} else if len(servers) == 1 {
		log.Infof("%s: only one server to use as bastion host", environment)
		return servers[0]
	}

	log.Infof("%s: pinging %d servers %d times, sleeping %d second(s) in between (maximum of %d seconds)", environment, len(servers), NUM_PINGS, PING_INTERVAL, PING_INTERVAL*NUM_PINGS+10)

	messages := make(chan Ping)

	const GOOS = runtime.GOOS

	for index, server := range servers {
		go func(index int, server string) {
			var t Ping
			t.server = server
			t.time = 1000000
			out, err := exec.Command("ping", "-c", strconv.Itoa(NUM_PINGS), "-i", strconv.Itoa(PING_INTERVAL), server).Output()
			if err != nil {
				log.Warnf("Unable to ping %s: %v", server, err)
			}
			t.output = string(out)
			for _, line := range strings.Split(string(out), "\n") {
				if strings.HasPrefix(line, "round-trip") || strings.HasPrefix(line, "rtt") {
					// round-trip min/avg/max/stddev = 32.526/32.526/32.526/0.000 ms
					// rtt min/avg/max/mdev = 10.0/1.0/1.0/1.0 ms
					averageTimeString := strings.Split(line, " = ")[1]
					averageTimeString = strings.Split(averageTimeString, "/")[1]
					averageTime, err := strconv.ParseFloat(averageTimeString, 64)
					if err != nil {
						log.Fatal(err)
					}
					averageTimeInt := int(averageTime * 1000)
					t.time = averageTimeInt
				}
			}
			messages <- t
		}(index, server)
	}
	var bestServer Ping
	for range servers {
		t := <-messages
		if verbose > 0 {
			log.Info(t.output)
		}
		if bestServer.server == "" || t.time < bestServer.time {
			bestServer = t
		}
	}

	return bestServer.server
}
