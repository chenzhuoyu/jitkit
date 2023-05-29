package ssa

import (
    `fmt`
    `math`
)

const (
    _P_term = math.MaxUint32
)

type Pos struct {
    B *BasicBlock
    I int
}

func pos(bb *BasicBlock, i int) Pos {
    return Pos { bb, i }
}

func (self Pos) String() string {
    if self.I == _P_term {
        return fmt.Sprintf("bb_%d.term", self.B.Id)
    } else {
        return fmt.Sprintf("bb_%d.ins[%d]", self.B.Id, self.I)
    }
}

func (self Pos) isPriorTo(other Pos) bool {
    return self.B.Id < other.B.Id || (self.I < other.I && self.B.Id == other.B.Id)
}
