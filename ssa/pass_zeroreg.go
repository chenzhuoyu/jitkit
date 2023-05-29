package ssa

import (
    `runtime`
)

// ZeroReg replaces read to %z or %nil to a register that was
// initialized to zero for architectures that does not have constant
// zero registers, such as `x86_64`.
type ZeroReg struct{}

func (ZeroReg) replace(cfg *CFG) {
    cfg.PostOrder().ForEach(func(bb *BasicBlock) {
        var ok bool
        var rr *Reg
        var use IrUsages

        /* create the instruction buffer */
        ins := bb.Ins
        bb.Ins = make([]IrNode, 0, len(ins))

        /* zero register replacer */
        replacez := func(v IrUsages, ins *[]IrNode, tail IrNode) {
            var z Reg
            var r *Reg

            /* insert a zeroing instruction if needed */
            for _, r = range v.Usages() {
                if r.Kind() == K_zero {
                    z = cfg.CreateRegister(false)
                    *ins = append(*ins, IrArchZero(z))
                    break
                }
            }

            /* substitute all the zero register usages */
            for _, r = range v.Usages() {
                if r.Kind() == K_zero {
                    *r = z
                }
            }

            /* add the instruction if needed */
            if tail != nil {
                *ins = append(*ins, tail)
            }
        }

        /* scan all the Phi nodes */
        for _, p := range bb.Pred {
            var z Reg
            var v *IrPhi

            /* insert a zeroing instruction to its predecessor if needed */
            for _, v = range bb.Phi {
                if v.V[p].Kind() == K_zero {
                    z = cfg.CreateRegister(false)
                    p.Ins = append(p.Ins, IrArchZero(z))
                    break
                }
            }

            /* substitute all the zero register usages */
            for _, v = range bb.Phi {
                if rr = v.V[p]; rr.Kind() == K_zero {
                    *rr = z
                }
            }
        }

        /* scan all the instructions */
        for _, v := range ins {
            if use, ok = v.(IrUsages); ok {
                replacez(use, &bb.Ins, v)
            } else {
                bb.Ins = append(bb.Ins, v)
            }
        }

        /* scan the terminator */
        if use, ok = bb.Term.(IrUsages); ok {
            replacez(use, &bb.Ins, nil)
        }
    })
}

//goland:noinspection GoBoolExpressions
func (self ZeroReg) Apply(cfg *CFG) {
    if runtime.GOARCH == "amd64" {
        self.replace(cfg)
    }
}
