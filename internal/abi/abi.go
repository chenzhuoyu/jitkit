package abi

import (
    `unsafe`

    `github.com/chenzhuoyu/jitkit/internal/rt`
)

type AbstractABI interface {
    RegisterMethod(id int, mt rt.Method) int
    RegisterFunction(id int, fn interface{}) unsafe.Pointer
}

var (
    ABI = ArchCreateABI()
)

func alignUp(n uintptr, a int) uintptr {
    return (n + uintptr(a) - 1) &^ (uintptr(a) - 1)
}
