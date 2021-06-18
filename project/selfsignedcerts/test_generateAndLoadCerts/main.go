/* 
  
Copyright (c) 2009 The Go Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
---------------------
MIT License

Copyright (c) 2016 Jacob Hoffman-Andrews

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
-----------------------------------------------------

MIT License

Copyright (c) 2020 Shane Utt

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
-----------------
https://medium.com/@Raulgzm/export-import-pem-files-in-go-67614624adc7
*/
package main

import (
	"bytes"
//	"os"
    "crypto"
	"crypto/rand"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"encoding/asn1"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

func main() {
	// get our ca and server certificate
	serverTLSConf, clientTLSConf, err := certsetup()
	if err != nil {
		panic(err)
	}

	// set up the httptest.Server using our certificate signed by our CA
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "success!")
	}))
	server.TLS = serverTLSConf
	server.StartTLS()
	defer server.Close()

	// communicate with the server using an http.Client configured to trust our CA
	transport := &http.Transport{
		TLSClientConfig: clientTLSConf,
	}
	http := http.Client{
		Transport: transport,
	}
	resp, err := http.Get(server.URL)
	if err != nil {
		panic(err)
	}

	// verify the response
	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	body := strings.TrimSpace(string(respBodyBytes[:]))
	if body == "success!" {
		fmt.Println(body)
	} else {
		panic("not successful!")
	}
}

func calculateSKID(pubKey crypto.PublicKey) ([]byte, error) {
	spkiASN1, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return nil, err
	}

	var spki struct {
		Algorithm        pkix.AlgorithmIdentifier
		SubjectPublicKey asn1.BitString
	}
	_, err = asn1.Unmarshal(spkiASN1, &spki)
	if err != nil {
		return nil, err
	}
	skid := sha1.Sum(spki.SubjectPublicKey.Bytes)
	return skid[:], nil
}

func certsetup() (serverTLSConf *tls.Config, clientTLSConf *tls.Config, err error) {

	// create our private and public key
	rootCaPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return
	}
	
	rootCaPrivKeyBytes, err := x509.MarshalECPrivateKey(rootCaPrivKey)
	
	rootCaskid, err := calculateSKID(&rootCaPrivKey.PublicKey);

	// set up our root CA certificate
	rootCa := &x509.Certificate{
	    Version: 1,
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Country:       []string{"ID"},
			Organization:  []string{"whereistimbo's Self Signed CA"},
			OrganizationalUnit: []string{"whereistimbo's Root CA"},
			CommonName: "whereistimbo's Self Signed Root CA",
			Province:      []string{"JB"},
			Locality:      []string{"Depok"},
			StreetAddress: []string{"Kukusan Beji"},
			PostalCode:    []string{"16425"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 1, 0),
		IsCA:                  true,
		SubjectKeyId: rootCaskid,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign |
		x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	// create the CA
	rootCaBytes, err := x509.CreateCertificate(rand.Reader, rootCa, rootCa, &rootCaPrivKey.PublicKey, rootCaPrivKey)
	if err != nil {
		return
	}

	// pem encode
	rootCaPEM := new(bytes.Buffer)
	pem.Encode(rootCaPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: rootCaBytes,
	})

	rootCaPrivKeyPEMBytesBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: rootCaPrivKeyBytes,
	}
	rootCaPrivKeyPEMBytes := new(bytes.Buffer)
	pem.Encode(rootCaPrivKeyPEMBytes, rootCaPrivKeyPEMBytesBlock)

	// create our private and public key
	intCaPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return
	}
	
	intCaPrivKeyBytes, err := x509.MarshalECPrivateKey(intCaPrivKey)
	
	intCaskid, err := calculateSKID(&intCaPrivKey.PublicKey);
	
	// set up our intermediate CA certificate
	intCa := &x509.Certificate{
	    Version: 1,
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Country:       []string{"ID"},
			Organization:  []string{"whereistimbo's Self Signed CA"},
			OrganizationalUnit: []string{"whereistimbo's Intermediate CA"},
			CommonName: "whereistimbo's Intermediate CA",
			Province:      []string{"JB"},
			Locality:      []string{"Depok"},
			StreetAddress: []string{"Kukusan Beji"},
			PostalCode:    []string{"16425"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 20),
		IsCA:                  true,
		SubjectKeyId: intCaskid,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign |
		x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	// create the CA
	intCaBytes, err := x509.CreateCertificate(rand.Reader, intCa, rootCa, &intCaPrivKey.PublicKey, rootCaPrivKey)
	if err != nil {
		return
	}

	// pem encode
	intCaPEM := new(bytes.Buffer)
	pem.Encode(intCaPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: intCaBytes,
	})

	intCaPrivKeyPEMBytesBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: intCaPrivKeyBytes,
	}
	intCaPrivKeyPEMBytes := new(bytes.Buffer)
	pem.Encode(intCaPrivKeyPEMBytes, intCaPrivKeyPEMBytesBlock)
    
	crl := &x509.RevocationList{
		RevokedCertificates: nil,
		Number: big.NewInt(3),
		ThisUpdate: time.Now(),
		NextUpdate: time.Now().AddDate(0, 1, 0),
	}
	
	crlBytes, err := x509.CreateRevocationList(rand.Reader, crl, intCa, intCaPrivKey)
	if err != nil {
		return
	}
	
	crlPEM := new(bytes.Buffer)
	pem.Encode(crlPEM, &pem.Block{
		Type:  "X509 CRL",
		Bytes: crlBytes,
	})
	
	
	// set up ocsp certificate
	/*
	ocspCertPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return
	}
	
	ocspCertPrivKeyBytes, err := x509.MarshalECPrivateKey(ocspCertPrivKey)
	
	ocspskid, err := calculateSKID(&ocspCertPrivKey.PublicKey);
	
	ocspCert := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Country:       []string{"ID"},
			Organization:  []string{"whereistimbo's Self Signed Certs"},
			OrganizationalUnit: []string{"whereistimbo's CRL"},
			CommonName: "whereistimbo's CRL",
			Province:      []string{"JB"},
			Locality:      []string{"Depok"},
			StreetAddress: []string{"Kukusan Beji"},
			PostalCode:    []string{"16425"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 6, 0),
		IsCA:         false,
		SubjectKeyId: ocspskid,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageOCSPSigning},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	ocspCertBytes, err := x509.CreateCertificate(rand.Reader, ocspCert, intCa, &ocspCertPrivKey.PublicKey, intCaPrivKey)
	if err != nil {
		return
	}

	ocspCertPEM := new(bytes.Buffer)
	pem.Encode(ocspCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ocspCertBytes,
	})

	ocspCertPrivKeyPEMBytesBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: ocspCertPrivKeyBytes,
	}
	ocspCertPrivKeyPEMBytes := new(bytes.Buffer)
	pem.Encode(ocspCertPrivKeyPEMBytes, ocspCertPrivKeyPEMBytesBlock)

	ocspCertPair, err := tls.X509KeyPair(ocspCertPEM.Bytes(), ocspCertPrivKeyPEMBytes.Bytes())
	if err != nil {
		return
	}
    */
	
	//set up server cert
	serverCertPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return
	}
	
	serverCertPrivKeyBytes, err := x509.MarshalECPrivateKey(serverCertPrivKey)
	
	serverskid, err := calculateSKID(&serverCertPrivKey.PublicKey);
	
	serverCert := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Country:       []string{"ID"},
			Organization:  []string{"whereistimbo's Self Signed Certs"},
			OrganizationalUnit: []string{"whereistimbo's CRL"},
			CommonName: "whereistimbo's CRL",
			Province:      []string{"JB"},
			Locality:      []string{"Depok"},
			StreetAddress: []string{"Kukusan Beji"},
			PostalCode:    []string{"16425"},
		},
		IPAddresses:  []net.IP{net.IPv4(172, 19, 11, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 6, 0),
		IsCA:         false,
		SubjectKeyId: serverskid,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		//OCSPServer:		[]string{"http://ocsp-crl-crt.whereistimbo.local/int/ocsp/"},
		//IssuingCertificateURL: []string{"http:/ocsp-crl-crt.whereistimbo.local/int/int.whereistimbo.crt"},
		//CRLDistributionPoints: []string{"http:/ocsp-crl-crt.whereistimbo.local/int/int.whereistimbo.crl"},
	}

	serverCertBytes, err := x509.CreateCertificate(rand.Reader, serverCert, intCa, &serverCertPrivKey.PublicKey, intCaPrivKey)
	if err != nil {
		return
	}

	serverCertPEM := new(bytes.Buffer)
	pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})

	serverCertPrivKeyPEMBytesBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: serverCertPrivKeyBytes,
	}
	serverCertPrivKeyPEMBytes := new(bytes.Buffer)
	pem.Encode(serverCertPrivKeyPEMBytes, serverCertPrivKeyPEMBytesBlock)

	serverCertPair, err := tls.X509KeyPair(serverCertPEM.Bytes(), serverCertPrivKeyPEMBytes.Bytes())
	if err != nil {
		return
	}

	serverTLSConf = &tls.Config{
		Certificates: []tls.Certificate{serverCertPair},
	}

	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(rootCaPEM.Bytes())
	certpool.AppendCertsFromPEM(intCaPEM.Bytes())
	
	clientTLSConf = &tls.Config{
		RootCAs: certpool,
	}

	return
}