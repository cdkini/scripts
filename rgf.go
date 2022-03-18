package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var r = regexp.MustCompile("(.+)\\:(\\d+)\\:\\s*(.+)")
const GlobalPadding int = 2

type RgResult struct {
	File    string
	Line    string
	Preview string
    Padding int
}

func (r RgResult) toString() string {
    preview :=  strings.Repeat(" ", r.Padding) + r.Preview
    return fmt.Sprintf("%s:%s%s", r.File, r.Line, preview)
}

func main() {
	pattern, path, err := ParseArgs()
	if err != nil {
		log.Fatal(err)
	}

	results, err := RunRg(pattern, path)
	if err != nil {
		log.Fatal(err)
	}

	ReviewResultsInFile(results)
}

func ParseArgs() (string, string, error) {
	args := os.Args[1:]

	switch len(args) {
	case 2:
		return args[0], args[1], nil
	case 1:
		return args[0], ".", nil
	default:
		return "", "", fmt.Errorf("Invalid number of arguments provided!")
	}
}

func RunRg(pattern, path string) ([]RgResult, error) {
	out, err := exec.Command("rg", "-n", pattern, path).Output()
	if err != nil {
		return nil, err
	}
	rawResults := strings.Split(string(out), "\n")
	results := parseResults(rawResults)

	return results, nil
}

func parseResults(rawResults []string) (results []RgResult) {
    longest := 0
	for _, result := range rawResults {
		if _, err := getIndexOfRgDelimiter(result); err == nil {
			matches := r.FindStringSubmatch(result)
			file := matches[1]
			line := matches[2]
			preview := matches[3]
            if refLength := len(file) + len(line); refLength > longest {
                longest = refLength
            }
			results = append(results, RgResult{file, line, preview, GlobalPadding})
		}
	}

    for i := range results {
        result := &results[i]
        padding := longest - (len(result.File) + len(result.Line))
        if padding > 0 {
            result.Padding += padding
        }
    }

	return results
}

func getIndexOfRgDelimiter(result string) (int, error) {
	firstSeen := false
	for i, char := range result {
		if char == ':' {
			if firstSeen {
				return i + 1, nil
			}
			firstSeen = true
		}
	}
	return -1, fmt.Errorf("The input string was not of the expected format!")
}

func ReviewResultsInFile(results []RgResult) {
	file, err := ioutil.TempFile(".", ".rg")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	if err := writeResults(file, results); err != nil {
		log.Fatal(err)
	}
	if err := openEditor(file); err != nil {
		log.Fatal(err)
	}
}

func writeResults(file *os.File, results []RgResult) error {
	w := bufio.NewWriter(file)
	for _, result := range results {
		fmt.Fprintln(w, result.toString())
	}
	return w.Flush()
}

func openEditor(file *os.File) error {
	cmd := exec.Command("nvim", file.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
