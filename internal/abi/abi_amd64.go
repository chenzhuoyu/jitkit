package abi

import (
    `fmt`
    `reflect`
    `sort`
    `strings`
    `unsafe`

    `github.com/chenzhuoyu/iasm/x86_64`
    `github.com/chenzhuoyu/jitkit/internal/rt`
)

const (
    PtrSize  = 8    // pointer size
    PtrAlign = 8    // pointer alignment
)

type Parameter struct {
    Mem        uintptr
    Reg        x86_64.Register64
    Type       reflect.Type
    InRegister bool
}

var (
    intType = reflect.TypeOf(0)
    ptrType = reflect.TypeOf(unsafe.Pointer(nil))
)

func mkReg(vt reflect.Type, reg x86_64.Register64) (p Parameter) {
    p.Reg = reg
    p.Type = vt
    p.InRegister = true
    return
}

func mkStack(vt reflect.Type, mem uintptr) (p Parameter) {
    p.Mem = mem
    p.Type = vt
    p.InRegister = false
    return
}

func (self Parameter) String() string {
    if self.InRegister {
        return fmt.Sprintf("%%%s", self.Reg)
    } else {
        return fmt.Sprintf("%d(%%rsp)", self.Mem)
    }
}

func (self Parameter) IsPointer() bool {
    switch self.Type.Kind() {
        case reflect.Bool          : fallthrough
        case reflect.Int           : fallthrough
        case reflect.Int8          : fallthrough
        case reflect.Int16         : fallthrough
        case reflect.Int32         : fallthrough
        case reflect.Int64         : fallthrough
        case reflect.Uint          : fallthrough
        case reflect.Uint8         : fallthrough
        case reflect.Uint16        : fallthrough
        case reflect.Uint32        : fallthrough
        case reflect.Uint64        : fallthrough
        case reflect.Uintptr       : return false
        case reflect.Chan          : fallthrough
        case reflect.Func          : fallthrough
        case reflect.Map           : fallthrough
        case reflect.Ptr           : fallthrough
        case reflect.UnsafePointer : return true
        case reflect.Float32       : fallthrough
        case reflect.Float64       : fallthrough
        case reflect.Complex64     : fallthrough
        case reflect.Complex128    : fallthrough
        case reflect.Array         : fallthrough
        case reflect.Struct        : panic("abi: unsupported types")
        default                    : panic("abi: invalid value type")
    }
}

var regOrder = [...]x86_64.Register64 {
    x86_64.RAX,
    x86_64.RBX,
    x86_64.RCX,
    x86_64.RDI,
    x86_64.RSI,
    x86_64.R8,
    x86_64.R9,
    x86_64.R10,
    x86_64.R11,
}

type _StackSlot struct {
    p bool
    m uintptr
}

type _StackAlloc struct {
    i int
    s uintptr
}

func (self *_StackAlloc) reg(vt reflect.Type) (p Parameter) {
    p = mkReg(vt, regOrder[self.i])
    self.i++
    return
}

func (self *_StackAlloc) stack(vt reflect.Type) (p Parameter) {
    p = mkStack(vt, self.s)
    self.s += vt.Size()
    return
}

func (self *_StackAlloc) spill(n uintptr, a int) uintptr {
    self.s = alignUp(self.s, a) + n
    return self.s
}

func (self *_StackAlloc) alloc(p []Parameter, vt reflect.Type) []Parameter {
    nb := vt.Size()
    vk := vt.Kind()

    /* zero-sized objects are allocated on stack */
    if nb == 0 {
        return append(p, mkStack(intType, self.s))
    }

    /* check for value type */
    switch vk {
        case reflect.Bool          : return self.valloc(p, reflect.TypeOf(false))
        case reflect.Int           : return self.valloc(p, intType)
        case reflect.Int8          : return self.valloc(p, reflect.TypeOf(int8(0)))
        case reflect.Int16         : return self.valloc(p, reflect.TypeOf(int16(0)))
        case reflect.Int32         : return self.valloc(p, reflect.TypeOf(int32(0)))
        case reflect.Int64         : return self.valloc(p, reflect.TypeOf(int64(0)))
        case reflect.Uint          : return self.valloc(p, reflect.TypeOf(uint(0)))
        case reflect.Uint8         : return self.valloc(p, reflect.TypeOf(uint8(0)))
        case reflect.Uint16        : return self.valloc(p, reflect.TypeOf(uint16(0)))
        case reflect.Uint32        : return self.valloc(p, reflect.TypeOf(uint32(0)))
        case reflect.Uint64        : return self.valloc(p, reflect.TypeOf(uint64(0)))
        case reflect.Uintptr       : return self.valloc(p, reflect.TypeOf(uintptr(0)))
        case reflect.Float32       : panic("abi: go117: not implemented: float32")
        case reflect.Float64       : panic("abi: go117: not implemented: float64")
        case reflect.Complex64     : panic("abi: go117: not implemented: complex64")
        case reflect.Complex128    : panic("abi: go117: not implemented: complex128")
        case reflect.Array         : panic("abi: go117: not implemented: arrays")
        case reflect.Chan          : return self.valloc(p, reflect.TypeOf((chan int)(nil)))
        case reflect.Func          : return self.valloc(p, reflect.TypeOf((func())(nil)))
        case reflect.Map           : return self.valloc(p, reflect.TypeOf((map[int]int)(nil)))
        case reflect.Ptr           : return self.valloc(p, reflect.TypeOf((*int)(nil)))
        case reflect.UnsafePointer : return self.valloc(p, ptrType)
        case reflect.Interface     : return self.valloc(p, ptrType, ptrType)
        case reflect.Slice         : return self.valloc(p, ptrType, intType, intType)
        case reflect.String        : return self.valloc(p, ptrType, intType)
        case reflect.Struct        : panic("abi: go117: not implemented: structs")
        default                    : panic("abi: invalid value type")
    }
}

func (self *_StackAlloc) ralloc(p []Parameter, vts ...reflect.Type) []Parameter {
    for _, vt := range vts { p = append(p, self.reg(vt)) }
    return p
}

func (self *_StackAlloc) salloc(p []Parameter, vts ...reflect.Type) []Parameter {
    for _, vt := range vts { p = append(p, self.stack(vt)) }
    return p
}

func (self *_StackAlloc) valloc(p []Parameter, vts ...reflect.Type) []Parameter {
    if self.i + len(vts) <= len(regOrder) {
        return self.ralloc(p, vts...)
    } else {
        return self.salloc(p, vts...)
    }
}

type FunctionLayout struct {
    Id   int
    Sp   uintptr
    Args []Parameter
    Rets []Parameter
}

func (self *FunctionLayout) String() string {
    if self.Id < 0 {
        return fmt.Sprintf("{func,%s}", self.formatFn())
    } else {
        return fmt.Sprintf("{meth/%d,%s}", self.Id, self.formatFn())
    }
}

func (self *FunctionLayout) StackMap() *rt.StackMap {
    var st []_StackSlot
    var mb rt.StackMapBuilder

    /* add arguments */
    for _, v := range self.Args {
        st = append(st, _StackSlot {
            m: v.Mem,
            p: v.IsPointer(),
        })
    }

    /* add stack-passed return values */
    for _, v := range self.Rets {
        if !v.InRegister {
            st = append(st, _StackSlot {
                m: v.Mem,
                p: v.IsPointer(),
            })
        }
    }

    /* sort by memory offset */
    sort.Slice(st, func(i int, j int) bool {
        return st[i].m < st[j].m
    })

    /* add the bits */
    for _, v := range st {
        mb.AddField(v.p)
    }

    /* build the stack map */
    return mb.Build()
}

func (self *FunctionLayout) formatFn() string {
    return fmt.Sprintf("$%#x,(%s),(%s)", self.Sp, self.formatSeq(self.Args), self.formatSeq(self.Rets))
}

func (self *FunctionLayout) formatSeq(v []Parameter) string {
    nb := len(v)
    mm := make([]string, len(v))

    /* convert each part */
    for i := 0; i < nb; i++ {
        mm[i] = v[i].String()
    }

    /* join them together */
    return strings.Join(mm, ",")
}

type AMD64ABI struct {
    FnTab map[int]*FunctionLayout
}

func ArchCreateABI() *AMD64ABI {
    return &AMD64ABI {
        FnTab: make(map[int]*FunctionLayout),
    }
}

func (self *AMD64ABI) Reserved() map[x86_64.Register64]int32 {
    return map[x86_64.Register64]int32 {
        x86_64.R14: 0, // current goroutine
        x86_64.R15: 1, // GOT reference
    }
}

func (self *AMD64ABI) LayoutFunc(id int, ft reflect.Type) *FunctionLayout {
    var sa _StackAlloc
    var fn FunctionLayout

    /* allocate the receiver if any (interface call always uses pointer) */
    if id >= 0 {
        fn.Args = sa.alloc(fn.Args, ptrType)
    }

    /* assign every argument */
    for i := 0; i < ft.NumIn(); i++ {
        fn.Args = sa.alloc(fn.Args, ft.In(i))
    }

    /* reset the register counter, and add a pointer alignment field */
    sa.i = 0
    sa.spill(0, PtrAlign)

    /* assign every return value */
    for i := 0; i < ft.NumOut(); i++ {
        fn.Rets = sa.alloc(fn.Rets, ft.Out(i))
    }

    /* assign spill slots */
    for i := 0; i < len(fn.Args); i++ {
        if fn.Args[i].InRegister {
            fn.Args[i].Mem = sa.spill(PtrSize, PtrAlign) - PtrSize
        }
    }

    /* add the final pointer alignment field */
    fn.Id = id
    fn.Sp = sa.spill(0, PtrAlign)
    return &fn
}

func (self *AMD64ABI) RegisterMethod(id int, mt rt.Method) int {
    self.FnTab[id] = self.LayoutFunc(mt.Id, mt.Vt.Pack().Method(mt.Id).Type)
    return mt.Id
}

func (self *AMD64ABI) RegisterFunction(id int, fn interface{}) (fp unsafe.Pointer) {
    vv := rt.UnpackEface(fn)
    vt := vv.Type.Pack()

    /* must be a function */
    if vt.Kind() != reflect.Func {
        panic("fn is not a function")
    }

    /* layout the function, and get the real function address */
    self.FnTab[id] = self.LayoutFunc(-1, vt)
    return *(*unsafe.Pointer)(vv.Value)
}
