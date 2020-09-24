package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"text/template"

	log "github.com/sirupsen/logrus"
)

const vaultProxyScriptTemplate = `# Source this file into your shell

vpdone() {
    if [ -n "${_OLD_COG_PROXY_VAULT_ADDR:-}" ] ; then
        VAULT_ADDR="${_OLD_COG_PROXY_VAULT_ADDR:-}"
        export VAULT_ADDR
        unset _OLD_COG_PROXY_VAULT_ADDR
    else
        unset VAULT_ADDR
    fi
    if [ -n "${_OLD_COG_PROXY_VAULT_CACERT:-}" ] ; then
        VAULT_CACERT="${_OLD_COG_PROXY_VAULT_CACERT:-}"
        export VAULT_CACERT
        unset _OLD_COG_PROXY_VAULT_CACERT
    else
        unset VAULT_CACERT
    fi
    if [ -n "${_OLD_COG_PROXY_REQUESTS_CA_BUNDLE:-}" ] ; then
        REQUESTS_CA_BUNDLE="${_OLD_COG_PROXY_REQUESTS_CA_BUNDLE:-}"
        export REQUESTS_CA_BUNDLE
        unset _OLD_COG_PROXY_REQUESTS_CA_BUNDLE
    else
        unset REQUESTS_CA_BUNDLE
    fi

    if [ -n "${_OLD_COG_PROXY_PS1:-}" ] ; then
        PS1="${_OLD_COG_PROXY_PS1:-}"
        export PS1
        unset _OLD_COG_PROXY_PS1
    fi

    kill -USR1 "$COG_VAULT_PROXY_PID" >/dev/null 2>&1
    unset COG_VAULT_PROXY_PID

    unset -f vpdone
}

COG_VAULT_PROXY_PID="{{ .CogPid }}"
export COG_VAULT_PROXY_PID

_OLD_COG_PROXY_VAULT_ADDR="$VAULT_ADDR"
VAULT_ADDR="{{ .ProxyAddress }}"
export VAULT_ADDR
_OLD_COG_PROXY_VAULT_CACERT="$VAULT_CACERT"
VAULT_CACERT="{{ .BundleFile }}"
export VAULT_CACERT
_OLD_COG_PROXY_REQUESTS_CA_BUNDLE="$REQUESTS_CA_BUNDLE"
REQUESTS_CA_BUNDLE="{{ .BundleFile }}"
export REQUESTS_CA_BUNDLE

_OLD_COG_PROXY_PS1="${PS1:-}"
PS1="(vpenv) ${PS1:-}"
export PS1
`

const vaultProxyInstructions = `Vault proxy initialized!

Run the following command in the terminal in which you want to use Vault:

source {{ .ScriptFile }}

This will setup the following environment variables:

VAULT_ADDR={{ .ProxyAddress }}
VAULT_CACERT={{ .BundleFile }}
REQUESTS_CA_BUNDLE={{ .BundleFile }}
COG_VAULT_PROXY=true # Useful for conditionally setting PS1

When you're done, run "vpdone" in the terminal. The above variables
will be unset and the proxy will terminate. You may also press Ctrl+C
here to terminate the proxy.
`

type proxyScriptSetup struct {
	ProxyAddress string
	BundleFile   string
	TempDir      string
	ScriptFile   string
	CogPid       int
}

func vaultProxyScript(proxyAddress, bundleFile, tempDir string) {
	scriptFile := filepath.Join(tempDir, "env-proxy")

	setup := proxyScriptSetup{
		ProxyAddress: proxyAddress,
		BundleFile:   bundleFile,
		TempDir:      tempDir,
		ScriptFile:   scriptFile,
		CogPid:       os.Getpid(),
	}

	scriptOut, err := os.Create(scriptFile)
	if err != nil {
		log.Fatal("Unable to create env script", err)
	}
	script := template.Must(template.New("").Parse(vaultProxyScriptTemplate))
	script.Execute(scriptOut, setup)
	scriptOut.Close()

	inst := template.Must(template.New("").Parse(vaultProxyInstructions))
	inst.Execute(os.Stdout, setup)

	// Just wait around while the user does their thing in another terminal
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGUSR1, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-termChan
}
