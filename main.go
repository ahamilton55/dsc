package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	defaultLogInput  = "/var/log/nginx/access.log"
	defaultLogOutput = "/var/log/stats.log"
	defaultTimeout   = 5
)

var (
	// reqStatusRegexp is used to pull out the request string and status code from each log line
	// Though regexp aren't always the most performant, it's much easier here to grab the small amount of info we want than with splitting strings
	reqStatusRegexp = `"([A-Z]+ .*)" ([2-5][0-9]+)`

	logInput  string
	logOutput string
	seconds   int64
)

func init() {
	// Flags to make it easier to test out the program
	// Defaults are set here too
	flag.StringVar(&logInput, "input", defaultLogInput, "input file used to get log entries")
	flag.StringVar(&logOutput, "output", defaultLogOutput, "output file used to collect stats")
	flag.Int64Var(&seconds, "timeout", defaultTimeout, "timeout in seconds to wait between collections")
}

func main() {
	re := regexp.MustCompile(reqStatusRegexp)
	flag.Parse()

	timeout := defaultTimeout * time.Second

	input := openFile(logInput)
	defer input.Close()

	var output *os.File
	if logOutput == "/dev/stdout" {
		output = os.Stdout
	} else {
		output = openFile(logOutput)
	}
	defer output.Close()

	reader := bufio.NewReader(input)

	for {
		statuses, badRoutes := stats(re, reader)

		// Specify statuses here since if we don't read any new long lines we, the map will not have any fields
		for _, level := range []string{"50x", "40x", "30x", "20x"} {
			fmt.Fprintf(output, "%s:%d|s\n", level, statuses[level])
		}

		for route, hits := range badRoutes {
			fmt.Fprintf(output, "%s:%d|s\n", route, hits)
		}

		wait := time.After(timeout)
		<-wait
	}
}

func openFile(filename string) *os.File {
	f, err := os.Open(filename)
	for err != nil {
		if err != nil {
			log.Printf("error opening log file: %v\n", err)
		}

		fileWait := time.After(2 * time.Second)
		<-fileWait

		f, err = os.Open(filename)
	}

	return f
}

// stats will read the file and pick out information using the provided regular expression
func stats(re *regexp.Regexp, reader *bufio.Reader) (map[string]int, map[string]int) {
	var readErr error
	var byteLine []byte
	s := make(map[string]int)
	r := make(map[string]int)

	byteLine, readErr = reader.ReadBytes('\n')

	// Read the file until we reach the end
	for readErr != io.EOF {
		line := re.FindStringSubmatch(string(byteLine))

		if len(line) >= 2 {
			code, route, err := getStatus(line)
			if err != nil {
				continue
			}

			s[code]++

			if route != "" {
				r[route]++
			}
		}

		byteLine, readErr = reader.ReadBytes('\n')
	}

	return s, r
}

// getStatus will find the status and route (if required) from a provided line
func getStatus(line []string) (string, string, error) {
	var s string
	var r string

	routeStatus := strings.Split(line[0], " ")
	status, err := strconv.Atoi(line[2])
	if err != nil {
		return s, r, err
	}

	switch {
	case status >= 200 && status < 300:
		s = "20x"
	case status >= 300 && status < 400:
		s = "30x"
	case status >= 400 && status < 500:
		s = "40x"
	case status >= 500:
		s = "50x"

		rawRoute, err := url.Parse(routeStatus[1])
		if err != nil {
			return "", "", err
		}
		r = rawRoute.Path
	}

	return s, r, nil
}
