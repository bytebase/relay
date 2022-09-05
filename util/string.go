package util

import (
	"fmt"
	"strconv"
	"strings"
)

// Replace last n letter with x
func RedactLastN(s string, n int) string {
	if n <= 0 {
		return s
	}
	out := []rune(s)
	end := 0
	if len(out)-n > 0 {
		end = len(out) - n
	}
	for i := len(out) - 1; i >= end; i-- {
		out[i] = 'x'
	}
	return string(out)
}

// Parse host
func ParseHost(host string) (h string, p int, err error) {
	if host == "" {
		return "", 0, fmt.Errorf("empty host")
	}

	isHttps := false
	host = strings.TrimSpace(host)
	if strings.HasPrefix(host, "http://") {
		host = strings.TrimPrefix(host, "http://")
	} else if strings.HasPrefix(host, "https://") {
		host = strings.TrimPrefix(host, "https://")
		isHttps = true
	}

	fields := strings.Split(host, ":")
	if len(fields) == 1 {
		h = fields[0]
		if isHttps {
			p = 433
		} else {
			p = 80
		}
	} else {
		port, err := strconv.Atoi(fields[len(fields)-1])
		if err != nil {
			return "", 0, fmt.Errorf("port is not a number: %s", fields[len(fields)-1])
		}
		p = port
		h = strings.Join(fields[0:len(fields)-1], ":")
	}
	return h, p, nil
}
