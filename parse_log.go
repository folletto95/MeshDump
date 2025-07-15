package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

var pattern = regexp.MustCompile(`(?P<ts>\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) (store|mqtt).*node=(?P<node>\S+) .*value=(?P<value>\S+)`)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		matches := pattern.FindStringSubmatch(line)
		if len(matches) > 0 {
			ts := matches[pattern.SubexpIndex("ts")]
			node := matches[pattern.SubexpIndex("node")]
			value := matches[pattern.SubexpIndex("value")]
			fmt.Printf("%s\t%s\t%s\n", ts, node, value)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}
}
