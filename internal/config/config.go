package config

import (
	"bytes"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
)

type FileData struct {
	Ref      string `yaml:"ref"`
	Location string `yaml:"location"`
}

type Data struct {
	Dirs  []string   `yaml:"dirs"`
	Mask  []string   `yaml:"mask"`
	Files []FileData `yaml:"files"`
}

type writeableSet interface {
	Len() int
	Item(n int) string
}

type stringSet []string
type stringerSet[G fmt.Stringer] []G

func (d FileData) String() string {
	return fmt.Sprintf("file[ref='%s',location='%s']", d.Ref, d.Location)
}

func (d *Data) String() string {
	buffer := bytes.Buffer{}
	if err := d.writeData(&buffer); err != nil {
		panic(err)
	}
	return buffer.String()
}

func (d *Data) writeData(writer io.Writer) error {
	if _, err := io.WriteString(writer, "config.data["); err != nil {
		return err
	}
	if err := d.writeDirs(writer); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, ","); err != nil {
		return err
	}
	if err := d.writeFiles(writer); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, "]"); err != nil {
		return err
	}
	return nil
}

func (d *Data) writeDirs(writer io.Writer) error {
	if _, err := io.WriteString(writer, "dirs["); err != nil {
		return err
	}
	if err := writeList(writer, stringSet(d.Dirs)); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, "]"); err != nil {
		return err
	}
	return nil
}

func (d *Data) writeFiles(writer io.Writer) error {
	if _, err := io.WriteString(writer, "files["); err != nil {
		return err
	}
	if err := writeList(writer, stringerSet[FileData](d.Files)); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, "]"); err != nil {
		return err
	}
	return nil
}

func (s stringSet) Len() int {
	return len(s)
}

func (s stringSet) Item(n int) string {
	return s[n]
}

func (s stringerSet[G]) Len() int {
	return len(s)
}

func (s stringerSet[G]) Item(n int) string {
	return s[n].String()
}

func New(name string) (*Data, error) {
	f, err := os.Open(filepath.Clean(name))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	result := Data{}
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}
	if err := isValid(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func isValid(data *Data) error {
	if err := isDirsValid(data.Dirs); err != nil {
		return err
	}
	if err := isMaskValid(data.Mask); err != nil {
		return err
	}
	if err := isFilesValid(data.Files); err != nil {
		return err
	}
	return nil
}

func isMaskValid(mask []string) error {
	for _, m := range mask {
		if m == "" {
			return errors.New("mask is empty")
		}
	}
	return nil
}

func isDirsValid(dirs []string) error {
	for _, d := range dirs {
		st, err := os.Stat(d)
		if err != nil {
			return err
		}
		if !st.IsDir() {
			return fmt.Errorf("%s is not a directory", d)
		}
	}
	return nil
}

func isFilesValid(files []FileData) error {
	for _, f := range files {
		st, err := os.Stat(filepath.Clean(f.Location))
		if err != nil {
			return err
		}
		if !st.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file", f.Location)
		}
	}
	return nil
}

func writeList(writer io.Writer, set writeableSet) error {
	for i := 0; i < set.Len(); i++ {
		if _, err := io.WriteString(writer, set.Item(i)); err != nil {
			return err
		}
		if i == set.Len()-1 {
			break
		}
		if _, err := io.WriteString(writer, ","); err != nil {
			return err
		}
	}
	return nil
}
