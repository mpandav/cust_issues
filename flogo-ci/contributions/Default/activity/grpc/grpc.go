package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/project-flogo/core/data/coerce"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
)

var activityMetadata = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

func init() {
	activity.Register(&Activity{}, New)
}

//type mashGRPCClienConn struct {
//	conn  *grpc.ClientConn
//	count int
//}
//
//type mashGRPCClinetConns struct {
//	connMap map[string]*mashGRPCClienConn
//	sync.Mutex
//}
//
//var conns = mashGRPCClinetConns{
//	connMap: make(map[string]*mashGRPCClienConn),
//}

// Activity is a GRPC activity
type Activity struct {
	settings   *Settings
	connection *grpc.ClientConn
}

// New creates a new javascript activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	settings := Settings{}
	err := metadata.MapToStruct(ctx.Settings(), &settings, true)
	if err != nil {
		return nil, err
	}

	logger := ctx.Logger()
	logger.Debugf("Setting: %b", settings)

	act := &Activity{
		settings: &settings,
	}

	//connection
	//Use round robin as default load balancing policy
	opts := []grpc.DialOption{grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`)}
	logger.Debug("enableTLS: ", settings.EnableTLS)
	if settings.EnableTLS {
		logger.Debug("RootCA: ", settings.RootCA)
		certificates, err := act.decodeClientCertificate(logger, settings.RootCA)
		if err != nil {
			return act, err
		}

		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(certificates) {
			return act, fmt.Errorf("credentials: failed to append certificates")
		}

		// mTLS
		if settings.EnableMTLS {
			logger.Debug("mTLS is enabled")
			clientCert, err := act.decodeClientCertificate(logger, settings.ClientCert)
			if err != nil {
				return act, err
			}
			clientKey, err := act.decodeClientCertificate(logger, settings.ClientKey)
			if err != nil {
				return act, err
			}
			certificate, err := tls.X509KeyPair(clientCert, clientKey)
			if err != nil {
				return nil, fmt.Errorf("failed to load client certification: %w", err)
			}
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{certificate},
				RootCAs:      cp,
			}
			opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
		} else {
			//TLS
			opts = append(opts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(cp, "")))
		}
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := getConnection(settings.HostURL, logger, opts)
	if err != nil {
		return act, err
	}

	act.connection = conn
	return act, nil
}

// Metadata return the metadata for the activity
func (a *Activity) Metadata() *activity.Metadata {
	return activityMetadata
}

// Eval executes the activity
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	logger := ctx.Logger()

	input := Input{}
	err = ctx.GetInputObject(&input)
	if err != nil {
		return false, err
	}

	//defer releaseConnection(a.settings.HostURL)

	logger.Debug("operating mode: ", a.settings.OperatingMode)

	output := Output{}
	switch a.settings.OperatingMode {
	case "grpc-to-grpc":
		err = a.gRPCTogRPCHandler(&input, &output, logger, a.connection)
		if err != nil {
			return false, err
		}
		err = ctx.SetOutputObject(&output)
		if err != nil {
			return false, err
		}
		return true, nil
	case "rest-to-grpc":
		err = a.restTogRPCHandler(&input, &output, logger, a.connection)
		if err != nil {
			return false, err
		}
		err = ctx.SetOutputObject(&output)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	logger.Error(errors.New("Invalid use of service , OperatingMode not recognised"))
	return false, errors.New("Invalid use of service , OperatingMode not recognised")
}

// getconnection returns single client connection object per hostaddress
func getConnection(hostAdds string, logger log.Logger, opts []grpc.DialOption) (*grpc.ClientConn, error) {
c, err := grpc.NewClient(hostAdds, opts...)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	return c, nil
}

func (a *Activity) decodeClientCertificate(logger log.Logger, cert string) ([]byte, error) {
	if cert == "" {
		return nil, fmt.Errorf("Certificate is Empty")
	}

	// case 1: if certificate comes from fileselctor it will be base64 encoded
	if strings.HasPrefix(cert, "{") {
		logger.Debug("Certificate received from file selector")
		certObj, err := coerce.ToObject(cert)
		if err == nil {
			certValue, ok := certObj["content"].(string)
			if !ok || certValue == "" {
				return nil, fmt.Errorf("No content found for certificate")
			}
			return base64.StdEncoding.DecodeString(strings.Split(certValue, ",")[1])
		}
		return nil, err
	}

	// case 2: if the certificate is defined as application property in the format "<encoding>,<encodedCertificateValue>"
	index := strings.IndexAny(cert, ",")
	if index > -1 {
		//some encoding is there
		logger.Debug("Certificate received from application property with encoding")
		encoding := cert[:index]
		certValue := cert[index+1:]

		if strings.EqualFold(encoding, "base64") {
			return base64.StdEncoding.DecodeString(certValue)
		}
		return nil, fmt.Errorf("Error parsing the certificate or given encoding may not be supported")
	}

	// case 3: if the certificate is defined as application property that points to a file
	if strings.HasPrefix(cert, "file://") {
		// app property pointing to a file
		logger.Debug("Certificate received from application property pointing to a file")
		fileName := cert[7:]
		return os.ReadFile(fileName)
	}

	// case 4: if certificate is defined as path to a file (in oss)
	if strings.Contains(cert, "/") || strings.Contains(cert, "\\") {
		logger.Debug("Certificate received from settings as file path")
		_, err := os.Stat(cert)
		if err != nil {
			logger.Errorf("Cannot find certificate file: %s", err.Error())
		}
		return os.ReadFile(cert)
	}

	// case 5: Attempt to decode as base64 string
	decode, err := base64.StdEncoding.DecodeString(cert)
	if err == nil {
		logger.Debug("Certificate received as base64 encoded string")
		return decode, nil
	}

	logger.Debug("Certificate received from application property without encoding")
	return []byte(cert), nil
}
