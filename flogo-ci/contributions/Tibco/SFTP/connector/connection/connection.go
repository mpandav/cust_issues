package connection

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var logCache = log.ChildLogger(log.RootLogger(), "SFTP-connection")
var factory = &SftpFactory{}

func init() {
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

// Settings corresponds to connection.json settings
type Settings struct {
	Name               string `md:"name,required"`
	Host               string `md:"host,required"`
	Port               int    `md:"port,required"`
	User               string `md:"user,required"`
	Password           string `md:"password,required"`
	RetryCount         int    `md:"retryCount,required"`
	RetryInterval      int    `md:"retryInterval,required"`
	PublicKeyAuth      bool   `md:"publicKeyFlag,required"`
	PrivateKey         string `md:"privateKey,required"`
	PrivateKeyPassword string `md:"privateKeyPassword,required"`
	HostKeyCheck       bool   `md:"hostKeyFlag,required"`
	KnownHostFile      string `md:"knownHostFile,required"`
}

// SftpFactory structure
type SftpFactory struct {
}

// Type method of connection.ManagerFactory must be implemented by SftpFactory
func (*SftpFactory) Type() string {
	return "SFTP"
}

func (s *Settings) Validate() error {
	if s.Host == "" {
		return errors.New("required parameter 'Host' not specified")
	}

	if s.Port < 1 {
		return errors.New("required parameter 'Port' not specified")
	}

	if s.User == "" {
		return errors.New("required parameter 'User' not specified")
	}

	if !s.PublicKeyAuth && s.Password == "" {
		return errors.New("required parameter 'Password' not specified")
	}

	if s.PublicKeyAuth && s.PrivateKey == "" {
		return errors.New("required parameter 'Private Key' not specified")
	}

	if s.HostKeyCheck && s.KnownHostFile == "" {
		return errors.New("required parameter 'Known Host File' not specified")
	}

	if s.RetryCount < 0 {
		return errors.New("parameter 'Connection Retry Count' cannot be negative")
	}

	if s.RetryInterval < 0 {
		return errors.New("parameter 'Connection Retry Interval' cannot be negative")
	}
	return nil
}

func (sharedConn *SftpSharedConfigManager) Connect(s *Settings) error {
	//2. Get ssh client config
	var config ssh.ClientConfig
	if s.PublicKeyAuth {
		pemContentBytes, err := decodeFileSelectorContent(s.PrivateKey, "Private Key")
		if err != nil {
			return fmt.Errorf("error while decoding private key: %s", err.Error())
		}

		var signer ssh.Signer
		if s.PrivateKeyPassword != "" {
			// Parse private key with passphrash
			signer, err = ssh.ParsePrivateKeyWithPassphrase(pemContentBytes, []byte(s.PrivateKeyPassword))
		} else {
			signer, err = ssh.ParsePrivateKey(pemContentBytes)
		}
		if err != nil {
			return fmt.Errorf("ssh parse private key failed: %s", err.Error())
		}

		if s.HostKeyCheck {
			knownhostFileNames := filepath.Join("sftp", s.Name) // create a temp file with connection name under sftp folder
			err := createTempFile(s.KnownHostFile, knownhostFileNames)
			if err != nil {
				return fmt.Errorf("error in creating temp host file : %s", err.Error())
			}
			hostKeyCallback, err := knownhosts.New(knownhostFileNames)
			if err != nil {
				return fmt.Errorf("failed to create host key callback: %s", err.Error())
			}

			config = ssh.ClientConfig{
				User: s.User,
				Auth: []ssh.AuthMethod{
					ssh.PublicKeys(signer),
				},
				HostKeyCallback: hostKeyCallback,
			}
			logCache.Infof("Connecting using Public Key Authentication with strict HostKey check.")
		} else {
			config = ssh.ClientConfig{
				User: s.User,
				Auth: []ssh.AuthMethod{
					ssh.PublicKeys(signer),
				},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}
			logCache.Infof("Connecting using Public Key Authentication without strict HostKey check.")
		}
	} else {
		if s.HostKeyCheck {
			knownhostFileNames := filepath.Join("sftp", s.Name) // create a temp file with connection name under sftp folder
			err := createTempFile(s.KnownHostFile, knownhostFileNames)
			if err != nil {
				return fmt.Errorf("error in creating temp host file : %s", err.Error())
			}
			hostKeyCallback, err := knownhosts.New(knownhostFileNames)
			if err != nil {
				return fmt.Errorf("failed to create host key callback: %s", err.Error())
			}

			config = ssh.ClientConfig{
				User: s.User,
				Auth: []ssh.AuthMethod{
					ssh.Password(s.Password),
				},
				HostKeyCallback: hostKeyCallback,
			}
			logCache.Infof("Connecting using User and Password with strict HostKey check.")
		} else {
			config = ssh.ClientConfig{
				User: s.User,
				Auth: []ssh.AuthMethod{
					ssh.Password(s.Password),
				},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}
			logCache.Infof("Connecting using User and Password without strict HostKey check.")
		}
	}

	//3. form the host:port string
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	//4. Connect to server
	conn, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		return fmt.Errorf("failed to dial: %s", err.Error())
	}
	//defer conn.Close()

	// Create an SFTP client on top of the SSH connection
	//logCache.Infof("Creating new SFTP client")
	client, err := sftp.NewClient(conn)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %s", err.Error())
	}
	//defer client.Close()

	sharedConn.connName = s.Name
	//sharedConn.Settings = s
	sharedConn.client = client
	sharedConn.conn = conn

	return nil
}

// NewManager method of connection.ManagerFactory must be implemented by SftpFactory
func (*SftpFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	sharedConn := &SftpSharedConfigManager{}
	var err error
	s := &Settings{}

	err = metadata.MapToStruct(settings, s, false)
	if err != nil {
		return nil, err
	}
	//1. Validate connection
	err = s.Validate()
	if err != nil {
		return nil, fmt.Errorf("sftp connection validation error: %s", err.Error())
	}

	if s.RetryCount == 0 {
		logCache.Debug("Maximum connection retry count is 0, no retry attempts will be made")
	}

	sharedConn.Settings = s

	err = sharedConn.Reconnect()
	if err != nil {
		return nil, err
	}

	return sharedConn, nil
}

// SftpSharedConfigManager structure
type SftpSharedConfigManager struct {
	connName string
	Settings *Settings
	client   *sftp.Client
	conn     *ssh.Client
}

// Type method of connection.Manager must be implemented by SftpSharedConfigManager
func (s *SftpSharedConfigManager) Type() string {
	return "SFTP"
}

// GetConnection method of connection.Manager must be implemented by SftpSharedConfigManager
func (s *SftpSharedConfigManager) GetConnection() interface{} {
	return s.client
}

// ReleaseConnection method of connection.Manager must be implemented by SftpSharedConfigManager
func (o *SftpSharedConfigManager) ReleaseConnection(connection interface{}) {
}

// GetSharedConfiguration returns connection.Manager based on connection selected
func GetSharedConfiguration(conn interface{}) (connection.Manager, error) {
	cManager, err := coerce.ToConnection(conn)
	if err != nil {
		return nil, err
	}
	return cManager, nil
}

// Start method would have business logic to start the shared resource. Since db is already initialized returning nil.
func (s *SftpSharedConfigManager) Start() error {
	return nil
}

// Stop method would do business logic to stop the the shared resource. Closing db connection in this method.
func (s *SftpSharedConfigManager) Stop() error {
	var errMsg string
	logCache.Infof("Closing SFTP client..")
	if s.client != nil {
		err := s.client.Close()
		if err != nil {
			errMsg = errMsg + fmt.Sprintf("Error closing SFTP client : %s", err.Error())
		} else {
			logCache.Infof("Closing SFTP client Successful!")
		}
	}

	logCache.Infof("Closing SSH client..")
	if s.conn != nil {
		err := s.conn.Close()
		if err != nil {
			errMsg = errMsg + fmt.Sprintf("Error closing SSH client : %s", err.Error())
		} else {
			logCache.Infof("Closing SSH client Successful!")
		}
	}

	//delete the on the fly created temp file as well as temp directory
	os.RemoveAll("sftp")

	if errMsg != "" {
		logCache.Infof(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

// Reconnect method will try to connect the SFTP server,
// if it fails then retrying the connection based on the retry count
// and interval configured using the exponential backoff retry mechanism.
func (s *SftpSharedConfigManager) Reconnect() error {
	var err error
	connected := false
	if err = s.Connect(s.Settings); err == nil {
		connected = true
		logCache.Infof("Connection Successful. Connection name : %s", s.Settings.Name)
	} else if s.Settings.RetryCount > 0 {
		// Retry logic using the exponential backoff retry mechanism
		for i := 0; i < s.Settings.RetryCount; i++ {
			delay := s.Settings.RetryInterval * (1 << i) // 20s, 40s, 80s, etc.
			logCache.Infof("Connection failed. Retrying in %v seconds", delay)
			time.Sleep(time.Duration(delay) * time.Second)

			logCache.Infof("Reconnect attempt %d of %d...", i+1, s.Settings.RetryCount)
			if err = s.Connect(s.Settings); err == nil {
				connected = true
				logCache.Infof("Connected successfully after %d retry attempt. Connection name : %s", i+1, s.Settings.Name)
				break
			}
		}
	}

	if !connected {
		return fmt.Errorf("could not connect to SFTP server: %s", err.Error())
	}

	return nil
}

func decodeFileSelectorContent(fieldVal string, field string) ([]byte, error) {
	if fieldVal == "" {
		return nil, fmt.Errorf("field '%s' is not configured", field)
	}

	//if input comes from fileselctor it will be base64 encoded
	if strings.HasPrefix(fieldVal, "{") {
		fieldObj, err := coerce.ToObject(fieldVal)
		if err == nil {
			fieldContent, ok := fieldObj["content"].(string)
			if !ok || fieldContent == "" {
				return nil, fmt.Errorf("invalid value of field '%s'", field)
			}

			index := strings.IndexAny(fieldContent, ",")
			if index > -1 {
				fieldContent = fieldContent[index+1:]
			}

			decodedVal, err := base64.StdEncoding.DecodeString(fieldContent)
			if err != nil {
				return nil, fmt.Errorf("invalid base64 encoded value of field '%s'", field)
			}
			return []byte(decodedVal), nil
		}
		return nil, err
	}

	decodedVal, err := base64.StdEncoding.DecodeString(fieldVal)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 encoded value of field '%s'. Check override value configured to the application property", field)
	}
	return []byte(decodedVal), nil
}

func createTempFile(knownHostFile string, knownhostFileNames string) error {
	//first create temp directory if it does not exist
	tempDir := "sftp"
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		err = os.Mkdir(tempDir, 0700)
		if err != nil {
			return fmt.Errorf("error in creating temp directory : %s", err.Error())
		}
	}

	f, err := os.Create(knownhostFileNames)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %s", err.Error())
	}
	defer f.Close()

	dec, err := decodeFileSelectorContent(knownHostFile, "Known Host File")
	if err != nil {
		return fmt.Errorf("error while decoding field 'Known Host File' : %s", err.Error())
	}
	f.Write(dec)
	return nil
}
