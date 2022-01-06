package system

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"time"
)

const (
	rsaPrivateKeyLen    = 4096
	certDurationInYears = 10
)

func createPrivateKey() *rsa.PrivateKey {
	// create our private and public ke
	key, err := rsa.GenerateKey(rand.Reader, rsaPrivateKeyLen)
	if err != nil {
		log.Panicf("unable to create RSA private key: %v", err)
	}

	return key
}

func createCertificateAuthorityCert(privateKey *rsa.PrivateKey) *x509.Certificate {
	// set up our CA certificate
	caCert := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization: []string{"Fringe Service CA"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(certDurationInYears, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, caCert, caCert, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Panicf("could not create CA certificate: %v", err)
	}

	// pem encode
	caPEM := new(bytes.Buffer)

	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		log.Panicf("could not create encore CA certificate: %v", err)
	}

	return caCert
}

func createServerCertificate(caCert *x509.Certificate, caPrivateKey *rsa.PrivateKey, certPrivateKey *rsa.PrivateKey, ips []net.IP) tls.Certificate {
	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization: []string{"Fringe Service"},
		},
		IPAddresses:  ips,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(certDurationInYears, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, &certPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		log.Panicf("unable to create server certificate: %v", err)
	}

	certPEM := new(bytes.Buffer)
	certPrivateKeyPEM := new(bytes.Buffer)

	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		log.Panicf("unable to encode server certificate PEM: %v", err)
	}

	err = pem.Encode(certPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivateKey),
	})
	if err != nil {
		log.Panicf("unable to encode server certificate PEM: %v", err)
	}

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivateKeyPEM.Bytes())
	if err != nil {
		log.Panicf("unable to create server x509 key pair: %v", err)
	}

	return serverCert
}

func TLSConfigWithSelfSignedCert(ips []net.IP) (serverTLSConf *tls.Config) {
	if len(ips) == 0 {
		return nil
	}

	caPrivateKey := createPrivateKey()
	certPrivateKey := createPrivateKey()
	caCert := createCertificateAuthorityCert(caPrivateKey)
	serverCert := createServerCertificate(caCert, caPrivateKey, certPrivateKey, ips)

	serverTLSConf = &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{serverCert},
	}

	return serverTLSConf
}
