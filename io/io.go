package io

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

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

func SaveBytes(data []byte, path string) error {
	file, err := os.Create(path)
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

func DirWalk(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, DirWalk(filepath.Join(dir, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(dir, file.Name()))
	}

	return paths
}
