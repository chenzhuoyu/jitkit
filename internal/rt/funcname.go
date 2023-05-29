package rt

import (
    `fmt`
    `runtime`
    `unsafe`
)

func FuncName(p unsafe.Pointer) string {
    if fn := runtime.FuncForPC(uintptr(p)); fn == nil {
        return "???"
    } else if fp := fn.Entry(); fp == uintptr(p) {
        return fn.Name()
    } else {
        return fmt.Sprintf("%s+%#x", fn.Name(), uintptr(p) - fp)
    }
}
