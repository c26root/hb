package common

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

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

	width, _, _ := terminal.GetSize(int(os.Stdout.Fd()))
	fmt.Println(strings.Repeat("*", width))
	for _, line := range lines {
		fmt.Println(line)
	}
	fmt.Println(strings.Repeat("*", width))
}

func Shuffle(vals []string) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for len(vals) > 0 {
		n := len(vals)
		randIndex := r.Intn(n)
		vals[n-1], vals[randIndex] = vals[randIndex], vals[n-1]
		vals = vals[:n-1]
	}
}
