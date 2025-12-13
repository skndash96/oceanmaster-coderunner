package manager

import (
	"fmt"
	"os"
)

func UploadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	// for now, just log the file content
	// io.Copy(os.Stdout, f)

	return nil
}
