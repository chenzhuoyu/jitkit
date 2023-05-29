package rt

import (
    _ `unsafe`
)

//go:linkname morestack_noctxt runtime.morestack_noctxt
func morestack_noctxt()

var (
    F_morestack_noctxt = FuncAddr(morestack_noctxt)
)
