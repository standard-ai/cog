package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	vault "github.com/hashicorp/vault/api"
	"github.com/pete0emerson/browser"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var loginCommand = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with google",
	Long:  "Get hashicorp vault credentials via Google OpenID Connect",
	Run: func(cmd *cobra.Command, args []string) {
		if verbose > 0 {
			log.Infof("buildVersion: %s buildDate: %s buildHash: %s buildOS: %s", buildVersion, buildDate, buildHash, buildOS)
		}
		token := oidcLogin(nil)
		writeTokenStore(token)
		fmt.Println("Login successful.")
	},
}

var roleFlag string

func init() {
	loginCommand.Flags().StringVarP(&roleFlag, "role", "r", "", "Vault auth role")
}

type Oidc struct {
	Client *vault.Client
	Path   string
	Data   url.Values
}

func NewOidc(client *vault.Client) *Oidc {
	o := Oidc{
		Client: client,
		Path:   "auth/oidc/oidc/",
		Data:   make(url.Values),
	}
	return &o
}

func (o *Oidc) authUrl(data map[string]interface{}) {
	path := o.Path + "auth_url"
	if roleFlag != "" {
		data["role"] = roleFlag
	} else if v := os.Getenv("COG_VAULT_ROLE"); v != "" {
		data["role"] = v
	}
	secret, err := o.Client.Logical().Write(path, data)
	if err != nil {
		log.Fatalf("Vault error accessing %s: %v", path, err)
	}
	authUrl := secret.Data["auth_url"]
	err = browser.OpenURL(authUrl.(string))
	if err != nil {
		fmt.Println("Could not use default web browser: ", err)

		fmt.Printf("Authenticate at this url: %q\n", authUrl)

	}
	url, err := url.Parse(fmt.Sprintf("%v", authUrl))
	o.Data.Set("nonce", url.Query().Get("nonce"))
}

func (o *Oidc) callback() string {
	path := o.Path + "callback"
	secret, err := o.Client.Logical().ReadWithData(path, o.Data)
	if err != nil {
		log.Fatalf("Vault error accessing %s: %v", path, err)
	}
	token, err := secret.TokenID()
	if err != nil {
		log.Fatalf("Vault error creating token: %v", err)
	}
	return token
}

func oidcCallback(data url.Values, done chan bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data.Set("state", r.URL.Query().Get("state"))
		data.Set("code", r.URL.Query().Get("code"))
		w.Write([]byte("You may close this window"))
		done <- true
	}
}

func handleLocalhostRedirect(data url.Values) {
	done := make(chan bool, 1)

	http.HandleFunc("/oidc/callback", oidcCallback(data, done))
	srv := &http.Server{Addr: ":8250"}
	go srv.ListenAndServe()
	<-done

	srv.Close()
}

func oidcLogin(client *vault.Client) string {
	if client == nil {
		client = envClient(nil)
	}
	oidc := NewOidc(client)
	oidc.authUrl(map[string]interface{}{
		"redirect_uri": "http://localhost:8250/oidc/callback",
	})
	handleLocalhostRedirect(oidc.Data)
	return oidc.callback()
}
