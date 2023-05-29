package rt

import (
    `fmt`
    `reflect`
    `testing`
    `unsafe`
)

const (
    _FUNCDATA_ArgsPointerMaps = 0
    _FUNCDATA_LocalsPointerMaps = 1
)

type funcInfo struct {
    fn    unsafe.Pointer
    datap unsafe.Pointer
}

type bitvector struct {
    n        int32 // # of bits
    bytedata *uint8
}

//go:nosplit
func addb(p *byte, n uintptr) *byte {
    return (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + n))
}

func (bv *bitvector) ptrbit(i uintptr) uint8 {
    b := *(addb(bv.bytedata, i/8))
    return (b >> (i % 8)) & 1
}

//go:linkname findfunc runtime.findfunc
func findfunc(_ uintptr) funcInfo

//go:linkname funcdata runtime.funcdata
func funcdata(_ funcInfo, _ uint8) unsafe.Pointer

//go:linkname stackmapdata runtime.stackmapdata
func stackmapdata(_ *StackMap, _ int32) bitvector

func stackMap(f interface{}) (*StackMap, *StackMap) {
    fv := reflect.ValueOf(f)
    if fv.Kind() != reflect.Func {
        panic("f must be reflect.Func kind!")
    }
    fi := findfunc(fv.Pointer())
    args := funcdata(fi, uint8(_FUNCDATA_ArgsPointerMaps))
    locals := funcdata(fi, uint8(_FUNCDATA_LocalsPointerMaps))
    return (*StackMap)(args), (*StackMap)(locals)
}

func dumpstackmap(m *StackMap) {
    for i := int32(0); i < m.N; i++ {
        fmt.Printf("bitmap #%d/%d: ", i, m.L)
        bv := stackmapdata(m, i)
        for j := int32(0); j < bv.n; j++ {
            fmt.Printf("%d ", bv.ptrbit(uintptr(j)))
        }
        fmt.Printf("\n")
    }
}

var keepalive struct {
    s  string
    i  int
    vp unsafe.Pointer
    sb interface{}
    fv uint64
}

func stackmaptestfunc(s string, i int, vp unsafe.Pointer, sb interface{}, fv uint64) (x *uint64, y string, z *int) {
    z = new(int)
    x = new(uint64)
    y = s + "asdf"
    keepalive.s = s
    keepalive.i = i
    keepalive.vp = vp
    keepalive.sb = sb
    keepalive.fv = fv
    return
}

func TestStackMap_Dump(t *testing.T) {
    args, locals := stackMap(stackmaptestfunc)
    println("--- args ---")
    dumpstackmap(args)
    println("--- locals ---")
    dumpstackmap(locals)
}
