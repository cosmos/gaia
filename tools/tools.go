//go:build tools
// +build tools

// This is the canonical way to enforce dependency inclusion in go.mod for tools that are not directly involved in the build process.
// See
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

package tools

//nolint

import (
	_ "github.com/vektra/mockery/v2"
)
