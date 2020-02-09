package plist

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

var isLikeRegex = regexp.MustCompile(`(?m)\"isLike\" => (\d)`)

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
	cmd := fmt.Sprintf("plutil -p - | grep \"%s\" | awk '/{value/ {print $NF}'", value)
	command := exec.Command("bash", "-c", cmd)
	out, err := execute(command, bytes.NewReader(*plist))
	if err != nil {
		return -1, err
	}

	intOut, err := strconv.Atoi(strings.Replace(strings.TrimSpace(out), "}", "", -1))
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Failed with string: %s", out))
		return -1, err
	}
	return intOut, nil
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
	cmd := fmt.Sprintf("plutil -p - | awk '/isLike/ {print $NF}'")
	command := exec.Command("bash", "-c", cmd)
	out, err := execute(command, bytes.NewReader(*plist))
	if err != nil {
		return false, err
	}

	boolOut, err := strconv.ParseBool(strings.TrimSpace(out))
	if err != nil {
		return false, err
	}
	return boolOut, nil
}

// IsLikeNative will determine if a plist blob indicates this comment is a like
func IsLikeNative(plist *[]byte) (bool, error) {
	if plist == nil {
		return false, fmt.Errorf("Empty plist")
	}
	cmd := fmt.Sprintf("/usr/bin/plutil -p -")
	command := exec.Command(cmd)
	out, err := execute(command, bytes.NewReader(*plist))
	if err != nil {
		return false, err
	}

	isLikeString := isLikeRegex.FindString(out)
	fmt.Println(out)

	boolOut, err := strconv.ParseBool(strings.TrimSpace(isLikeString))
	if err != nil {
		return false, err
	}
	return boolOut, nil
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

	out, err := ExtractValueAtLine(plist, line)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}
