package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	computeBeta "google.golang.org/api/compute/v0.beta"
)

func TestFindLatestInstanceGroupManagerVersion(t *testing.T) {
	require.Equal(t, "", findLatestInstanceGroupManagerVersion(nil))
	require.Equal(t, "", findLatestInstanceGroupManagerVersion(
		[]*computeBeta.InstanceGroupManagerVersion{},
	))

	require.Equal(t, "abc-8", findLatestInstanceGroupManagerVersion(
		[]*computeBeta.InstanceGroupManagerVersion{
			{Name: "abc-5"},
			{Name: "abc-6"},
			{Name: "abc-8"},
			{Name: "abc-3"},
		},
	))
}
