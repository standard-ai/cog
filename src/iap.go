package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"

	jwt "github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type IAPCredentials struct {
	Token  string
	Expiry time.Time
}

func NewIAPCredentials() *IAPCredentials {
	iapCreds := &IAPCredentials{}
	iapCreds.Refresh()
	return iapCreds
}

func (iapCreds *IAPCredentials) BearerToken() string {
	if iapCreds.Expiry.Before(time.Now()) {
		iapCreds.Refresh()
	}
	return "Bearer " + iapCreds.Token
}

func (iapCreds *IAPCredentials) Refresh() {
	token := iapGenerateIdToken()

	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		log.Fatalf("Unparseable Vault IAP token: %v", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		log.Fatal("Can't convert Vault IAP token's claims to standard claims")
	}

	var expiry time.Time
	switch expVal := claims["exp"].(type) {
	case float64:
		expiry = time.Unix(int64(expVal), 0)
	case json.Number:
		v, _ := expVal.Int64()
		expiry = time.Unix(v, 0)
	}

	iapCreds.Token = token
	iapCreds.Expiry = expiry
}

func iapGenerateIdToken() string {
	serviceAccount := viper.GetString("vault_iap_service_account")
	clientID := viper.GetString("vault_iap_client_id")

	ctx := context.Background()
	c, err := credentials.NewIamCredentialsClient(ctx)
	if err != nil {
		log.Fatalf("Could not obtain IAM Credentials Client for Vault IAP: %v", err)
	}
	defer c.Close()

	req := &credentialspb.GenerateIdTokenRequest{
		Name:         serviceAccount,
		Audience:     clientID,
		IncludeEmail: true,
	}
	resp, err := c.GenerateIdToken(ctx, req)
	if err != nil {
		log.Fatalf("Could not obtain Vault IAP token: %v", err)
	}
	return resp.Token
}

func iapProxy(iapCreds *IAPCredentials) *httputil.ReverseProxy {
	targetURL := viper.GetString("vault_address")

	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Could not parse URL '%s': %v", targetURL, err)
	}

	// This is just a modified version of httputil.NewSingleHostReverseProxy
	// so we can add the Authorization header. Sadly the httputil functionality
	// is not written in an extensible way, so we have to duplicate it here
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		// add the IAP authorization header
		req.Header.Set("Authorization", iapCreds.BearerToken())
	}
	return &httputil.ReverseProxy{Director: director}
}

// This is also copied from httputil to support the above function
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
