package io

import (
	"bufio"
	"os"
	"path/filepath"
)

func GetLines(path string) []string {
	pwd, _ := os.Getwd()
	fp, err := os.Open(filepath.Join(pwd, path))
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	var lines []string
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return lines
}
