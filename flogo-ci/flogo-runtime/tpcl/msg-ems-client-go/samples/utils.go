package samples

import (
	"flag"
	"fmt"
	"github.com/tibco/msg-ems-client-go/tibems"
	"os"
)

var SslUsage = `

   where tls options are:

   -ssl_trusted
   -ssl_issuer
   -ssl_identity
   -ssl_private_key
   -ssl_password
   -ssl_expected_hostname
   -ssl_custom
   -ssl_no_verify_host
   -ssl_no_verify_hostname
   -ssl_ciphers
   -ssl_trace
   -ssl_debug_trace
`

func CreateSSLParams() (*tibems.SSLParams, string, error) {
	tlsFlags := flag.NewFlagSet("TLS Options", flag.ContinueOnError)
	sslTrusted := tlsFlags.String("ssl_trusted", "", "")
	sslIssuer := tlsFlags.String("ssl_issuer", "", "")
	sslIdentity := tlsFlags.String("ssl_identity", "", "")
	sslPrivateKey := tlsFlags.String("ssl_private_key", "", "")
	sslExpectedHostname := tlsFlags.String("ssl_expected_hostname", "", "")
	sslCustom := tlsFlags.Bool("ssl_custom", false, "")
	sslNoVerifyHost := tlsFlags.Bool("ssl_no_verify_host", false, "")
	sslNoVerifyHostName := tlsFlags.Bool("ssl_no_verify_hostname", false, "")
	sslCiphers := tlsFlags.String("ssl_ciphers", "", "")
	sslTrace := tlsFlags.Bool("ssl_trace", false, "")
	sslDebugTrace := tlsFlags.Bool("ssl_debug_trace", false, "")
	sslPassword := tlsFlags.String("ssl_password", "", "")
	err := tlsFlags.Parse(os.Args[1:])
	if err != nil {
		return nil, "", err
	}
	sslParams, err := tibems.CreateSSLParams()
	if err != nil {
		return nil, "", err
	}
	if *sslTrusted != "" {
		err = sslParams.AddTrustedCertFile(*sslTrusted, tibems.SslEncodingAuto)
		if err != nil {
			sslParams.Close()
			return nil, "", err
		}
	}
	if *sslIssuer != "" {
		err = sslParams.AddIssuerCertFile(*sslIssuer, tibems.SslEncodingAuto)
		if err != nil {
			sslParams.Close()
			return nil, "", err
		}
	}
	if *sslIdentity != "" {
		err = sslParams.SetIdentityFile(*sslIdentity, tibems.SslEncodingAuto)
		if err != nil {
			sslParams.Close()
			return nil, "", err
		}
	}
	if *sslPrivateKey != "" {
		err = sslParams.SetPrivateKeyFile(*sslPrivateKey, tibems.SslEncodingAuto)
		if err != nil {
			sslParams.Close()
			return nil, "", err
		}
	}
	if *sslExpectedHostname != "" {
		err = sslParams.SetExpectedHostName(*sslExpectedHostname)
		if err != nil {
			sslParams.Close()
			return nil, "", err
		}
	}
	if *sslCustom {
		err = sslParams.SetHostNameVerifier(func(connectedHostname string, expectedHostname string, certificateName string) error {
			if connectedHostname == "" {
				connectedHostname = "(null)"
			}
			if expectedHostname == "" {
				expectedHostname = "(null)"
			}
			if certificateName == "" {
				certificateName = "(null)"
			}
			fmt.Printf("CUSTOM VERIFIER:\n"+
				"    connected: [%s]\n"+
				"    expected:  [%s]\n"+
				"    certCN:    [%s]\n",
				connectedHostname,
				expectedHostname,
				certificateName)

			return nil
		})
		if err != nil {
			sslParams.Close()
			return nil, "", err
		}
	}
	if *sslNoVerifyHost {
		err = sslParams.SetVerifyHost(false)
		if err != nil {
			sslParams.Close()
			return nil, "", err
		}
	}
	if *sslNoVerifyHostName {
		err = sslParams.SetVerifyHostName(false)
		if err != nil {
			sslParams.Close()
			return nil, "", err
		}
	}
	if *sslCiphers != "" {
		err = sslParams.SetCiphers(*sslCiphers)
		if err != nil {
			sslParams.Close()
			return nil, "", err
		}
	}
	if *sslTrace {
		tibems.SetSSLTrace(*sslTrace)
	}
	if *sslDebugTrace {
		tibems.SetSSLDebugTrace(*sslDebugTrace)
	}

	return sslParams, *sslPassword, nil
}
