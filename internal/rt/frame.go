package rt

type Stack struct {
    Sp uintptr
    Nb uintptr
}

type Frame struct {
    SpTab     []Stack
    ArgSize   uintptr
    ArgPtrs   *StackMap
    LocalPtrs *StackMap
}
