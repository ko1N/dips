package taskfs

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/ko1N/dips/pkg/taskstorage"
)

type DiskFS struct {
	// the full path of the root folder of this filesystem
	path string

	inputs  []*taskstorage.FileUrl
	outputs []*taskstorage.FileUrl
}

func (self *DiskFS) AddInput(input *taskstorage.FileUrl) error {
	for _, in := range self.inputs {
		if in.FilePath == input.FilePath {
			return fmt.Errorf("input file already added\n")
		}
	}

	self.inputs = append(self.inputs, input)

	// connect to storage
	storage, err := taskstorage.ConnectStorage(input)
	if err != nil {
		return fmt.Errorf("Unable to connect to storage: %s", err)
	}

	// get reader
	reader, err := storage.GetFileReader(input)
	if err != nil {
		return fmt.Errorf("Failed to get reader from storage: %s", err)
	}
	defer reader.Close()

	// convert path
	diskPath, err := self.ToFullPath(input)
	if err != nil {
		return fmt.Errorf("Failed to create full path for file '%s': %s", input.URL.String(), err)
	}

	// create containing folder
	diskDir, _ := filepath.Split(diskPath)
	err = os.MkdirAll(diskDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Failed to create directory '%s': %s", diskDir, err)
	}

	// write file to disk
	writer, err := os.Create(diskPath)
	if err != nil {
		return fmt.Errorf("Failed to create file '%s': %s", diskPath, err)
	}
	defer writer.Close()

	_, err = io.Copy(writer, reader)
	return err
}

func (self *DiskFS) AddOutput(output *taskstorage.FileUrl) error {
	for _, out := range self.outputs {
		if out.FilePath == output.FilePath {
			return fmt.Errorf("output file already added\n")
		}
	}

	// for outputs we won't upload the files unless `Close` is called.
	self.outputs = append(self.outputs, output)

	// convert path
	diskPath, err := self.ToFullPath(output)
	if err != nil {
		return fmt.Errorf("Failed to create full path for file '%s': %s", output.URL.String(), err)
	}

	// create containing folder
	diskDir, _ := filepath.Split(diskPath)
	err = os.MkdirAll(diskDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Failed to create directory '%s': %s", diskDir, err)
	}

	return nil
}

func (self *DiskFS) RootPath() string {
	return self.path
}

func (self *DiskFS) ToFullPath(file *taskstorage.FileUrl) (string, error) {
	// path inside of root
	fullpath := path.Join(self.path, file.FilePath)
	containsPath, _ := IsSubPath(self.path, fullpath)
	if !containsPath {
		return "", fmt.Errorf("can't access file outside of sandbox")
	}
	return fullpath, nil
}

func (self *DiskFS) Flush() error {
	for _, output := range self.outputs {
		// convert path
		diskPath, err := self.ToFullPath(output)
		if err != nil {
			return fmt.Errorf("Failed to create full path for file '%s': %s", output.URL.String(), err)
		}

		if _, err := os.Stat(diskPath); err == nil {
			fmt.Printf("Flushing file '%s'\n", output.FilePath)

			// connect to storage
			storage, err := taskstorage.ConnectStorage(output)
			if err != nil {
				return fmt.Errorf("Unable to connect to storage: %s", err)
			}

			// get reader
			reader, err := os.Open(diskPath)
			if err != nil {
				return fmt.Errorf("Failed to get reader for file '%s': %s", diskPath, err)
			}

			// write to storage
			writer, err := storage.GetFileWriter(output)
			if err != nil {
				return fmt.Errorf("Failed to get writer for storage: %s", err)
			}
			defer writer.Close()

			_, err = io.Copy(writer, reader)
			if err != nil {
				return fmt.Errorf("Failed to write to storage: %s", err)
			}
		}
	}

	return nil
}

func (self *DiskFS) Close() error {
	if self.path != "" {
		os.RemoveAll(self.path)
	}
	return nil
}

func CreateDiskFS() (*DiskFS, error) {
	// create temp folder
	// TODO: make temp storage configurable
	tempFolder := path.Join(".", "temp", uuid.New().String())
	err := os.MkdirAll(tempFolder, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &DiskFS{
		path: tempFolder,
	}, nil
}
