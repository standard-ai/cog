package main

import (
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	vault "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

func vaultSignKey(sshMountPoint, sshRole string, key *Key) error {
	if verbose > 0 {
		log.Info("SSH Mount Point: ", sshMountPoint)
		log.Info("SSH Role: ", sshRole)
	}
	client := authedClient(nil)
	c := client.SSHWithMountPoint(sshMountPoint)
	data := map[string]interface{}{
		"public_key": strings.TrimSuffix(string(key.pub), "\n"),
	}
	secret, err := c.SignKey(sshRole, data)
	if err != nil {
		return err
	}
	certText := []byte(secret.Data["signed_key"].(string))
	cert, _, _, _, err := ssh.ParseAuthorizedKey(certText)
	if err != nil {
		return err
	}
	key.cert = cert.(*ssh.Certificate)
	return nil
}

func writeTokenStore(token string) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	vaultTokenFile := usr.HomeDir + "/.vault-token"
	err = ioutil.WriteFile(vaultTokenFile, []byte(token), 0644)
	if err != nil {
		log.Fatal("Error writing ~/.vault-token")
	}
}

func fetchTokenStore() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	vaultTokenFile := usr.HomeDir + "/.vault-token"
	if _, err := os.Stat(vaultTokenFile); err != nil {
		if _, err := os.Create(vaultTokenFile); err != nil {
			log.Fatal("Could not create ~/.vault-token: ", err)
		}
	}
	data, err := ioutil.ReadFile(vaultTokenFile)
	if err != nil {
		log.Fatal("Error reading ~/.vault-token: ", err)
	}
	data = []byte(strings.Trim(string(data), "\n"))
	return string(data)
}

// Return true if we found a token in env or ~/.vault-token
func setTokenFromStore(client *vault.Client) bool {
	token := client.Token()
	if token != "" {
		if verbose > 0 {
			log.Info("Found token in env")
		}
		return true
	}
	token = fetchTokenStore()
	client.SetToken(token)
	if token != "" {
		if verbose > 0 {
			log.Info("Found token in ~/.vault-token")
		}
		return true
	}
	return false
}

func canLookupToken(client *vault.Client) bool {
	_, err := client.Auth().Token().LookupSelf()
	if err != nil {
		if strings.Contains(err.Error(), "Code: 403") {
			return false
		} else if strings.Contains(err.Error(), "cannot be used") {
			return false
		}
		log.Fatal(err)
	}
	return true
}

func envClient(iapCreds *IAPCredentials) *vault.Client {
	config := vault.DefaultConfig()
	err := config.ReadEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	config.Address = viper.GetString("vault_address")
	if verbose > 0 {
		log.Info("Using vault server ", config.Address)
	}
	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatal("Error creating vault client: ", err)
	}
	serviceAccount := viper.GetString("vault_iap_service_account")
	clientID := viper.GetString("vault_iap_client_id")
	if serviceAccount != "" && clientID != "" {
		if verbose > 0 {
			log.Info("Adding IAP authorization to vault client")
		}
		if iapCreds == nil {
			iapCreds = NewIAPCredentials()
		}
		client.AddHeader("Authorization", iapCreds.BearerToken())
	} else {
		if verbose > 0 {
			log.Info("No IAP authorization detected")
		}
	}
	return client
}

func authedClient(iapCreds *IAPCredentials) *vault.Client {
	client := envClient(iapCreds)
	if setTokenFromStore(client) && canLookupToken(client) {
		if verbose > 0 {
			log.Info("Using existing credentials")
		}
		return client
	} else if nonInteractive {
		log.Fatal("Cached token invalid in non-interactive mode - aborting.")
	} else {
		serviceAccount := viper.GetString("vault_iap_service_account")
		clientID := viper.GetString("vault_iap_client_id")
		if serviceAccount != "" && clientID != "" {
			if verbose > 0 {
				log.Info("Cached token invalid, attempting vault oidc login")
			}
			client.SetToken("")
			token := oidcLogin(client)
			writeTokenStore(token)
			client.SetToken(token)
		}
	}
	if !canLookupToken(client) {
		log.Fatal("Could not authenticate with vault.")
	}
	return client
}
