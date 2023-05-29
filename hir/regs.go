package hir

import (
    `fmt`
)

type Register interface {
    fmt.Stringer
    Z() bool
    A() uint8
}

type (
    GenericRegister uint8
    PointerRegister uint8
)

const (
    ArgMask    = 0x7f
    ArgGeneric = 0x00
    ArgPointer = 0x80
)

const (
    R0 GenericRegister = iota
    R1
    R2
    R3
    R4
    Rz
)

const (
    P0 PointerRegister = iota
    P1
    P2
    P3
    P4
    P5
    Pn
)

var GenericRegisters = map[GenericRegister]string {
    R0: "r0",
    R1: "r1",
    R2: "r2",
    R3: "r3",
    R4: "r4",
    Rz: "z",
}

var PointerRegisters = map[PointerRegister]string {
    P0: "p0",
    P1: "p1",
    P2: "p2",
    P3: "p3",
    P4: "p4",
    P5: "p5",
    Pn: "nil",
}

func (self GenericRegister) Z() bool { return self == Rz }
func (self PointerRegister) Z() bool { return self == Pn }

func (self GenericRegister) A() uint8 { return uint8(self) | ArgGeneric }
func (self PointerRegister) A() uint8 { return uint8(self) | ArgPointer }

func (self GenericRegister) String() string {
    if v := GenericRegisters[self]; v == "" {
        panic(fmt.Sprintf("invalid generic register: 0x%02x", uint8(self)))
    } else {
        return v
    }
}

func (self PointerRegister) String() string {
    if v := PointerRegisters[self]; v == "" {
        panic(fmt.Sprintf("invalid pointer register: 0x%02x", uint8(self)))
    } else {
        return v
    }
}
