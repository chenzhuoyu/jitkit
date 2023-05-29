package rt

import (
    `unsafe`
)

type MemZeroFn struct {
    Fn unsafe.Pointer
    Sz []uintptr
}

var (
    MemZero = asmmemzero()
)

func (self MemZeroFn) ForSize(n uintptr) unsafe.Pointer {
    return unsafe.Pointer(uintptr(self.Fn) + self.Sz[n / ZeroStep])
}
