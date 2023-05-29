package pgen

import (
    `fmt`
    `math`
    `unsafe`

    `github.com/chenzhuoyu/iasm/x86_64`
    `github.com/chenzhuoyu/jitkit/hir`
    `github.com/chenzhuoyu/jitkit/internal/abi`
    `github.com/chenzhuoyu/jitkit/internal/rt`
)

type _SwapPair struct {
    rs hir.Register
    rd hir.Register
    rr x86_64.Register64
}

type _CodeGenExtension struct {
    rets []_SwapPair
}

/** Prologue & Epilogue **/

func (self *CodeGen) abiPrologue(p *x86_64.Program) {
    for i, v := range self.ctxt.desc.Args {
        if v.InRegister {
            p.MOVQ(v.Reg, self.ctxt.argv(i))
        }
    }
}

func (self *CodeGen) abiEpilogue(p *x86_64.Program) {
    for _, v := range self.abix.rets {
        p.XCHGQ(self.r(v.rs), v.rr)
        self.regs[v.rs], self.regs[v.rd] = self.regs[v.rd], self.regs[v.rs]
    }
}

/** Stack Checking & Growing **/

const (
    _M_memcpyargs  = 0
    _G_stackguard0 = 0x10
)

func (self *CodeGen) abiStackGrow(p *x86_64.Program) {
    self.internalSpillArgs(p)
    p.MOVQ(uintptr(rt.F_morestack_noctxt), R12)
    p.CALLQ(R12)
    self.internalUnspillArgs(p)
}

func (self *CodeGen) abiStackCheck(p *x86_64.Program, to *x86_64.Label, sp uintptr) {
    p.LEAQ (Ptr(RSP, -self.ctxt.size() - int32(sp)), R12)
    p.CMPQ (Ptr(R14, _G_stackguard0), R12)
    p.JBE  (to)
}

func (self *CodeGen) internalSpillArgs(p *x86_64.Program) {
    for _, v := range self.ctxt.desc.Args {
        if v.InRegister {
            p.MOVQ(v.Reg, Ptr(RSP, int32(v.Mem) +abi.PtrSize))
        }
    }
}

func (self *CodeGen) internalUnspillArgs(p *x86_64.Program) {
    for _, v := range self.ctxt.desc.Args {
        if v.InRegister {
            p.MOVQ(Ptr(RSP, int32(v.Mem) +abi.PtrSize), v.Reg)
        }
    }
}

/** Reserved Register Management **/

func (self *CodeGen) abiSaveReserved(p *x86_64.Program) {
    for rr := range self.ctxt.regr {
        p.MOVQ(rr, self.ctxt.rslot(rr))
    }
}

func (self *CodeGen) abiLoadReserved(p *x86_64.Program) {
    for rr := range self.ctxt.regr {
        p.MOVQ(self.ctxt.rslot(rr), rr)
    }
}

func (self *CodeGen) abiSpillReserved(p *x86_64.Program) {
    for rr := range self.ctxt.regr {
        if lr := self.rindex(rr); lr != nil {
            p.MOVQ(rr, self.ctxt.slot(lr))
        }
    }
}

func (self *CodeGen) abiRestoreReserved(p *x86_64.Program) {
    for rr := range self.ctxt.regr {
        if lr := self.rindex(rr); lr != nil {
            p.MOVQ(self.ctxt.slot(lr), rr)
        }
    }
}

/** Argument & Return Value Management **/

func (self *CodeGen) abiLoadInt(p *x86_64.Program, i int, d hir.GenericRegister) {
    p.MOVQ(self.ctxt.argv(i), self.r(d))
}

func (self *CodeGen) abiLoadPtr(p *x86_64.Program, i int, d hir.PointerRegister) {
    p.MOVQ(self.ctxt.argv(i), self.r(d))
}

func (self *CodeGen) abiStoreInt(p *x86_64.Program, s hir.GenericRegister, i int) {
    self.internalStoreRet(p, s, i)
}

func (self *CodeGen) abiStorePtr(p *x86_64.Program, s hir.PointerRegister, i int) {
    self.internalStoreRet(p, s, i)
}

func (self *CodeGen) internalStoreRet(p *x86_64.Program, s hir.Register, i int) {
    var r hir.Register
    var m abi.Parameter

    /* if return with stack, store directly */
    if m = self.ctxt.desc.Rets[i]; !m.InRegister {
        p.MOVQ(self.r(s), self.ctxt.retv(i))
        return
    }

    /* check if the value is the very register required for return */
    if self.r(s) == m.Reg {
        return
    }

    /* if return with free registers, simply overwrite with new value */
    if r = self.rindex(m.Reg); r == nil {
        p.MOVQ(self.r(s), m.Reg)
        return
    }

    /* if not, mark the register to store later */
    self.abix.rets = append(self.abix.rets, _SwapPair {
        rs: s,
        rd: r,
        rr: m.Reg,
    })
}

/** Memory Copy & Zeroing **/

var memcpyargs = [256]bool {
    RAX: true,
    RBX: true,
    RCX: true,
}

func (self *CodeGen) abiBlockCopy(p *x86_64.Program, pd hir.PointerRegister, ps hir.PointerRegister, nb hir.GenericRegister) {
    rd := self.r(pd)
    rs := self.r(ps)
    rl := self.r(nb)

    /* save all the registers, if they will be clobbered */
    for _, lr := range self.ctxt.regs {
        if rr := self.r(lr); rt.R_memmove[rr] || memcpyargs[rr] {
            p.MOVQ(rr, self.ctxt.slot(lr))
        }
    }

    /* enumerate different register cases */
    switch {
        case rs == RBX && rl == RCX : p.MOVQ(rd, RAX)
        case rs == RBX && rl != RCX : p.MOVQ(rd, RAX); p.MOVQ  (rl, RCX)
        case rs != RBX && rl == RCX : p.MOVQ(rd, RAX); p.MOVQ  (rs, RBX)
        case rs == RCX && rl == RBX : p.MOVQ(rd, RAX); p.XCHGQ (RBX, RCX)
        case rs == RCX && rl != RBX : p.MOVQ(rd, RAX); p.MOVQ  (RCX, RBX); p.MOVQ(rl, RCX)
        case rs != RCX && rl == RBX : p.MOVQ(rd, RAX); p.MOVQ  (RBX, RCX); p.MOVQ(rs, RBX)
        default                     : p.MOVQ(rd, RAX); p.MOVQ  (rs, RBX);  p.MOVQ(rl, RCX)
    }

    /* call the function */
    p.MOVQ(uintptr(rt.F_memmove), RDI)
    p.CALLQ(RDI)

    /* restore all the registers, if they were clobbered */
    for _, lr := range self.ctxt.regs {
        if rr := self.r(lr); rt.R_memmove[rr] || memcpyargs[rr] {
            p.MOVQ(self.ctxt.slot(lr), rr)
        }
    }
}

func (self *CodeGen) abiBlockZero(p *x86_64.Program, pd hir.PointerRegister, nb int64) {
    var dp int32
    var rd x86_64.Register64

    /* check for block size */
    if nb <= 0 || nb > math.MaxInt32 {
        panic("abiBlockZero: invalid block size")
    }

    /* use XMM for larger blocks */
    if nb >= 16 {
        p.PXOR(XMM15, XMM15)
    }

    /* use loops to reduce the code length */
    if rd = self.r(pd); nb >= 128 {
        r := x86_64.CreateLabel("loop")
        t := x86_64.CreateLabel("begin")

        /* set up the zeroing loop, use 8x loop for more efficient pipelining */
        p.MOVQ (rd, RDI)
        p.MOVL (nb / 128, EAX)
        p.JMP  (t)
        p.Link (r)
        p.ADDQ (128, RDI)
        p.Link (t)

        /* generate the zeroing instructions */
        for i := int32(0); i < 8; i++ {
            p.MOVDQU(XMM15, Ptr(RDI, i * 16))
        }

        /* decrease & check loop counter */
        p.SUBL (1, EAX)
        p.JNZ  (r)

        /* replace the register */
        rd = RDI
        nb %= 128
    }

    /* clear every 16-byte block */
    for nb >= 16 {
        p.MOVDQU(XMM15, Ptr(rd, dp))
        dp += 16
        nb -= 16
    }

    /* only 1 byte left */
    if nb == 1 {
        p.MOVB(0, Ptr(rd, dp))
        return
    }

    /* still bytes need to be zeroed */
    if nb != 0 {
        p.XORL(EAX, EAX)
    }

    /* clear every 8-byte block */
    if nb >= 8 {
        p.MOVQ(RAX, Ptr(rd, dp))
        dp += 8
        nb -= 8
    }

    /* clear every 4-byte block */
    if nb >= 8 {
        p.MOVL(EAX, Ptr(rd, dp))
        dp += 4
        nb -= 4
    }

    /* clear every 2-byte block */
    if nb >= 2 {
        p.MOVW(AX, Ptr(rd, dp))
        dp += 2
        nb -= 2
    }

    /* last byte */
    if nb > 0 {
        p.MOVB(AL, Ptr(rd, dp))
    }
}

/** Function & Method Call **/

var argumentOrder = [6]x86_64.Register64 {
    RDI,
    RSI,
    RDX,
    RCX,
    R8,
    R9,
}

var argumentRegisters = map[x86_64.Register64]bool {
    RDI : true,
    RSI : true,
    RDX : true,
    RCX : true,
    R8  : true,
    R9  : true,
}

var reservedRegisters = map[x86_64.Register64]bool {
    RBX: true,
    R12: true,
    R13: true,
    R14: true,
    R15: true,
}

func ri2reg(ri uint8) hir.Register {
    if ri & hir.ArgPointer == 0 {
        return hir.GenericRegister(ri & hir.ArgMask)
    } else {
        return hir.PointerRegister(ri & hir.ArgMask)
    }
}

func checkfp(fp unsafe.Pointer) uintptr {
    if fp == nil {
        panic("checkfp: nil function")
    } else {
        return uintptr(fp)
    }
}

func checkptr(ri uint8, arg abi.Parameter) bool {
    return arg.IsPointer() == ((ri & hir.ArgPointer) != 0)
}

func (self *CodeGen) abiCallGo(p *x86_64.Program, v *hir.Ir) {
    self.internalCallFunction(p, v, nil, func(fp *hir.CallHandle) {
        p.MOVQ(checkfp(fp.Func), R12)
        p.CALLQ(R12)
    })
}

func (self *CodeGen) abiCallNative(p *x86_64.Program, v *hir.Ir) {
    rv := hir.Register(nil)
    fp := hir.LookupCall(v.Iv)

    /* native function can have at most 1 return value */
    if v.Rn > 1 {
        panic("abiCallNative: native function can only have at most 1 return value")
    }

    /* passing arguments on stack is currently not implemented */
    if int(v.An) > len(argumentOrder) {
        panic("abiCallNative: not implemented: passing arguments on stack for native functions")
    }

    /* save all the allocated registers (except reserved registers) before function call */
    for _, lr := range self.ctxt.regs {
        if rr := self.r(lr); !reservedRegisters[rr] {
            p.MOVQ(rr, self.ctxt.slot(lr))
        }
    }

    /* load all the parameters */
    for i := 0; i < int(v.An); i++ {
        rr := ri2reg(v.Ar[i])
        rd := argumentOrder[i]

        /* check for zero source and spilled arguments */
        if rr.Z() {
            p.XORL(x86_64.Register32(rd), x86_64.Register32(rd))
        } else if rs := self.r(rr); argumentRegisters[rs] {
            p.MOVQ(self.ctxt.slot(rr), rd)
        } else {
            p.MOVQ(rs, rd)
        }
    }

    /* call the function */
    p.MOVQ(checkfp(fp.Func), RAX)
    p.CALLQ(RAX)

    /* store the result */
    if v.Rn != 0 {
        if rv = ri2reg(v.Rr[0]); !rv.Z() {
            p.MOVQ(RAX, self.r(rv))
        }
    }

    /* restore all the allocated registers (except reserved registers and result) after function call */
    for _, lr := range self.ctxt.regs {
        if rr := self.r(lr); (lr != rv) && !reservedRegisters[rr] {
            p.MOVQ(self.ctxt.slot(lr), rr)
        }
    }
}

func (self *CodeGen) abiCallMethod(p *x86_64.Program, v *hir.Ir) {
    self.internalCallFunction(p, v, v.Pd, func(fp *hir.CallHandle) {
        p.MOVQ(self.ctxt.slot(v.Ps), R12)
        p.CALLQ(Ptr(R12, int32(rt.GoItabFuncBase) + int32(fp.Slot) * abi.PtrSize))
    })
}

func (self *CodeGen) internalSetArg(p *x86_64.Program, ri uint8, arg abi.Parameter, clobberSet map[x86_64.Register64]bool) {
    if !checkptr(ri, arg) {
        panic("internalSetArg: passing arguments in different kind of registers")
    } else if !arg.InRegister {
        self.internalSetStack(p, ri2reg(ri), arg)
    } else {
        self.internalSetRegister(p, ri2reg(ri), arg, clobberSet)
    }
}

func (self *CodeGen) internalSetStack(p *x86_64.Program, rr hir.Register, arg abi.Parameter) {
    if rr.Z() {
        p.MOVQ(0, Ptr(RSP, int32(arg.Mem)))
    } else {
        p.MOVQ(self.r(rr), Ptr(RSP, int32(arg.Mem)))
    }
}

func (self *CodeGen) internalSetRegister(p *x86_64.Program, rr hir.Register, arg abi.Parameter, clobberSet map[x86_64.Register64]bool) {
    if rr.Z() {
        p.XORL(x86_64.Register32(arg.Reg), x86_64.Register32(arg.Reg))
    } else if lr := self.r(rr); clobberSet[lr] {
        p.MOVQ(self.ctxt.slot(rr), arg.Reg)
    } else if clobberSet[arg.Reg] = true; self.rindex(arg.Reg) != nil {
        p.MOVQ(self.ctxt.slot(rr), arg.Reg)
    } else {
        p.MOVQ(lr, arg.Reg)
    }
}

func (self *CodeGen) internalCallFunction(p *x86_64.Program, v *hir.Ir, this hir.Register, makeFuncCall func(fp *hir.CallHandle)) {
    ac := 0
    fp := hir.LookupCall(v.Iv)
    fv := abi.ABI.FnTab[fp.Id]
    rm := make(map[hir.Register]int32)
    cs := make(map[x86_64.Register64]bool)

    /* find the function */
    if fv == nil {
        panic(fmt.Sprintf("internalCallFunction: invalid function ID: %d", v.Iv))
    }

    /* "this" is an implicit argument, so exclude from argument count */
    if this != nil {
        ac = 1
    }

    /* check for argument and return value count */
    if int(v.Rn) != len(fv.Rets) || int(v.An) != len(fv.Args) - ac {
        panic("internalCallFunction: argument or return value count mismatch")
    }

    /* save all the allocated registers before function call */
    for _, lr := range self.ctxt.regs {
        p.MOVQ(self.r(lr), self.ctxt.slot(lr))
    }

    /* load all the arguments */
    for i, vv := range fv.Args {
        if i == 0 && this != nil {
            self.internalSetArg(p, this.A(), vv, cs)
        } else {
            self.internalSetArg(p, v.Ar[i - ac], vv, cs)
        }
    }

    /* call the function with reserved registers restored */
    self.abiLoadReserved(p)
    makeFuncCall(fp)
    self.abiSaveReserved(p)

    /* if the function returns a value with a used register, spill it on stack */
    for i, retv := range fv.Rets {
        if rr := ri2reg(v.Rr[i]); !rr.Z() {
            if !retv.InRegister {
                rm[rr] = int32(retv.Mem)
            } else if self.rindex(retv.Reg) != nil {
                p.MOVQ(retv.Reg, self.ctxt.slot(rr))
            }
        }
    }

    /* save all the non-spilled arguments */
    for i, retv := range fv.Rets {
        if rr := ri2reg(v.Rr[i]); !rr.Z() {
            if retv.InRegister && self.rindex(retv.Reg) == nil {
                rm[rr] = -1
                p.MOVQ(retv.Reg, self.r(rr))
            }
        }
    }

    /* restore all the allocated registers (except return values) after function call */
    for _, lr := range self.ctxt.regs {
        if _, ok := rm[lr]; !ok {
            p.MOVQ(self.ctxt.slot(lr), self.r(lr))
        }
    }

    /* store all the stack-based return values */
    for rr, mem := range rm {
        if mem != -1 {
            p.MOVQ(Ptr(RSP, mem), self.r(rr))
        }
    }
}
