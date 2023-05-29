package rt

import (
    `fmt`
    `unsafe`

    `github.com/chenzhuoyu/iasm/x86_64`
    `github.com/chenzhuoyu/jitkit/loader`
)

const (
    ZeroStep    = 16
    MaxZeroSize = 65536
)

func toaddr(p *x86_64.Label) uintptr {
    if v, err := p.Evaluate(); err != nil {
        panic(err)
    } else {
        return uintptr(v)
    }
}

func asmmemzero() MemZeroFn {
    p := x86_64.DefaultArch.CreateProgram()
    x := make([]*x86_64.Label, MaxZeroSize / ZeroStep + 1)

    /* create all the labels */
    for i := range x {
        x[i] = x86_64.CreateLabel(fmt.Sprintf("zero_%d", i * ZeroStep))
    }

    /* fill backwards */
    for n := MaxZeroSize; n >= ZeroStep; n -= ZeroStep {
        p.Link(x[n / ZeroStep])
        p.MOVDQU(x86_64.XMM15, x86_64.Ptr(x86_64.RDI, int32(n - ZeroStep)))
    }

    /* finish the function */
    p.Link(x[0])
    p.RET()

    /* assemble the function */
    c := p.Assemble(0)
    r := make([]uintptr, len(x))

    /* resolve all the labels */
    for i, v := range x {
        r[i] = toaddr(v)
    }

    /* load the function */
    defer p.Free()
    return MemZeroFn {
        Sz: r,
        Fn: *(*unsafe.Pointer)(loader.Loader(c).Load("_frugal_memzero", Frame{})),
    }
}
