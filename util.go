package main

import (
	"regexp"
	"strconv"
)

// VersionLessThan returns true if a < b
func VersionLessThan(a, b string) bool {
	xa, err := extractInts(a)
	if err != nil {
		panic(err)
	}

	xb, err := extractInts(b)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(xa); i++ {
		if len(xb) >= i {
			if xa[i] == xb[i] {
				continue
			}

			return xa[i] < xb[i]
		}
	}

	return false
}

// VersionGreaterThan returns true if a > b
func VersionGreaterThan(a, b string) bool {
	return !VersionLessThan(a, b)
}

func extractInts(v string) ([]int, error) {
	re := regexp.MustCompile("[0-9]+")
	matches := re.FindAllString(v, -1)

	r := make([]int, 0)
	for _, m := range matches {
		n, err := strconv.Atoi(m)
		if err != nil {
			return nil, err
		}
		r = append(r, n)
	}

	return r, nil
}
