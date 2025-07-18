package cmd

import (
	"flag"
	"fmt"
	"github.com/project-flogo/core/engine/secret"
	"os"
)

func init() {
	Registry("encryptsecret", &encryptedCommand{})
	Registry("-encryptsecret", &encryptedCommand{})
	Registry("--encryptsecret", &encryptedCommand{})

}

type encryptedCommand struct {
	debug bool
}

func (b *encryptedCommand) Name() string {
	return "-encryptsecret"
}

func (b *encryptedCommand) Description() string {
	return "Encrypt string to secret"
}

func (b *encryptedCommand) Run(args []string, appJson string) error {
	return execEncrypting(args[0])
}

func (b *encryptedCommand) AddFlags(fs *flag.FlagSet) {
}

func (b *encryptedCommand) PrintUsage() {
	execName := os.Args[0]

	usage := "Command: \n" +
		"        " + execName + " -encryptsecret <plain value> \n" +
		"Usage:\n" +
		"        " + execName + " -encryptsecret MyPassword\n"
	fmt.Println(usage)
}

func (b *encryptedCommand) IsShimCommand() bool {
	return false
}

func (b *encryptedCommand) Parse() bool {
	return true
}

func execEncrypting(plainV string) error {
	// utlize the secret handler which set in secret.go
	encrypted, err := secret.GetSecretValueHandler().EncodeValue(plainV)
	if err != nil {
		fmt.Printf("Fail to encrypt secret due to %s \n", err.Error())
		return fmt.Errorf("Fail to encrypt secret due to %s \n", err.Error())
	}

	fmt.Printf("\n%s \n", "The encrypted value is:")
	fmt.Println(encrypted)
	return nil
}
