package rt

import (
    `unsafe`
)

//go:linkname writeBarrier runtime.writeBarrier
var writeBarrier uintptr

//go:nosplit
//go:linkname gcWriteBarrier runtime.gcWriteBarrier
func gcWriteBarrier()

var (
    V_pWriteBarrier  = unsafe.Pointer(&writeBarrier)
    F_gcWriteBarrier = FuncAddr(gcWriteBarrier)
)
