package db

import (
  "github.com/apxxxxxxe/rfcui/feed"

	"encoding/gob"
	"os"
	"path/filepath"
	"strings"

	"errors"
)

func SaveInterface(t *feed.Feed, filename string) error {
	pwd, _ := os.Getwd()
	fp := filepath.Join(pwd, "save")

	if !isDir(fp) {
		if err := os.Mkdir(fp, 0777); err != nil {
			return err
		}
	}

	file := filepath.Join(pwd, "save", formatFilename(filename))
	var f *os.File
	var err error
	if isFile(file) {
		f, err = os.Open(file)
	} else {
		f, err = os.Create(file)
	}
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

func LoadInterface(filename string) (*feed.Feed, error) {
	pwd, _ := os.Getwd()

	fp := filepath.Join(pwd, "save", formatFilename(filename))
	if !isFile(fp) {
		return nil, errors.New("file is not exist: " + fp)
	}

	f, err := os.Open(filepath.Join(fp))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var t *feed.Feed
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

func isFile(filename string) bool {
	_, err := os.Stat(formatFilename(filename))
	return err == nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	os.IsNotExist(err)
	if err != nil || !info.IsDir() {
		return false
	}
	return true
}
