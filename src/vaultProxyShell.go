package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	log "github.com/sirupsen/logrus"
)

const vaultProxyShellInstructions = `Vault proxy shell initialized!

You may now run commands that access Vault here. The shell has the
following environment variables set:

VAULT_ADDR={{ .ProxyAddress }}
VAULT_CACERT={{ .BundleFile }}
REQUESTS_CA_BUNDLE={{ .BundleFile }}
COG_VAULT_PROXY=true # Useful for conditionally setting PS1

When you're done, exit this shell and the proxy will terminate.

`

type proxyShellSetup struct {
	ProxyAddress string
	BundleFile   string
}

func vaultProxyShell(proxyAddress, bundleFile string) {
	inst := template.Must(template.New("").Parse(vaultProxyShellInstructions))
	inst.Execute(os.Stdout, proxyShellSetup{
		ProxyAddress: proxyAddress,
		BundleFile:   bundleFile,
	})

	shell := os.Getenv("SHELL")
	if len(shell) == 0 {
		shell = "/bin/bash"
	}

	// Spawn a login shell
	baseShell := "-" + filepath.Base(shell)

	env := append(
		os.Environ(),
		"VAULT_ADDR="+proxyAddress,
		"VAULT_CACERT="+bundleFile,
		"REQUESTS_CA_BUNDLE="+bundleFile,
		"COG_VAULT_PROXY=true",
	)

	cmd := &exec.Cmd{
		Path:   shell,
		Args:   []string{baseShell},
		Env:    env,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	err := cmd.Start()
	if err != nil {
		log.Fatal("Could not start login shell: ", err)
	}

	// Wait around for the user to do their thing and exit the shell
	cmd.Wait()
}
