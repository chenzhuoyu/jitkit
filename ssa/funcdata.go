package ssa

import (
    `fmt`
    `strings`

    `github.com/chenzhuoyu/jitkit/internal/rt`
)

type FuncData struct {
    Code     []byte
    Layout   *FuncLayout
    Liveness map[Pos]SlotSet
    StackMap map[uintptr]*rt.StackMap
}

type FuncLayout struct {
    Ins   []IrNode
    Start map[int]int
    Block map[int]*BasicBlock
}

func (self *FuncLayout) String() string {
    ni := len(self.Ins)
    ns := len(self.Start)
    ss := make([]string, 0, ni + ns)

    /* print every instruction */
    for i, ins := range self.Ins {
        if bb, ok := self.Block[i]; !ok {
            ss = append(ss, fmt.Sprintf("%06x |     %s", i, ins))
        } else {
            ss = append(ss, fmt.Sprintf("%06x | bb_%d:", i, bb.Id), fmt.Sprintf("%06x |     %s", i, ins))
        }
    }

    /* join them together */
    return fmt.Sprintf(
        "FuncLayout {\n%s\n}",
        strings.Join(ss, "\n"),
    )
}
