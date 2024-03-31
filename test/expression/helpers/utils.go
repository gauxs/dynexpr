package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	mainFilePath string = "/cmd/dynexpr/main.go"
)

func GenerateExpressionBuilder(destinationDirPath string) error {
	curWorkingDir, err := os.Getwd()
	if err != nil {
		return err
	}
	curDir, err := FindGoMod(curWorkingDir)
	if err != nil {
		return err
	}

	execArgs := []string{"run", curDir + mainFilePath, curDir + destinationDirPath}
	cmd := exec.Command("go", execArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error in running command: %v and error is: %v", cmd, err)
		return err
	}

	return nil
}

// FindGoMod searches for the go.mod file by traversing upwards from the given directory.
func FindGoMod(dir string) (string, error) {
	for {
		// Check if go.mod exists in the current directory
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		// Move to the parent directory
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parentDir
	}
}
