package plist

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

func execute(cmd *exec.Cmd) (string, error) {
	var out bytes.Buffer
	var err2 bytes.Buffer
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
func GetLineNumber(plist []byte, value string) (int, error) {
	// Convert
	cmd := fmt.Sprintf("plutil -p - | grep \"%s\" | awk '/{value/ {print $NF}'", value)
	convert1 := exec.Command("bash", "-c", cmd)
	convert1.Stdin = bytes.NewReader(plist)
	out, err := execute(convert1)
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
func ExtractValueAtLine(plist []byte, line int) (string, error) {
	cmd := fmt.Sprintf("plutil -extract '$objects.%d' xml1 -o - - | awk -F \"[<>]\" '/<string>/ {print $3}'", line)
	convert1 := exec.Command("bash", "-c", cmd)
	convert1.Stdin = bytes.NewReader(plist)
	out, err := execute(convert1)
	if err != nil {
		return "", err
	}

	return out, nil
}

// IsLike will determine if a plist blob indicates this comment is a like
func IsLike(plist []byte) (bool, error) {
	cmd := fmt.Sprintf("plutil -p - | awk '/isLike/ {print $NF}'")
	convert1 := exec.Command("bash", "-c", cmd)
	convert1.Stdin = bytes.NewReader(plist)
	out, err := execute(convert1)
	if err != nil {
		return false, err
	}

	boolOut, err := strconv.ParseBool(strings.TrimSpace(out))
	if err != nil {
		return false, err
	}
	return boolOut, nil
}

// GetValue will parse the given plist blob and search for the requested value
func GetValue(plist []byte, value string) (string, error) {
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

/*
func plistToJSON(plist []byte) (string, error) {
	if plist == nil {
		return "", fmt.Errorf("Empty PList")
	}

	d1 := plist
	err := ioutil.WriteFile("/tmp/dat1.plist", d1, 0644)
	if err != nil {
		return "", err
	}
	//defer os.Remove("/tmp/dat1.plist")

	// Convert
	cmd := "plutil -convert xml1 -o - /tmp/dat1.plist | sed -e 's/date/string/g' | plutil -convert json -o - -"
	convert1 := exec.Command("bash", "-c", cmd)
	out, err := executeCMD(convert1)
	if err != nil {
		return "", err
	}

	//defer os.Remove("/tmp/out.plist")

	return out, nil

}*/
