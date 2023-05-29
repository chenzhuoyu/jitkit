//go:build !go1.17 || go1.21

package loader

import (
    `github.com/chenzhuoyu/jitkit/internal/rt`
)

// triggers a compilation error
const (
    _ = panic("Unsupported Go version. Supported versions are 1.17 ~ 1.20")
)

func registerFunction(_ string, _ uintptr, _ uintptr, _ rt.Frame) {
    panic("Unsupported Go version. Supported versions are 1.17 ~ 1.20")
}
