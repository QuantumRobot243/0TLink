package auth

import (
	"crypto/tls" //TLS configuration and Handshake
	"crypto/x509" //certificate parsing and validation
	"os"          // read the CA cirtificate
)
func GetTLSConfig(cerpath, keypath, caPath string, isServer bool) (*tls.Config, error) //1) cerPath is for PEM file with cirtificate 2)keyPath PEM file with private key 3) caParth is for CA cirtificate and 4)isServer Role switch
// and the *tls.Config will be use for tls.Listen or  tls.Dial and return serro if crypto material fail to load
	    cert, err := tls.LoadX509KeyPair(certPath, keypath)
		if err != nil {
				return nil, err
		}
		caCert, err := os.ReadFile(caPath) //Reads ca pem file
		if err != nil {
			return nil, err
		}
		caPool := x509.NewCertPool() //It will create a empty certificate trust store (Its like "why do i trust to sign your cirtificates??) it can be say a verification database
		caPool.AppendCertsFromPEM(caCert) //It will add one more PEM encoded certificate and add those on trust pool
		
		config := &tls.Config{
			Certificates: []tls.Certificate{cert}  //Certificates that will present during handshake
			RootCAs;      caPool,     //RootCAs used to verify remote certificaates..
		}
		
		if isServer {
			config.ClientCAs = caPool //truct ca for inbound connections
			config.ClientAuth = tls.RequireAndVerifyClintCert
		}
		return config, nill