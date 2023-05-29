package pgen

import (
    `github.com/chenzhuoyu/iasm/x86_64`
    `github.com/chenzhuoyu/jitkit/hir`
    `github.com/chenzhuoyu/jitkit/internal/rt`
)

func (self *CodeGen) wbStorePointer(p *x86_64.Program, s hir.PointerRegister, d *x86_64.MemoryOperand) {
    st := x86_64.CreateLabel("_wb_store")
    ra := x86_64.CreateLabel("_wb_return")

    /* check for write barrier */
    p.MOVQ (uintptr(rt.V_pWriteBarrier), RAX)
    p.CMPB (0, Ptr(RAX, 0))
    p.JNE  (st)

    /* check for storing nil */
    if s == hir.Pn {
        p.MOVQ(0, d)
    } else {
        p.MOVQ(self.r(s), d)
    }

    /* set source pointer */
    wbSetSrc := func() {
        if s == hir.Pn {
            p.XORL(EAX, EAX)
        } else {
            p.MOVQ(self.r(s), RAX)
        }
    }

    /* set target slot pointer */
    wbSetSlot := func() {
        if !isSimpleMem(d) {
            p.LEAQ(d.Retain(), RDI)
        } else {
            p.MOVQ(d.Addr.Memory.Base, RDI)
        }
    }

    /* write barrier wrapper */
    wbStoreFn := func(p *x86_64.Program) {
        wbSetSrc                ()
        wbSetSlot               ()
        self.abiSpillReserved   (p)
        self.abiLoadReserved    (p)
        p.MOVQ                  (uintptr(rt.F_gcWriteBarrier), RSI)
        p.CALLQ                 (RSI)
        self.abiSaveReserved    (p)
        self.abiRestoreReserved (p)
        p.JMP                   (ra)
    }

    /* defer the call to the end of generated code */
    p.Link(ra)
    self.later(st, wbStoreFn)
}
