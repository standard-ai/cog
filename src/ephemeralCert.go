package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	log "github.com/sirupsen/logrus"
)

type ephemeralCA struct {
	Template *x509.Certificate
	Key      *rsa.PrivateKey
	CertPEM  []byte
}

type ephemeralCert struct {
	CertPEM []byte
	KeyPEM  []byte
}

const rsaKeyBits = 2048

func pemEncodeCert(certBytes []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
}

func pemEncodeRSAKey(key *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

func caSubject() pkix.Name {
	return pkix.Name{
		Organization: []string{"Cog CA"},
		Country:      []string{"US"},
		Province:     []string{"California"},
		Locality:     []string{"San Francisco"},
	}
}

func certSubject() pkix.Name {
	return pkix.Name{
		Organization: []string{"Cog"},
		Country:      []string{"US"},
		Province:     []string{"California"},
		Locality:     []string{"San Francisco"},
	}
}

func newEphemeralCA() *ephemeralCA {
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(0xfeed),
		Subject:               caSubject(),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	key, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	if err != nil {
		log.Fatal("failed to generate CA key: ", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		log.Fatal("Failed to create CA certificate: ", err)
	}

	return &ephemeralCA{
		Template: template,
		Key:      key,
		CertPEM:  pemEncodeCert(certBytes),
	}
}

func newEphemeralCert(hostname string, ca *ephemeralCA) *ephemeralCert {
	template := &x509.Certificate{
		SerialNumber: big.NewInt(0xbee),
		Subject:      certSubject(),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		SubjectKeyId: []byte{1, 2, 3, 4, 5}, // luggage
		DNSNames:     []string{hostname},
	}

	key, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	if err != nil {
		log.Fatal("failed to generate server certificate key: ", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, ca.Template, &key.PublicKey, ca.Key)
	if err != nil {
		log.Fatal("Failed to create server certificate: ", err)
	}

	return &ephemeralCert{
		CertPEM: pemEncodeCert(certBytes),
		KeyPEM:  pemEncodeRSAKey(key),
	}
}
