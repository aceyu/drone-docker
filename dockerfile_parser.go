package docker

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var (
	DefaultDomain = "registry.docker.i.fbank.com"
	reg           = regexp.MustCompile(`(?i:^\s*from)\s+(.+)`)
)

func ModifyDockerfile(dockerfile string) error {
	lines, err := readFile(dockerfile)
	if err != nil {
		return fmt.Errorf("Error read docker file: %s", err)
	}
	f, err := os.OpenFile(dockerfile, os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("Error open docker file: %s", err)
	}
	defer f.Close()
	_, err = io.WriteString(f, strings.Join(lines, "\n"))
	if err != nil {
		return fmt.Errorf("Error modify docker file: %s", err)
	}
	return nil
}

func readFile(dockerfile string) ([]string, error) {
	f, err := os.OpenFile(dockerfile, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	lines := make([]string, 0)
	for {
		line, err := r.ReadString('\n')
		line = strings.TrimSpace(line)
		line = readAndModify(line)
		lines = append(lines, line)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	return lines, nil
}

func readAndModify(line string) string {
	if from := reg.FindAllStringSubmatch(line, -1); len(from) != 0 {
		if len(from[0]) == 2 {
			domain, remainder := splitDockerDomain(from[0][1])
			return fmt.Sprintf("FROM %s/%s", domain, remainder)
		}
	}
	return line
}

func splitDockerDomain(name string) (domain, remainder string) {
	i := strings.IndexRune(name, '/')
	if i == -1 || (!strings.ContainsAny(name[:i], ".:") && name[:i] != "localhost") {
		domain, remainder = DefaultDomain, name
	} else {
		domain, remainder = name[:i], name[i+1:]
	}
	if domain == DefaultDomain && !strings.ContainsRune(remainder, '/') {
		remainder = "library/" + remainder
	}
	return
}
