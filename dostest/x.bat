@ECHO OFF
set NLM=^


set NL=^^^%NLM%%NLM%^%NLM%%NLM%
set openssl=C:\Users\User\Downloads\openssl\openssl.exe

REM create directory for root ca
if not exist ".\\ca" mkdir .\\ca
if not exist ".\\ca\\root" mkdir .\\ca\root
if not exist ".\\ca\\root\\certs" mkdir .\\ca\\root\\certs
if not exist ".\\ca\\root\\crl" mkdir .\\ca\\root\\crl
if not exist ".\\ca\\root\\private" mkdir .\\ca\\root\\private
type nul >".\\ca\\root\\index.txt"
echo 1000 >".\\ca\\root\\serial"

REM create openssl_root.cnf
for /f "tokens=1 delims=:" %%a in ('findstr /n /b "openssl_root.cnf" "%~f0"') do set ln=%%a

type nul > ".\\ca\\root\\openssl_root.cnf"
(for /f "usebackq skip=%ln% delims=" %%a in ("%~f0") do (
    if [%%a]==[EOF] (
      goto :continue1
    ) else (
      call echo %%a>>".\\ca\\root\\openssl_root.cnf"
    ) 
  )
) 
goto :eof
:continue1
type ".\\ca\\root\\openssl_root.cnf"

REM create root CA
%openssl% ecparam -genkey -name secp384r1 -noout -out .\\ca\\root\\private\\ca.whereistimbo.key.pem
%openssl% req -config .\\ca\\root\\openssl_root.cnf -new -x509 -sha384 -extensions v3_ca -key .\\ca\\root\\private\\ca.whereistimbo.key.pem -out .\\ca\\root\\certs\\ca.whereistimbo.crt.pem

REM create directory for intermediate ca
if not exist ".\\ca\\intermediate" mkdir .\\ca\intermediate
if not exist ".\\ca\\intermediate\\certs" mkdir .\\ca\\intermediate\\certs
if not exist ".\\ca\\intermediate\\crl" mkdir .\\ca\\intermediate\\crl
if not exist ".\\ca\\intermediate\\csr" mkdir .\\ca\\intermediate\\csr
if not exist ".\\ca\\intermediate\\private" mkdir .\\ca\\intermediate\\private
type nul >".\\ca\\intermediate\\index.txt"
echo 1000 >".\\ca\\intermediate\\serial"
echo 1000 >".\\ca\\intermediate\\crlnumber"
REM create openssl_intermediate.cnf
for /f "tokens=1 delims=:" %%a in ('findstr /n /b "openssl_intermediate.cnf" "%~f0"') do set ln=%%a
type nul > ".\\ca\\intermediate\\openssl_intermediate.cnf"
(for /f "usebackq skip=%ln% delims=" %%a in ("%~f0") do (
    if [%%a]==[EOF] (
      goto :continue2
    ) else (
      call echo %%a>>".\\ca\\intermediate\\openssl_intermediate.cnf"
    ) 
  )
) 
goto :eof
:continue2
type ".\\ca\\intermediate\\openssl_intermediate.cnf"

REM create intermediary CA signing request
%openssl% ecparam -genkey -name secp384r1 -noout -out .\\ca\\intermediate\\private\\int.whereistimbo.key.pem
%openssl% req -config .\\ca\\intermediate\\openssl_intermediate.cnf -new -sha384 -key .\\ca\\intermediate\\private\\int.whereistimbo.key.pem -out .\\ca\\intermediate\\csr\\int.whereistimbo.csr

REM sign the intermediary CA signing request
%openssl% ca -config .\\ca\\root\\openssl_root.cnf  -extensions v3_intermediate_ca -days 390 -md sha384 -in .\\ca\\intermediate\\csr\\int.whereistimbo.csr -out .\\ca\\intermediate\\certs\\int.whereistimbo.crt.pem

REM create crl
%openssl% ca -config .\\ca\\intermediate\\openssl_intermediate.cnf -gencrl -out .\\ca\\intermediate\\crl\\crlwhereistimbo.crl

REM create openssl_server.cnf
for /f "tokens=1 delims=:" %%a in ('findstr /n /b "openssl_server.cnf" "%~f0"') do set ln=%%a
type nul > ".\\ca\\intermediate\\openssl_server.cnf"
(for /f "usebackq skip=%ln% delims=" %%a in ("%~f0") do (
    if [%%a]==[EOF] (
      goto :continue3
    ) else (
      call echo %%a>>".\\ca\\intermediate\\openssl_server.cnf"
    ) 
  )
) 
goto :eof
:continue3
type ".\\ca\\intermediate\\openssl_server.cnf"
goto :eof

REM create ocsp certificate signing request
%openssl% ecparam -genkey -name secp384r1 -noout -out .\\ca\\intermediate\\private\\ocsp.whereistimbo.key.pem
%openssl% req -config .\\ca\\intermediate\\openssl_server.cnf -new -sha384  -key .\\ca\\intermediate\\private\\ocsp.whereistimbo.key.pem -out .\\ca\\intermediate\\csr\\ocsp.whereistimbo.csr.pem -extensions server_cert 

REM sign the ocsp csr using intermediary ca
%openssl% ca -config .\\ca\\intermediate\\openssl_intermediate.cnf -extensions ocsp -days 390 -notext -md sha384 -in .\\ca\\intermediate\\csr\\ocsp.whereistimbo.csr.pem -out .\\ca\\intermediate\\certs\\ocsp.whereistimbo.crt.pem
echo 'debug 1'

REM create server csr
%openssl% ecparam -genkey -name secp384r1 -noout -out .\\ca\\intermediate\\private\\whereistimbo.local.key.pem
%openssl% req -config .\\ca\\intermediate\\openssl_server.cnf -new -sha384 -key .\\ca\\intermediate\\private\\whereistimbo.local.key.pem -out .\\ca\\intermediate\\csr\\whereistimbo.local.csr
echo 'debug 2'

REM sign server CSR
%openssl% ca -config .\\ca\\intermediate\\openssl_server.cnf -extensions server_cert -days 90 -in .\\ca\\intermediate\\csr\\whereistimbo.local.csr -out .\\ca\\intermediate\\certs\\whereistimbo.local.crt.pem
echo 'debug 3'


~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
~~~~~~~~~~~~~~~~~ openssl_root.cnf ~~~~~~~~~~~~~~~~~~~~
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
openssl_root.cnf
# OpenSSL Root CA configuration file%NL%

[ ca ]
default_ca = CA_default%NL%

[ CA_default ]
# Directory and file locations.
dir               = .\\ca\\root
certs             = $dir\\certs
crl_dir           = $dir\\crl
new_certs_dir     = $dir\\certs
database          = $dir\\index.txt
serial            = $dir\\serial
RANDFILE          = $dir\\private\\.rand%NL%

# The root key and root certificate.
private_key       = $dir\\private\\ca.whereistimbo.key.pem
certificate       = $dir\\certs\\ca.whereistimbo.crt.pem%NL%

# For certificate revocation lists.
crlnumber         = $dir\\crlnumber
crl               = $dir\\crl\\intermediate.crl.pem
crl_extensions    = crl_ext
default_crl_days  = 30%NL%

# SHA-1 is deprecated, so use SHA-2 or SHA-3 instead.
default_md        = sha384%NL%

name_opt          = ca_default
cert_opt          = ca_default
default_days      = 397
preserve          = no
policy            = policy_strict%NL%

[ policy_strict ]
# The root CA should only sign intermediate certificates that match.
# See the POLICY FORMAT section of `man ca`.
countryName             = match
stateOrProvinceName     = match
organizationName        = supplied
organizationalUnitName  = optional
commonName              = supplied
emailAddress            = optional%NL%

[ policy_loose ]
# Allow the intermediate CA to sign a more diverse range of certificates.
# See the POLICY FORMAT section of the `ca` man page.
countryName             = optional
stateOrProvinceName     = optional
localityName            = optional
organizationName        = optional
organizationalUnitName  = optional
commonName              = supplied
emailAddress            = optional%NL%

[ req ]
prompt				= no
# Options for the `req` tool (`man req`).
default_bits        = 2048
distinguished_name  = req_distinguished_name
string_mask         = utf8only%NL%

# SHA-1 is deprecated, so use SHA-2 or SHA-3 instead.
default_md          = sha256%NL%

# Extension to add when the -x509 option is used.
x509_extensions     = v3_ca%NL%

[ req_distinguished_name ]
countryName             = ID
stateOrProvinceName     = JB
localityName            = Depok
0.organizationName      = whereistimbo\'s Root CA
organizationalUnitName  = whereistimbo\'s Dev and Testing Div
commonName				= whereistimbo
emailAddress            = ca@whereistimbo.local%NL%


[ v3_ca ]
# Extensions for a typical CA (`man x509v3_config`).
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true
keyUsage = critical, digitalSignature, cRLSign, keyCertSign%NL%

[ v3_intermediate_ca ]
# Extensions for a typical intermediate CA (`man x509v3_config`).
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true, pathlen:0
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
crlDistributionPoints = @crl_info
authorityInfoAccess = @ocsp_info%NL%

[ usr_cert ]
# Extensions for client certificates (`man x509v3_config`).
basicConstraints = CA:FALSE
nsCertType = client, email
nsComment = "OpenSSL Generated Client Certificate"
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer
keyUsage = critical, nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, emailProtection%NL%

[ server_cert ]
# Extensions for server certificates (`man x509v3_config`).
basicConstraints = CA:FALSE
nsCertType = server
nsComment = "OpenSSL Generated Server Certificate"
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer:always
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
crlDistributionPoints = @crl_info
authorityInfoAccess = @ocsp_info
subjectAltName = @alt_names%NL%

[alt_names]
DNS.0 = CN Name Here%NL%

[ crl_ext ]
# Extension for CRLs (`man x509v3_config`).
authorityKeyIdentifier=keyid:always%NL%

[ ocsp ]
# Extension for OCSP signing certificates (`man ocsp`).
basicConstraints = CA:FALSE
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer
keyUsage = critical, digitalSignature
extendedKeyUsage = critical, OCSPSigning%NL%

[crl_info]
URI.0 = http://crl.whereistimbo.local/crlwhereistimbo.crl%NL%

[ocsp_info]
caIssuers;URI.0 = http://ocsp.whereistimbo.local/whereistimboroot.crt
OCSP;URI.0 = http://ocsp.whereistimbo.local/%NL%
EOF

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
~~~~~~~~~~~~~ openssl_intermediate.cnf ~~~~~~~~~~~~~~~~
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
openssl_intermediate.cnf
# OpenSSL Intermediary CA configuration file%NL%

[ ca ]
default_ca = CA_default%NL%

[ CA_default ]
# Directory and file locations.
dir               = .\\ca\\intermediate
certs             = $dir\\certs
crl_dir           = $dir\\crl
new_certs_dir     = $dir\\certs
database          = $dir\\index.txt
serial            = $dir\\serial
RANDFILE          = $dir\\private\\.rand%NL%

# The root key and root certificate.
private_key       = $dir\\private\\int.whereistimbo.key.pem
certificate       = $dir\\certs\\int.whereistimbo.crt.pem%NL%

# For certificate revocation lists.
crlnumber         = $dir\\crlnumber
crl               = $dir\\crl\\crlwhereistimbo.crl
crl_extensions    = crl_ext
default_crl_days  = 30%NL%

# SHA-1 is deprecated, so use SHA-2 or SHA-3 instead.
default_md        = sha384%NL%

name_opt          = ca_default
cert_opt          = ca_default
default_days      = 390
preserve          = no
policy            = policy_loose%NL%

[ policy_loose ]
# Allow the intermediate CA to sign a more diverse range of certificates.
# See the POLICY FORMAT section of the `ca` man page.
countryName             = optional
stateOrProvinceName     = optional
localityName            = optional
organizationName        = optional
organizationalUnitName  = optional
commonName              = supplied
emailAddress            = optional%NL%

[ req ]
prompt				= no
# Options for the `req` tool (`man req`).
default_bits        = 2048
distinguished_name  = req_distinguished_name
string_mask         = utf8only%NL%

# SHA-1 is deprecated, so use SHA-2 or SHA-3 instead.
default_md          = sha256%NL%

# Extension to add when the -x509 option is used.
x509_extensions     = v3_ca%NL%

[ req_distinguished_name ]
countryName             = ID
stateOrProvinceName     = JB
localityName            = Depok
0.organizationName      = whereistimbo\'s Intermediary CA
organizationalUnitName  = whereistimbo\'s Dev and Testing Div
commonName				= whereistimbo
emailAddress            = ca@whereistimbo.local%NL%


[ v3_ca ]
# Extensions for a typical CA (`man x509v3_config`).
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true
keyUsage = critical, digitalSignature, cRLSign, keyCertSign%NL%

[ v3_intermediate_ca ]
# Extensions for a typical intermediate CA (`man x509v3_config`).
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true, pathlen:0
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
crlDistributionPoints = @crl_info
authorityInfoAccess = @ocsp_info%NL%

[ usr_cert ]
# Extensions for client certificates (`man x509v3_config`).
basicConstraints = CA:FALSE
nsCertType = client, email
nsComment = "OpenSSL Generated Client Certificate"
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer
keyUsage = critical, nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, emailProtection%NL%

[ server_cert ]
# Extensions for server certificates (`man x509v3_config`).
basicConstraints = CA:FALSE
nsCertType = server
nsComment = "OpenSSL Generated Server Certificate"
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer:always
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
crlDistributionPoints = @crl_info
authorityInfoAccess = @ocsp_info%NL%

[ crl_ext ]
# Extension for CRLs (`man x509v3_config`).
authorityKeyIdentifier=keyid:always%NL%

[ ocsp ]
# Extension for OCSP signing certificates (`man ocsp`).
basicConstraints = CA:FALSE
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer
keyUsage = critical, digitalSignature
extendedKeyUsage = critical, OCSPSigning%NL%

[crl_info]
URI.0 = http://crl.whereistimbo.local/crlwhereistimbo.crl%NL%

[ocsp_info]
caIssuers;URI.0 = http://ocsp.whereistimbo.local/whereistimboroot.crt
OCSP;URI.0 = http://ocsp.whereistimbo.local/%NL%
EOF

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
~~~~~~~~~~~~~~~~ openssl_server.cnf ~~~~~~~~~~~~~~~~~~~
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
openssl_server.cnf
# OpenSSL Server Default configuration file%NL%

[ ca ]
default_ca = CA_default%NL%

[ CA_default ]
# Directory and file locations.
dir               = .\\ca\\intermediate
certs             = $dir\\certs
crl_dir           = $dir\\crl
new_certs_dir     = $dir\\certs
database          = $dir\\index.txt
serial            = $dir\\serial
RANDFILE          = $dir\\private\\.rand%NL%

# The root key and root certificate.
private_key       = $dir\\private\\int.whereistimbo.key.pem
certificate       = $dir\\certs\\int.whereistimbo.crt.pem%NL%

# For certificate revocation lists.
crlnumber         = $dir\\crlnumber
crl               = $dir\\crl\\crlwhereistimbo.crl
crl_extensions    = crl_ext
default_crl_days  = 30%NL%

# SHA-1 is deprecated, so use SHA-2 or SHA-3 instead.
default_md        = sha384%NL%

name_opt          = ca_default
cert_opt          = ca_default
default_days      = 397
preserve          = no
policy            = policy_loose%NL%

[ policy_loose ]
# Allow the intermediate CA to sign a more diverse range of certificates.
# See the POLICY FORMAT section of the `ca` man page.
countryName             = optional
stateOrProvinceName     = optional
localityName            = optional
organizationName        = optional
organizationalUnitName  = optional
commonName              = supplied
emailAddress            = optional%NL%

[ req ]
prompt				= no
# Options for the `req` tool (`man req`).
default_bits        = 2048
distinguished_name  = req_distinguished_name
string_mask         = utf8only%NL%

# SHA-1 is deprecated, so use SHA-2 or SHA-3 instead.
default_md          = sha384%NL%

# Extension to add when the -x509 option is used.
x509_extensions     = v3_ca%NL%

[ req_distinguished_name ]
countryName             = ID
stateOrProvinceName     = JB
localityName            = Depok
0.organizationName      = whereistimbo\'s Local Net and Server
organizationalUnitName  = whereistimbo\'s Dev and Testing Div
commonName				= whereistimbo.local
emailAddress            = ca@whereistimbo.local%NL%


[ v3_ca ]
# Extensions for a typical CA (`man x509v3_config`).
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true
keyUsage = critical, digitalSignature, cRLSign, keyCertSign%NL%

[ v3_intermediate_ca ]
# Extensions for a typical intermediate CA (`man x509v3_config`).
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:true, pathlen:0
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
crlDistributionPoints = @crl_info
authorityInfoAccess = @ocsp_info%NL%

[ usr_cert ]
# Extensions for client certificates (`man x509v3_config`).
basicConstraints = CA:FALSE
nsCertType = client, email
nsComment = "OpenSSL Generated Client Certificate"
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer
keyUsage = critical, nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, emailProtection%NL%

[ server_cert ]
# Extensions for server certificates (`man x509v3_config`).
basicConstraints = CA:FALSE
nsCertType = server
nsComment = "Grilled Cheese Generated Server Certificate"
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer:always
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
crlDistributionPoints = @crl_info
authorityInfoAccess = @ocsp_info
subjectAltName = @alt_names%NL%

[alt_names]
DNS.0 = whereistimbo.local

[ crl_ext ]
# Extension for CRLs (`man x509v3_config`).
authorityKeyIdentifier=keyid:always%NL%

[ ocsp ]
# Extension for OCSP signing certificates (`man ocsp`).
basicConstraints = CA:FALSE
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer
keyUsage = critical, digitalSignature
extendedKeyUsage = critical, OCSPSigning%NL%

[crl_info]
URI.0 = http://crl.whereistimbo.local/crlwhereistimbo.crl%NL%

[ocsp_info]
caIssuers;URI.0 = http://ocsp.whereistimbo.local/whereistimboroot.crt
OCSP;URI.0 = http://ocsp.whereistimbo.local/%NL%
EOF