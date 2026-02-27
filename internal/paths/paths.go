package paths

import (
	"os"
	"path/filepath"
	prettyprint "taskflow/pkg/pretty_print"
)

func CheckDirectory(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		prettyprint.Warn("Directory not found: %s", path)
	} else {
		prettyprint.Debug("Directory found: %s", path)
	}
}

func GetWebPath() (webPath string, err error) {
	workDir, err := os.Getwd()
	if err != nil {
		return workDir, err
	}

	webPath = filepath.Join(workDir, "web")

	// Проверяем существование папок
	CheckDirectory(webPath)
	CheckDirectory(filepath.Join(webPath, "html"))
	CheckDirectory(filepath.Join(webPath, "css"))
	CheckDirectory(filepath.Join(webPath, "js"))

	return webPath, nil
}
