package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var isLikeRegex = regexp.MustCompile(`(?m)\"isLike\" => (\d)`)
var lineNumberRegex = regexp.MustCompile(`(?m){value = (\d)}`)

func execute(cmd *exec.Cmd, stdin io.Reader) (string, error) {
	var out bytes.Buffer
	var err2 bytes.Buffer
	cmd.Stdin = stdin
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	if err2.String() != "" {
		return "", fmt.Errorf(err2.String())
	}

	return out.String(), nil
}

// GetLineNumber returns the line number a key's value is located in the given plist
func GetLineNumber(plist *[]byte, value string) (int, error) {
	if plist == nil {
		return -1, fmt.Errorf("Empty plist")
	}

	command := exec.Command("/usr/bin/plutil", "-p", "-")
	out, err := execute(command, bytes.NewReader(*plist))
	if err != nil {
		return -1, err
	}

	lineNumber := -1
	lineNumberLines := strings.Split(out, "\n")
	for _, line := range lineNumberLines {
		if strings.Contains(line, value) {
			lineNumberMatches := lineNumberRegex.FindStringSubmatch(out)

			if len(lineNumberMatches) > 0 {
				lineNumber, err = strconv.Atoi(lineNumberMatches[1])
				if err != nil {
					return -1, err
				}
				break
			}
		}
	}

	return lineNumber, nil
}

// ExtractValueAtLine returns the value at a given line number in the plist
func ExtractValueAtLine(plist *[]byte, line int) (string, error) {
	if plist == nil {
		return "", fmt.Errorf("Empty plist")
	}
	cmd := fmt.Sprintf("plutil -extract '$objects.%d' xml1 -o - - | awk -F \"[<>]\" '/<string>/ {print $3}'", line)
	command := exec.Command("bash", "-c", cmd)
	out, err := execute(command, bytes.NewReader(*plist))
	if err != nil {
		return "", err
	}

	return out, nil
}

// IsLike will determine if a plist blob indicates this comment is a like
func IsLike(plist *[]byte) (bool, error) {
	if plist == nil {
		return false, fmt.Errorf("Empty plist")
	}
	command := exec.Command("/usr/bin/plutil", "-p", "-")
	out, err := execute(command, bytes.NewReader(*plist))
	if err != nil {
		return false, err
	}

	isLikeMatches := isLikeRegex.FindStringSubmatch(out)

	isLike := false
	if len(isLikeMatches) > 0 {
		isLike, err = strconv.ParseBool(isLikeMatches[1])
		if err != nil {
			return false, err
		}
	}
	return isLike, nil
}

// GetValue will parse the given plist blob and search for the requested value
func GetValue(plist *[]byte, value string) (string, error) {
	if plist == nil {
		return "", fmt.Errorf("Empty plist")
	}
	line, err := GetLineNumber(plist, value)
	if err != nil {
		return "", err
	}
	if line == -1 {
		return "", fmt.Errorf("%s not found in document", value)
	}

	out, err := ExtractValueAtLine(plist, line)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func main() {
	var (
		plistPath string
		value     string
	)

	flag.StringVar(&plistPath, "plist", "", "Path to the plist file")
	flag.StringVar(&value, "value", "", "Value to search the plist for")
	flag.Parse()

	f, err := os.Open(plistPath)
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	out, err := GetLineNumber(&b, value)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)

	out2, err := ExtractValueAtLine(&b, out)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out2)
}
