package hir

import (
    `fmt`
    `runtime`
    `unsafe`

    `github.com/chenzhuoyu/jitkit/internal/abi`
    `github.com/chenzhuoyu/jitkit/internal/rt`
)

type (
    CallType uint8
)

const (
    CCall CallType = iota
    GCall
    ICall
)

type CallState interface {
    Gr(id GenericRegister) uint64
    Pr(id PointerRegister) unsafe.Pointer
    SetGr(id GenericRegister, val uint64)
    SetPr(id PointerRegister, val unsafe.Pointer)
}

type CallHandle struct {
    Id    int
    Slot  int
    Type  CallType
    Func  unsafe.Pointer
    proxy func(CallContext)
}

func (self *CallHandle) Name() string {
    return runtime.FuncForPC(uintptr(self.Func)).Name()
}

func (self *CallHandle) Call(r CallState, p *Ir) {
    self.proxy(CallContext {
        repo: r,
        kind: self.Type,
        argc: p.An,
        retc: p.Rn,
        argv: p.Ar,
        retv: p.Rr,
        itab: p.Ps,
        data: p.Pd,
    })
}

func (self *CallHandle) String() string {
    return fmt.Sprintf("*%#x[%s]", self.Func, self.Name())
}

type CallContext struct {
    kind CallType
    repo CallState
    itab PointerRegister
    data PointerRegister
    argc uint8
    retc uint8
    argv [8]uint8
    retv [8]uint8
}

func (self CallContext) Au(i int) uint64 {
    if p := self.argv[i]; p &ArgPointer != 0 {
        panic("invoke: invalid int argument")
    } else {
        return self.repo.Gr(GenericRegister(p & ArgMask))
    }
}

func (self CallContext) Ap(i int) unsafe.Pointer {
    if p := self.argv[i]; p &ArgPointer == 0 {
        panic("invoke: invalid pointer argument")
    } else {
        return self.repo.Pr(PointerRegister(p & ArgMask))
    }
}

func (self CallContext) Ru(i int, v uint64) {
    if p := self.retv[i]; p &ArgPointer != 0 {
        panic("invoke: invalid int return value")
    } else {
        self.repo.SetGr(GenericRegister(p &ArgMask), v)
    }
}

func (self CallContext) Rp(i int, v unsafe.Pointer) {
    if p := self.retv[i]; p &ArgPointer == 0 {
        panic("invoke: invalid pointer return value")
    } else {
        self.repo.SetPr(PointerRegister(p &ArgMask), v)
    }
}

func (self CallContext) Itab() *rt.GoItab {
    if self.kind != ICall {
        panic("invoke: itab is not available")
    } else {
        return (*rt.GoItab)(self.repo.Pr(self.itab))
    }
}

func (self CallContext) Data() unsafe.Pointer {
    if self.kind != ICall {
        panic("invoke: data is not available")
    } else {
        return self.repo.Pr(self.data)
    }
}

func (self CallContext) Verify(args string, rets string) bool {
    return self.verifySeq(args, self.argc, self.argv) && self.verifySeq(rets, self.retc, self.retv)
}

func (self CallContext) verifySeq(s string, n uint8, v [8]uint8) bool {
    nb := int(n)
    ne := len(s)

    /* sanity check */
    if ne > len(v) {
        panic("invoke: invalid descriptor")
    }

    /* check for value count */
    if nb != ne {
        return false
    }

    /* check for every argument */
    for i := 0; i < nb; i++ {
        switch s[i] {
            case 'i' : if v[i] &ArgPointer != 0 { return false }
            case '*' : if v[i] &ArgPointer == 0 { return false }
            default  : panic("invoke: invalid descriptor char: " + s[i:i + 1])
        }
    }

    /* all checked ok */
    return true
}

var (
    funcTab []*CallHandle
)

func LookupCall(id int64) *CallHandle {
    if id < 0 || id >= int64(len(funcTab)) {
        panic("invalid function ID")
    } else {
        return funcTab[id]
    }
}

func RegisterICall(mt rt.Method, proxy func(CallContext)) (h *CallHandle) {
    h       = new(CallHandle)
    h.Id    = len(funcTab)
    h.Type  = ICall
    h.Slot  = abi.ABI.RegisterMethod(h.Id, mt)
    h.proxy = proxy
    funcTab = append(funcTab, h)
    return
}

func RegisterGCall(fn interface{}, proxy func(CallContext)) (h *CallHandle) {
    h       = new(CallHandle)
    h.Id    = len(funcTab)
    h.Type  = GCall
    h.Func  = abi.ABI.RegisterFunction(h.Id, fn)
    h.proxy = proxy
    funcTab = append(funcTab, h)
    return
}

func RegisterCCall(fn unsafe.Pointer, proxy func(CallContext)) (h *CallHandle) {
    h       = new(CallHandle)
    h.Id    = len(funcTab)
    h.Type  = CCall
    h.Func  = fn
    h.proxy = proxy
    funcTab = append(funcTab, h)
    return
}
