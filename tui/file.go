package tui

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func getLines(path string) (int, []string, error) {
	pwd, _ := os.Getwd()
	fp, err := os.Open(filepath.Join(pwd, path))
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

func writeLine(path string, line string) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Fprintln(file, line)
}

func deleteLine(path string, line string) error {
	_, lines, err := getLines(path)
	if err != nil {
		return err
	}

	os.Remove(path)

	for _, l := range lines {
		if l != line {
			writeLine(path, l)
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
