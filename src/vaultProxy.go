package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yosh/gocertifi"

	"github.com/spf13/viper"
)

const vaultProxyPort = "16200"

const vaultProxyTempDir = "cog-vault-proxy"
const vaultProxyBundleFile = "ca-vault-proxy.crt"

func writeCABundle(bundleFile string, ca *ephemeralCA) {
	bundleOut, err := os.Create(bundleFile)
	if err != nil {
		log.Fatal("Unable to create CA certificate bundle file", err)
	}

	// Include gocertifi roots so Python tools can talk to the rest of the world as well
	bundleOut.WriteString(gocertifi.PEMCerts)
	bundleOut.WriteString("# Vault Proxy Ephemeral CA\n")
	bundleOut.Write(ca.CertPEM)
	bundleOut.Close()
}

func startProxy(cert *ephemeralCert) {
	iapCreds := NewIAPCredentials()

	// Authenticate ourselves into vault if we need to
	client := authedClient(iapCreds)
	if client.Token() == "" {
		log.Fatal("Failed to authenticate with vault")
	}

	p := iapProxy(iapCreds)

	mux := http.NewServeMux()
	mux.Handle("/", p)

	serverCert, err := tls.X509KeyPair(cert.CertPEM, cert.KeyPEM)
	if err != nil {
		log.Fatal("Failed to generate an X.509 key pair: ", err)
	}

	serverTLSConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}

	s := &http.Server{
		Addr:      ":" + vaultProxyPort,
		TLSConfig: serverTLSConfig,
		Handler:   mux,
	}

	go func() {
		log.Fatal(s.ListenAndServeTLS("", ""))
	}()
}

func setupVaultProxy() (proxyAddress, bundleFile, tempDir string, alreadySet bool) {

	checkVaultAddr := os.Getenv("VAULT_ADDR")
	checkVaultCACert := os.Getenv("VAULT_CACERT")
	checkRequestsCABundle := os.Getenv("REQUESTS_CA_BUNDLE")
	if checkVaultAddr != "" && checkVaultCACert != "" && checkRequestsCABundle != "" {
		return checkVaultAddr, checkRequestsCABundle, checkVaultCACert, true
	}

	vaultProxyHost := viper.GetString("vault_proxy_host")
	proxyAddress = fmt.Sprintf("https://%s:%s", vaultProxyHost, vaultProxyPort)

	ca := newEphemeralCA()
	cert := newEphemeralCert(vaultProxyHost, ca)

	tempDir, err := ioutil.TempDir("", vaultProxyTempDir)
	if err != nil {
		log.Fatal("Unable to create temp dir", err)
	}

	bundleFile = filepath.Join(tempDir, vaultProxyBundleFile)
	writeCABundle(bundleFile, ca)

	startProxy(cert)
	return proxyAddress, bundleFile, tempDir, false
}

var nonInteractive bool

func init() {
	vaultProxyCommand.AddCommand(vaultProxyEnvCommand)
	vaultProxyCommand.AddCommand(vaultProxyShellCommand)
	vaultProxyCommand.AddCommand(vaultProxyRunCommand)
	vaultProxyCommand.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Run in non-interactive mode (auth failures would lead to hard fail)")
}

var vaultProxyEnvCommand = &cobra.Command{
	Use:   "env",
	Short: "Proxy access to Vault through IAP and set up env vars",
	Long:  "Proxy access to Vault through IAP and set up env vars",
	Run: func(cmd *cobra.Command, args []string) {
		if verbose > 0 {
			log.Infof("buildVersion: %s buildDate: %s buildHash: %s buildOS: %s", buildVersion, buildDate, buildHash, buildOS)
		}

		proxyAddress, bundleFile, tempDir, alreadySet := setupVaultProxy()
		if alreadySet {
			fmt.Println("You are already inside an activated vault-proxy environment.")
		} else {
			defer os.RemoveAll(tempDir)
			vaultProxyScript(proxyAddress, bundleFile, tempDir)
			fmt.Println("\nVault proxy terminated")
		}
	},
}

var vaultProxyRunCommand = &cobra.Command{
	Use:   "run",
	Short: "Proxy access to Vault through IAP and run a command",
	Long:  "Proxy access to Vault through IAP and run a command",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose > 0 {
			log.Infof("buildVersion: %s buildDate: %s buildHash: %s buildOS: %s", buildVersion, buildDate, buildHash, buildOS)
		}

		proxyAddress, bundleFile, tempDir, alreadySet := setupVaultProxy()
		if !alreadySet {
			defer os.RemoveAll(tempDir)
		}
		vaultProxyRun(proxyAddress, bundleFile, args)
	},
}

var vaultProxyShellCommand = &cobra.Command{
	Use:   "shell",
	Short: "Proxy access Vault through IAP and run a shell",
	Long:  "Proxy access Vault through IAP and run a shell",
	Run: func(cmd *cobra.Command, args []string) {
		if verbose > 0 {
			log.Infof("buildVersion: %s buildDate: %s buildHash: %s buildOS: %s", buildVersion, buildDate, buildHash, buildOS)
		}

		proxyAddress, bundleFile, tempDir, alreadySet := setupVaultProxy()
		if alreadySet {
			fmt.Println("You are already inside an activated vault-proxy environment.")
		} else {
			defer os.RemoveAll(tempDir)
			vaultProxyShell(proxyAddress, bundleFile)
			fmt.Println("\nVault proxy terminated")
		}
	},
}

var vaultProxyCommand = &cobra.Command{
	Use:   "vault-proxy",
	Short: "Proxy access to Vault through IAP",
	Long:  "Proxy access to Vault through IAP",
}
