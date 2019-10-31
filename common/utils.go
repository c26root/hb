package common

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

func PrintInfo(l int) {
	var n int
	var lines []string
	flag.Visit(func(flag *flag.Flag) {
		if flag.Value.String() != "" {
			line := fmt.Sprintf("%s: %s", flag.Name, flag.Value)
			lines = append(lines, line)
		}
	})
	lines = append(lines, fmt.Sprintf("number of scans: %d", l))

	for _, line := range lines {
		if len(line) >= n {
			n = len(line)
		}
	}

	width, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 50
	}
	fmt.Println(strings.Repeat("*", width))
	for _, line := range lines {
		fmt.Println(line)
	}
	fmt.Println(strings.Repeat("*", width))
}
