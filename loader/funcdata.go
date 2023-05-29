package loader

import (
    `sync`
    _ `unsafe`
)

const (
    _PCDATA_UnsafePoint       = 0
    _PCDATA_StackMapIndex     = 1
    _PCDATA_UnsafePointUnsafe = -2
)

//go:linkname lastmoduledatap runtime.lastmoduledatap
//goland:noinspection GoUnusedGlobalVariable
var lastmoduledatap *_ModuleData

//go:linkname moduledataverify1 runtime.moduledataverify1
func moduledataverify1(_ *_ModuleData)

var (
    modLock sync.Mutex
    modList []*_ModuleData
)

func toZigzag(v int) int {
    return (v << 1) ^ (v >> 31)
}

func encodeFirst(v int) []byte {
    return encodeValue(v + 1)
}

func encodeValue(v int) []byte {
    return encodeVariant(toZigzag(v))
}

func encodeVariant(v int) []byte {
    var u int
    var r []byte

    /* split every 7 bits */
    for v > 127 {
        u = v & 0x7f
        v = v >> 7
        r = append(r, byte(u) | 0x80)
    }

    /* check for last one */
    if v == 0 {
        return r
    }

    /* add the last one */
    r = append(r, byte(v))
    return r
}

func registerModule(mod *_ModuleData) {
    modLock.Lock()
    modList = append(modList, mod)
    lastmoduledatap.next = mod
    lastmoduledatap = mod
    modLock.Unlock()
}
