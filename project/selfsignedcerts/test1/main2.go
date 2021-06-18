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

The MIT License (MIT)
Copyright (c) Microsoft Corporation

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and 
associated documentation files (the "Software"), to deal in the Software without restriction, 
including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, 
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial 
portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT 
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. 
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE 
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
-----------------
//https://medium.com/@Raulgzm/export-import-pem-files-in-go-67614624adc7
*/

package main

import (
//	"bufio"
	"os"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha1"
//	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"encoding/asn1"
//	"fmt"
//	"io/ioutil"
	"math/big"
//	"net"
//	"net/http"
//	"net/http/httptest"
//	"strings"
	"time"
	"encoding/base64"
	"os/exec"
)

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

func main() {

	rootCaPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return
	}
	
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
	
	rootCaFile, err := os.Create("rootCa.cer")
	_, err = rootCaFile.Write(rootCaPEM.Bytes())
	rootCaFile.Close()
	
	rootCaBase64 := base64.StdEncoding.EncodeToString(rootCaPEM.Bytes())
	//fmt.Println(rootCaBase64)
	
	cmd := exec.Command("powershell", "-command", "$rootca = [System.Security.Cryptography.X509Certificates.X509Certificate2]::new([System.Convert]::FromBase64String(\""+ rootCaBase64 +"\")); $store = [System.Security.Cryptography.X509Certificates.X509Store]::new(\"Root\",\"CurrentUser\"); $store.Open([System.Security.Cryptography.X509Certificates.OpenFlags]::ReadWrite); $store.Add($rootca)")
	_ = cmd.Run()

	// create our private and public key
	intCaPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return
	}
	
	
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
	
	intCaFile, err := os.Create("intCa.cer")
	_, err = intCaFile.Write(intCaPEM.Bytes())
	intCaFile.Close()
	//https://medium.com/@Raulgzm/export-import-pem-files-in-go-67614624adc7
	
	
	
	
}