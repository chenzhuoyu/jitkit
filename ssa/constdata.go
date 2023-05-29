package ssa

import (
    `fmt`
    `unsafe`
)

type _ConstData struct {
    i bool
    v int64
    c Constness
    p unsafe.Pointer
}

func (self _ConstData) String() string {
    if self.i {
        return fmt.Sprintf("(i64) %d", self.v)
    } else {
        return fmt.Sprintf("(%s ptr) %p", self.c, self.p)
    }
}

func constint(v int64) _ConstData {
    return _ConstData {
        v: v,
        i: true,
    }
}

func constptr(p unsafe.Pointer, cc Constness) _ConstData {
    return _ConstData {
        p: p,
        c: cc,
        i: false,
    }
}
