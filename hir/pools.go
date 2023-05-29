package hir

import (
    `sync`

    `github.com/chenzhuoyu/jitkit/internal/rt`
)

var (
    instrPool   sync.Pool
    builderPool sync.Pool
)

func newInstr(op OpCode) *Ir {
    if v := instrPool.Get(); v == nil {
        return allocInstr(op)
    } else {
        return resetInstr(op, v.(*Ir))
    }
}

func freeInstr(p *Ir) {
    instrPool.Put(p)
}

func allocInstr(op OpCode) (p *Ir) {
    p = new(Ir)
    p.Op = op
    return
}

func resetInstr(op OpCode, p *Ir) *Ir {
    *p = Ir{Op: op}
    return p
}

func newBuilder() *Builder {
    if v := builderPool.Get(); v == nil {
        return allocBuilder()
    } else {
        return resetBuilder(v.(*Builder))
    }
}

func freeBuilder(p *Builder) {
    builderPool.Put(p)
}

func allocBuilder() (p *Builder) {
    p       = new(Builder)
    p.refs  = make(map[string]*Ir, 64)
    p.pends = make(map[string][]**Ir, 64)
    return
}

func resetBuilder(p *Builder) *Builder {
    p.i    = 0
    p.head = nil
    p.tail = nil
    rt.MapClear(p.refs)
    rt.MapClear(p.pends)
    return p
}
