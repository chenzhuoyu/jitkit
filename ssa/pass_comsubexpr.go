package ssa

import (
    `fmt`
)

type _Vid interface {
    IrDefinitions
    vid() []string
}

func (self *IrLoadArg) vid() []string {
    return []string {
        fmt.Sprintf("#%d", self.I),
    }
}

func (self *IrConstInt) vid() []string {
    return []string {
        fmt.Sprintf("$%d", self.V),
    }
}

func (self *IrConstPtr) vid() []string {
    return []string {
        fmt.Sprintf("$%p", self.P),
    }
}

func (self *IrLEA) vid() []string {
    return []string {
        fmt.Sprintf("(& %s %s)", self.Mem, self.Off),
    }
}

func (self *IrUnaryExpr) vid() []string {
    return []string {
        fmt.Sprintf("(%s %s)", self.Op, self.V),
    }
}

func (self *IrBinaryExpr) vid() []string {
    x := self.X
    y := self.Y

    /* commutative operations, sort the operands */
    switch self.Op {
        case IrOpAdd : fallthrough
        case IrOpMul : fallthrough
        case IrOpAnd : fallthrough
        case IrOpOr  : fallthrough
        case IrOpXor : fallthrough
        case IrCmpEq : fallthrough
        case IrCmpNe : if x > y { x, y = y, x }
    }

    /* build the value ID */
    return []string {
        fmt.Sprintf("(%s %s %s)", self.Op, x, y),
    }
}

func (self *IrBitTestSet) vid() []string {
    return []string {
        fmt.Sprintf("(&# %s %s)", self.X, self.Y),
        fmt.Sprintf("(|# %s %s)", self.X, self.Y),
    }
}

type _VidMap struct {
    p *_VidMap
    m map[string]Reg
}

func (self *_VidMap) derive() *_VidMap {
    return &_VidMap {
        p: self,
        m: make(map[string]Reg),
    }
}

func (self *_VidMap) lookup(vid string) (Reg, bool) {
    if r, ok := self.m[vid]; ok {
        return r, true
    } else if self.p == nil {
        return 0, false
    } else {
        return self.p.lookup(vid)
    }
}

func (self *_VidMap) define(vid string, reg Reg) {
    self.m[vid] = reg
}

// CSE performs the Common Sub-expression Elimintation optimization.
type CSE struct{}

func (self CSE) dfs(cfg *CFG, bb *BasicBlock, vm *_VidMap) {
    ins := bb.Ins
    vals := vm.derive()

    /* scan every instructions */
    for i, v := range ins {
        var r Reg
        var d _Vid
        var ok bool

        /* check if the instruction have VIDs */
        if d, ok = v.(_Vid); !ok {
            continue
        }

        /* calculate the VIDs */
        repc := i
        vids := d.vid()
        defs := d.Definitions()

        /* replace each VID with a copy instruction */
        for j, vid := range vids {
            s := defs[j]
            r, ok = vals.lookup(vid)

            /* skip zero registers */
            if s.Kind() == K_zero {
                continue
            }

            /* add to definations if not found */
            if !ok {
                vals.define(vid, *s)
                continue
            }

            /* allocate one slot for the new instruction */
            repc++
            bb.Ins = append(bb.Ins, nil)
            copy(bb.Ins[repc + 1:], bb.Ins[repc:])

            /* insert a new copy instruction */
            bb.Ins[repc] = IrCopy(*s, r)
            *s = s.Zero()
        }

        /* all the definations are been replaced */
        if repc == i + len(defs) {
            bb.Ins = append(bb.Ins[:i], bb.Ins[i + 1:]...)
        }
    }

    /* DFS the dominator tree */
    for _, v := range cfg.DominatorOf[bb.Id] {
        self.dfs(cfg, v, vals)
    }
}

func (self CSE) Apply(cfg *CFG) {
    self.dfs(cfg, cfg.Root, nil)
}
