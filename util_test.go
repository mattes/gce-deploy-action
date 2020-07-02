package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionLessThan(t *testing.T) {
	table := []struct {
		a, b   string
		expect bool
	}{
		// a < b

		{"1", "1", false},
		{"2", "1", false},
		{"1", "2", true},

		{"v1.2.3", "v1.2.3", false},
		{"v1.2.3", "v1.3.3", true},
		{"v1.2.3", "v1.2.4", true},
		{"v1.2.3", "v2.2.3", true},

		{"instance-12", "instance-12", false},
		{"instance-12", "instance-11", false},
		{"instance-12", "instance-13", true},

		{"1.1", "2", true},
		{"1.2.3", "2", true},
		{"2", "1.1", false},
		{"2", "1.2.3", false},

		{"instance-9-14be32d", "instance-17-1df06f1", true},
	}

	for _, test := range table {
		out := VersionLessThan(test.a, test.b)
		require.Equal(t, test.expect, out)

		out = VersionGreaterThan(test.a, test.b)
		require.Equal(t, !test.expect, out)
	}
}

func TestExtractInts(t *testing.T) {
	table := []struct {
		in     string
		expect []int
	}{
		{"123", []int{123}},
		{"abc-123", []int{123}},
		{"abc-123-def", []int{123}},
		{"abc-123-def-456", []int{123, 456}},
		{"v1.2.3", []int{1, 2, 3}},
		{"v.1.2.3", []int{1, 2, 3}},
		{"1.2.3", []int{1, 2, 3}},
		{"1/2/3", []int{1, 2, 3}},
		{"1_2_3", []int{1, 2, 3}},
		{"01_02_03", []int{1, 2, 3}},
		{"+1", []int{1}},
		{"-1", []int{1}},
		{"+0", []int{0}},
		{"-0", []int{0}},
		{"0", []int{0}},
	}

	for _, test := range table {
		out, err := extractInts(test.in)
		require.NoError(t, err, test.in)
		require.Equal(t, test.expect, out, test.in)
	}
}
