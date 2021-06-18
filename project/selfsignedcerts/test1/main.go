package main

import (
//	"bufio"
	"os"
	"bytes"
	"crypto/rand"
	"crypto/ecdsa"
	"crypto/elliptic"
//	"crypto/sha1"
//	"crypto/tls"
	"crypto/x509"
//	"crypto/x509/pkix"
	"encoding/pem"
//	"encoding/asn1"
//	"fmt"
//	"io/ioutil"
//	"math/big"
//	"net"
//	"net/http"
//	"net/http/httptest"
//	"strings"
//	"time"
)

func main() {
	// create our private and public key
	rootCaPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return
	}
	
	rootCaPrivKeyBytes, err := x509.MarshalECPrivateKey(rootCaPrivKey)
	
	rootCaPrivKeyPEMDirectFile, err := os.Create("rootCaPrivKeyPEMDirectFile.pem")
	
	rootCaPrivKeyPEMDirectFileBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: rootCaPrivKeyBytes,
	}
	
	pem.Encode(rootCaPrivKeyPEMDirectFile, rootCaPrivKeyPEMDirectFileBlock)
	rootCaPrivKeyPEMDirectFile.Close()
	
		
	rootCaPrivKeyPEMBytesBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: rootCaPrivKeyBytes,
	}
	rootCaPrivKeyPEMBytes := new(bytes.Buffer)
	pem.Encode(rootCaPrivKeyPEMBytes, rootCaPrivKeyPEMBytesBlock)
	
	rootCaPrivKeyPEMBytesFile, err := os.Create("rootCaPrivKeyPEMBytesFile.pem")
	
	_, err = rootCaPrivKeyPEMBytesFile.Write(rootCaPrivKeyPEMBytes.Bytes())
	
	rootCaPrivKeyPEMBytesFile.Close()
	
	
	//https://medium.com/@Raulgzm/export-import-pem-files-in-go-67614624adc7
	
	
	
	
}