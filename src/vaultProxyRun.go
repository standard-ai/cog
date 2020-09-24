package main

import (
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func vaultProxyRun(proxyAddress, bundleFile string, args []string) {
	if verbose > 0 {
		log.Infof("Running command under vault proxy: %s\n", strings.Join(args, " "))
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(
		os.Environ(),
		"VAULT_ADDR="+proxyAddress,
		"VAULT_CACERT="+bundleFile,
		"REQUESTS_CA_BUNDLE="+bundleFile,
		"COG_VAULT_PROXY=true",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
