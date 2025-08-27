package injector

import (
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
)

type Rule struct {
	Ref         string
	Location    string
	NewRef      string
	NewLocation string
}

func NewRule(ref string, location string) (Rule, error) {
	hash, err := calculateHash(location)
	if err != nil {
		return Rule{}, err
	}
	return Rule{Ref: ref, Location: location,
		NewRef: modifyName(ref, hash), NewLocation: modifyName(location, hash)}, nil
}

func modifyName(file string, hash uint32) string {
	if file == "" {
		return ""
	}
	ext := filepath.Ext(file)
	name := filepath.Base(file)
	return filepath.Join(filepath.Dir(file), fmt.Sprintf("%s-%x%s", name[:len(name)-len(ext)], hash, ext))
}

func calculateHash(file string) (uint32, error) {
	f, err := os.Open(file)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()
	hash := crc32.NewIEEE()
	if _, err := io.Copy(hash, f); err != nil {
		return 0, err
	}
	return hash.Sum32(), nil
}
