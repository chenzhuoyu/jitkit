package rt

import (
    `reflect`
    `unsafe`

    `github.com/chenzhuoyu/jitkit/internal/abi`
)

//go:noescape
//go:linkname memmove runtime.memmove
//goland:noinspection GoUnusedParameter
func memmove(to unsafe.Pointer, from unsafe.Pointer, n uintptr)

var (
    F_memmove = FuncAddr(memmove)
    R_memmove = resolveClobberSet(memmove)
    S_memmove = abi.ABI.LayoutFunc(-1, reflect.TypeOf(memmove))
)
