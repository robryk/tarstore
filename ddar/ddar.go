package ddar

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

type Archive string

func (a Archive) Get(name string) (*os.File, error) {
	tempFile, err := ioutil.TempFile("", "ddar-extracted")
	if err != nil {
		return nil, err
	}
	if err := os.Remove(tempFile.Name()); err != nil {
		log.Printf("can't remove temporary file %s: %v", tempFile.Name(), err)
	}
	cmd := exec.Command("ddar", "x", string(a), name)
	cmd.Stdout = tempFile
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		tempFile.Close()
		return nil, err
	}
	if _, err := tempFile.Seek(0, os.SEEK_SET); err != nil {
		tempFile.Close()
		return nil, err
	}
	return tempFile, nil
}

func (a Archive) Add(name string, contents io.Reader) error {
	cmd := exec.Command("ddar", "c", string(a), "-N", name)
	cmd.Stdin = contents
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
