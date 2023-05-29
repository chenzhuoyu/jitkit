package hir

import (
    `fmt`
    `strings`
)

type Program struct {
    Head *Ir
}

func (self Program) Free() {
    for p, q := self.Head, self.Head; p != nil; p = q {
        q = p.Ln
        p.Free()
    }
}

func (self Program) Disassemble() string {
    ret := make([]string, 0, 64)
    ref := make(map[*Ir]string)

    /* scan all the branch target */
    for p := self.Head; p != nil; p = p.Ln {
        if p.IsBranch() {
            if p.Op != OP_bsw {
                if _, ok := ref[p.Br]; !ok {
                    ref[p.Br] = fmt.Sprintf("L_%d", len(ref))
                }
            } else {
                for _, lb := range p.Switch() {
                    if lb != nil {
                        if _, ok := ref[lb]; !ok {
                            ref[lb] = fmt.Sprintf("L_%d", len(ref))
                        }
                    }
                }
            }
        }
    }

    /* dump all the instructions */
    for p := self.Head; p != nil; p = p.Ln {
        var ok bool
        var vv string

        /* check for label reference */
        if vv, ok = ref[p]; ok {
            ret = append(ret, vv + ":")
        }

        /* indent each line */
        for _, ln := range strings.Split(p.Disassemble(ref), "\n") {
            ret = append(ret, "    " + ln)
        }
    }

    /* join them together */
    return strings.Join(ret, "\n")
}
