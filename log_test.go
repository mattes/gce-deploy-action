package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatLog(t *testing.T) {
	assert.Equal(t, "::command::value\n", formatLog("command", nil, "value"))
	assert.Equal(t, "::command::\n", formatLog("command", nil, ""))
	assert.Equal(t, "::command foo=bar::value\n", formatLog("command", map[string]string{"foo": "bar"}, "value"))
	assert.Equal(t, "::command abc=def,foo=bar::value\n", formatLog("command", map[string]string{"foo": "bar", "abc": "def"}, "value"))
}
