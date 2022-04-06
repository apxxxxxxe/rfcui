package db

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"strings"

	"errors"
)

func SaveInterfaces(t interface{}) error {
	pwd, _ := os.Getwd()
	fp := filepath.Join(pwd, "save")

	if !IsDir(fp) {
		if err := os.Mkdir(fp, 0777); err != nil {
			return err
		}
	}

	file := filepath.Join(pwd, "save", "Interfaces")
	var f *os.File
	var err error
	if IsFile(file) {
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

func LoadInterfaces() (interface{}, error) {
	pwd, _ := os.Getwd()

	fp := filepath.Join(pwd, "save", "Interfaces")
	if !IsFile(fp) {
		return nil, errors.New("file is not exist: " + fp)
	}

	f, err := os.Open(fp)
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

func IsFile(filename string) bool {
	_, err := os.Stat(formatFilename(filename))
	return err == nil
}

func IsDir(path string) bool {
	info, err := os.Stat(path)
	os.IsNotExist(err)
	if err != nil || !info.IsDir() {
		return false
	}
	return true
}

func serializeGob(feeds []*feed.Feed) ([]byte, error) {
  buf := bytes.NewBuffer(nil)
  enc := gob.NewEncoder(buf)
  err := enc.Encode(feeds)
  if err != nil {
    return nil, err
  }
  return buf.Bytes(), nil
}

func saveBytes(data []byte, path string) error {
  // ファイルを書き込み用にオープン (mode=0666)
  file, err := os.Create(filepath.Join(".", path))
  if err != nil {
    panic(err)
  }
  defer file.Close()

  _, err = file.Write(data)
  if err != nil {
    panic(err)
  }
  return nil
}

func decode(data []byte) []feed.Feed {
  var feeds []feed.Feed
  buf := bytes.NewBuffer(data)
  _ = gob.NewDecoder(buf).Decode(&feeds)
  return feeds
}
