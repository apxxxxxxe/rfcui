package db

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"strings"

	"errors"
)

func SaveInterface(t interface{}, filename string) error {
	pwd, _ := os.Getwd()

	if err := os.Mkdir(filepath.Join(pwd, "save"), 0777); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(pwd, "save", formatFilename(filename)))
	if err != nil {
		return err
	}

	defer f.Close()
	enc := gob.NewEncoder(f)

	if err := enc.Encode(t); err != nil {
		return err
	}
	return nil
}

func LoadInterface(filename string) (interface{}, error) {
	pwd, _ := os.Getwd()

	p := filepath.Join(pwd, "save", formatFilename(filename))
	if !fileExists(p) {
		return nil, errors.New("file is not exist: " + p)
	}

	f, err := os.Open(filepath.Join(p))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var t interface{}
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&t); err != nil {
		return nil, err
	}
	return t, nil
}

func formatFilename(name string) string {
	characters := [][]string{
		{"\\", "￥"},
		{":", "："},
		{"*", "＊"},
		{"?", "？"},
		{"<", "＜"},
		{">", "＞"},
		{"|", "｜"},
		{"/", "／"},
		{" ", ""},
	}

	result := name
	for _, c := range characters {
		result = strings.ReplaceAll(result, c[0], c[1])
	}
	return result
}

func fileExists(filename string) bool {
	_, err := os.Stat(formatFilename(filename))
	return err == nil
}
