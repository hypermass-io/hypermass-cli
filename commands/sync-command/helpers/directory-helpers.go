package helpers

import (
	"log"
	"os"
	"path/filepath"
)

func GetStreamPathFromConfig(targetDirectory string) string {
	var folderPath = ""

	if len(targetDirectory) > 0 {
		// stream config wins
		folderPath = targetDirectory
	} else {
		ex, err := os.Executable()
		if err != nil {
			log.Println(err)
			log.Println("Failed to get current path, please report this message to support")
			os.Exit(1)
		}
		folderPath = filepath.Dir(ex)
	}
	return folderPath
}
