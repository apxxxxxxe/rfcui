package io

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func GetLines(path string) ([]string, error) {
	pwd, _ := os.Getwd()
	fp, err := os.Open(filepath.Join(pwd, path))
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	var lines []string
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func WriteLine(path string, line string) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fmt.Fprintln(f, line)
}

func DeleteLine(path string, line string) error {
	lines, err := GetLines(path)
	if err != nil {
		return err
	}

	os.Remove(path)

	result := lines
	for i, l := range lines {
		if l != line {
			WriteLine(path, l)
		}
	}

	return nil
}

func removeDuplicate(arr []string) []string {
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
