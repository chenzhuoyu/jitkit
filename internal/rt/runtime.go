package rt

import (
    `reflect`
    `unsafe`
)

const (
    kindMask = (1 << 5) - 1
)

var (
    reflectTypeItab = findReflectTypeItab()
)

type GoType struct {
    size    uintptr
    ptrdata uintptr
    hash    uint32
    tflag   uint8
    align   uint8
    falign  uint8
    kind    uint8
    equal   func(unsafe.Pointer, unsafe.Pointer) bool
    gcdata  *byte
    str     int32
    ptr     int32
}

func (self *GoType) Kind() reflect.Kind {
    return reflect.Kind(self.kind & kindMask)
}

func (self *GoType) Pack() (t reflect.Type) {
    (*GoIface)(unsafe.Pointer(&t)).Itab = reflectTypeItab
    (*GoIface)(unsafe.Pointer(&t)).Value = unsafe.Pointer(self)
    return
}

func (self *GoType) IsPtr() bool {
    return self.Kind() == reflect.Ptr || self.Kind() == reflect.UnsafePointer
}

func (self *GoType) String() string {
    return self.Pack().String()
}

type GoItab struct {
    intf unsafe.Pointer
    typ  *GoType
    hash uint32
    _    [4]byte
    ftab [1]uintptr
}

const (
    GoItabFuncBase = unsafe.Offsetof(GoItab{}.ftab)
)

type GoIface struct {
    Itab  *GoItab
    Value unsafe.Pointer
}

type GoEface struct {
    Type  *GoType
    Value unsafe.Pointer
}

type GoSlice struct {
    Ptr unsafe.Pointer
    Len int
    Cap int
}

//go:noescape
//go:linkname mapclear runtime.mapclear
//goland:noinspection GoUnusedParameter
func mapclear(t *GoType, h unsafe.Pointer)

func MapClear(m interface{}) {
    v := UnpackEface(m)
    mapclear(v.Type, v.Value)
}

func FuncAddr(f interface{}) unsafe.Pointer {
    if vv := UnpackEface(f); vv.Type.Kind() != reflect.Func {
        panic("f is not a function")
    } else {
        return *(*unsafe.Pointer)(vv.Value)
    }
}

func BytesFrom(p unsafe.Pointer, n int, c int) (r []byte) {
    (*GoSlice)(unsafe.Pointer(&r)).Ptr = p
    (*GoSlice)(unsafe.Pointer(&r)).Len = n
    (*GoSlice)(unsafe.Pointer(&r)).Cap = c
    return
}

func UnpackType(t reflect.Type) *GoType {
    return (*GoType)((*GoIface)(unsafe.Pointer(&t)).Value)
}

func UnpackEface(v interface{}) GoEface {
    return *(*GoEface)(unsafe.Pointer(&v))
}

func findReflectTypeItab() *GoItab {
    v := reflect.TypeOf(struct{}{})
    return (*GoIface)(unsafe.Pointer(&v)).Itab
}
