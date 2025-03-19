package main

import "os"

type FileManager interface {
	read() []byte
}

type TextManager struct {
	filepath string
}

func (t TextManager) read() []byte {
	content, err := os.ReadFile(t.filepath)

	if err != nil {
		panic("file not found")
	}

	return content
}
