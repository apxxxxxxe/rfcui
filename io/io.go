package io

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func IsFile(filename string) bool {
	_, err := os.OpenFile(filename, os.O_RDONLY, 0)
	return !os.IsNotExist(err)
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

func GetLines(path string) (int, []string, error) {
	fp, err := os.Open(path)
	if err != nil {
		return 0, nil, err
	}
	defer fp.Close()

	var lines []string
	scanner := bufio.NewScanner(fp)
	lineCount := 0
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		lineCount++
	}
	if err := scanner.Err(); err != nil {
		return 0, nil, err
	}
	return lineCount, lines, nil
}

func WriteLine(path string, line string) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Fprintln(file, line)
}

func DeleteLine(path string, line string) error {
	_, lines, err := GetLines(path)
	if err != nil {
		return err
	}

	os.Remove(path)

	for _, l := range lines {
		if l != line {
			WriteLine(path, l)
		}
	}

	return nil
}

func RemoveDuplicate(arr []string) []string {
	results := make([]string, 0, len(arr))
	encountered := map[string]bool{}
	for i := 0; i < len(arr); i++ {
		if !encountered[arr[i]] {
			encountered[arr[i]] = true
			results = append(results, arr[i])
		}
	}
	return results
}
