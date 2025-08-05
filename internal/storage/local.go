package storage

import (
	"io"
	"os"
	"path/filepath"
)

type Provider interface {
	Store(filename string, data io.Reader) error
}

type Local struct {
	basePath string
}

func NewLocal(basePath string) *Local {
	return &Local{basePath: basePath}
}

func (l *Local) Store(filename string, data io.Reader) error {
	err := os.MkdirAll(l.basePath, 0755)
	if err != nil {
		return err
	}

	filePath := filepath.Join(l.basePath, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	return err
}
