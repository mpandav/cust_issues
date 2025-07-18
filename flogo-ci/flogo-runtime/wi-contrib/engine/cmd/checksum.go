package cmd

import (
	"flag"
	"fmt"
	"github.com/tibco/wi-contrib/engine"
	"os"
	"path/filepath"
)

func init() {
	Registry("-checkSum", &checkSum{})
	Registry("--checkSum", &checkSum{})
}

type checkSum struct {
	printMetadata bool
	writeToFile   bool
}

func (b *checkSum) Run(args []string, appJson string) error {

	outputDir := os.Getenv("OUTPUT_DIR")
	if outputDir == "" {
		outputDir, _ = os.Getwd()
	}
	checksum := engine.GetSharedData("checkSum")

	checkSumStr := checksum.(string)

	if !b.writeToFile {
		fmt.Println(checkSumStr)
	}
	if b.printMetadata {
		buildMetadata := engine.GetSharedData("buildMetadata")
		fmt.Println(buildMetadata)
	}

	if b.writeToFile {
		file, err := os.Create(filepath.Join(outputDir, ".checksum"))
		if err != nil {
			panic("Failed to create file: " + err.Error())
		}
		defer file.Close()

		_, err = file.WriteString(checkSumStr)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func (b *checkSum) AddFlags(fs *flag.FlagSet) {
	fs.BoolVar(&b.printMetadata, "o", false, "Optional, Properties output json file option")

	fs.BoolVar(&b.writeToFile, "w", false, "Write checksum to file")

}

func (b *checkSum) Name() string {
	return "checksum"
}

func (b *checkSum) Description() string {
	return "Print the checksum"
}

func (b *checkSum) PrintUsage() {
}

func (b *checkSum) IsShimCommand() bool {
	return false
}

func (b *checkSum) Parse() bool {
	return true
}
