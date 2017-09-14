package main

import (
	"bufio"
	"strings"
	"testing"
)

var (
	logLines = `10.10.180.161 - 50.112.166.232, 192.33.28.238 - - - [02/Aug/2015:15:56:14 +0000]  https https https "GET /our-products HTTP/1.1" 200 35967 "-" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.81 Safari/537.36"
50.112.166.232 - 50.112.166.232, 192.33.28.238 - - - [02/Aug/2015:15:56:14 +0000]  http https https "GET /our-products HTTP/1.0" 404 52176 "-" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.81 Safari/537.36"
50.112.166.232 - 50.112.166.232, 192.33.28.238, 50.112.166.232,127.0.0.1 - - - [02/Aug/2015:15:56:14 +0000]  http https,http https,http "GET /api/v1/user HTTP/1.1" 200 3350 "https://release.dollarshaveclub.com/our-products" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.81 Safari/537.36"
10.10.180.161 - 50.112.166.232, 192.33.28.238 - - - [02/Aug/2015:15:56:14 +0000]  https https https "GET /api/v1/user HTTP/1.1" 200 3350 "https://release.dollarshaveclub.com/our-products" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.81 Safari/537.36"
50.112.166.232 - 50.112.166.232, 192.33.28.238 - - - [02/Aug/2015:15:56:14 +0000]  http https https "GET /api/v1/user HTTP/1.0" 200 3350 "https://release.dollarshaveclub.com/our-products" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.81 Safari/537.36"
50.112.166.232 - 50.112.166.232, 192.33.28.238, 50.112.166.232,127.0.0.1 - - - [02/Aug/2015:15:56:27 +0000]  http https,http https,http "POST /api/v1/subscriptions/build HTTP/1.1" 200 6058 "https://release.dollarshaveclub.com/blades" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.81 Safari/537.36"
10.10.180.40 - 50.112.166.232, 192.33.28.238 - - - [02/Aug/2015:15:56:27 +0000]  https https https "POST /api/v1/subscriptions/build HTTP/1.1" 503 6058 "https://release.dollarshaveclub.com/blades" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.81 Safari/537.36"
50.112.166.232 - 50.112.166.232, 192.33.28.238 - - - [02/Aug/2015:15:56:27 +0000]  http https https "POST /api/v1/subscriptions/build HTTP/1.0" 200 6058 "https://release.dollarshaveclub.com/blades" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.81 Safari/537.36"
10.10.180.161 - 50.112.166.232, 192.33.28.238 - - - [02/Aug/2015:15:56:28 +0000]  https https https "POST /api/v1/subscriptions/build HTTP/1.1" 301 6735 "https://release.dollarshaveclub.com/blades" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.81 Safari/537.36"
50.112.166.232 - 50.112.166.232, 192.33.28.238 - - - [02/Aug/2015:15:56:28 +0000]  http https https "POST /api/v1/subscriptions/build HTTP/1.0" 200 6735 "https://release.dollarshaveclub.com/blades" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.81 Safari/537.36"`
)

func TestGetStatus(t *testing.T) {
	tests := []struct {
		name           string
		line           []string
		expectedStatus string
		expectedRoute  string
	}{
		{
			name:           "20x",
			line:           []string{"GET /our-products/shave?action=redeem-gift-card HTTP/1.1", "200"},
			expectedStatus: "20x",
			expectedRoute:  "",
		},
		{
			name:           "30x",
			line:           []string{"POST /our-products/shave?action=redeem-gift-card HTTP/1.1", "301"},
			expectedStatus: "30x",
			expectedRoute:  "",
		},
		{
			name:           "40x",
			line:           []string{"GET /our-products/shave?action=redeem-gift-card HTTP/1.1", "403"},
			expectedStatus: "40x",
			expectedRoute:  "",
		},
		{
			name:           "50x",
			line:           []string{"GET /our-products/shave?action=redeem-gift-card HTTP/1.1", "503"},
			expectedStatus: "50x",
			expectedRoute:  "/our-products/shave",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, r, err := getStatus(test.line)
			if err != nil {
				t.Errorf("unexpected error for test %s: %v", test.name, err)
			}
			if s != test.expectedStatus {
				t.Errorf("unexpected status: expected %s, got %s", test.expectedStatus, s)
			}
			if r != test.expectedRoute {
				t.Errorf("unexpected route: expected %s, got %s", test.expectedRoute, r)
			}
		})
	}
}

func BenchmarkGetStatus(b *testing.B) {
	line := []string{`"GET /our-products/shave?action=redeem-gift-card HTTP/1.1" 503`, "GET /our-products/shave?action=redeem-gift-card HTTP/1.1", "503"}
	for i := 0; i < b.N; i++ {
		getStatus(line)
	}
}

func TestStats(t *testing.T) {

	expectedStats := []struct {
		lvl   string
		value int
	}{
		{
			lvl:   "20x",
			value: 6,
		},
		{
			lvl:   "30x",
			value: 1,
		},
		{
			lvl:   "40x",
			value: 1,
		},
		{
			lvl:   "50x",
			value: 1,
		},
	}

	expectedRoutes := []struct {
		path  string
		value int
	}{
		{
			path:  "/api/v1/subscriptions/build",
			value: 1,
		},
	}

	reader := bufio.NewReader(strings.NewReader(logLines))

	s, r := stats(reader)

	for _, stat := range expectedStats {
		if s[stat.lvl] != stat.value {
			t.Errorf("unexpected number of requests for status level %s: expected %d, got %d\n", stat.lvl, stat.value, s[stat.lvl])
		}
	}

	for _, route := range expectedRoutes {
		if _, ok := r[route.path]; !ok {
			t.Errorf("missing path in returned routes: expected %s", route.path)
			continue
		}
		if r[route.path] != route.value {
			t.Errorf("unexpected number of requests for route %s: expected %d, got %d\n", route.path, route.value, r[route.path])
		}
	}
}

func TestLineParse(t *testing.T) {
	line := `50.112.166.232 - 50.112.166.232, 192.33.28.238 - - - [02/Aug/2015:15:56:28 +0000]  http https https "POST /api/v1/subscriptions/build HTTP/1.0" 200 6735 "https://release.dollarshaveclub.com/blades" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.81 Safari/537.36"`

	expectedReq := "POST /api/v1/subscriptions/build HTTP/1.0"
	expectedStatus := "200"

	rtn := parseLine(line)
	if rtn[0] != expectedReq {
		t.Errorf("unexpected request line returned, expected %s, got %s\n", expectedReq, rtn[0])
	}
	if rtn[1] != expectedStatus {
		t.Errorf("unexpected status code returned, expected %s, got %s\n", expectedStatus, rtn[1])
	}
}

// Keep the compiler for possibly optimizing this too much
// https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go
var result []string

func BenchmarkLineParse(b *testing.B) {
	line := `50.112.166.232 - 50.112.166.232, 192.33.28.238 - - - [02/Aug/2015:15:56:28 +0000]  http https https "POST /api/v1/subscriptions/build HTTP/1.0" 200 6735 "https://release.dollarshaveclub.com/blades" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.81 Safari/537.36"`

	var rtn []string

	for i := 0; i < b.N; i++ {
		rtn = parseLine(line)
	}

	result = rtn
}
