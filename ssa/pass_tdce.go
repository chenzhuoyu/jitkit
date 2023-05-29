package ssa

// TDCE removes trivial dead-code such as unused register definations from CFG.
type TDCE struct{}

func (TDCE) Apply(cfg *CFG) {
    for {
        done := true
        decl := make(map[Reg]struct{})

        /* Phase 1: Mark all the definations */
        cfg.PostOrder().ForEach(func(bb *BasicBlock) {
            var ok bool
            var defs IrDefinitions

            /* mark all definations in Phi nodes */
            for _, v := range bb.Phi {
                for _, r := range v.Definitions() {
                    decl[*r] = struct{}{}
                }
            }

            /* mark all definations in instructions if any */
            for _, v := range bb.Ins {
                if defs, ok = v.(IrDefinitions); ok {
                    for _, r := range defs.Definitions() {
                        decl[*r] = struct{}{}
                    }
                }
            }

            /* mark all definations in terminators if any */
            if defs, ok = bb.Term.(IrDefinitions); ok {
                for _, r := range defs.Definitions() {
                    decl[*r] = struct{}{}
                }
            }
        })

        /* Phase 2: Find all register usages */
        cfg.PostOrder().ForEach(func(bb *BasicBlock) {
            var ok bool
            var use IrUsages

            /* mark all usages in Phi nodes */
            for _, v := range bb.Phi {
                for _, r := range v.Usages() {
                    delete(decl, *r)
                }
            }

            /* mark all usages in instructions if any */
            for _, v := range bb.Ins {
                if use, ok = v.(IrUsages); ok {
                    for _, r := range use.Usages() {
                        delete(decl, *r)
                    }
                }
            }

            /* mark usages in the terminator if any */
            if use, ok = bb.Term.(IrUsages); ok {
                for _, r := range use.Usages() {
                    delete(decl, *r)
                }
            }
        })

        /* Phase 3: Remove all unused declarations */
        cfg.PostOrder().ForEach(func(bb *BasicBlock) {
            var ok bool
            var defs IrDefinitions

            /* replace unused Phi assigments with zero registers */
            for _, v := range bb.Phi {
                for _, r := range v.Definitions() {
                    if _, ok = decl[*r]; ok && r.Kind() != K_zero {
                        *r, done = r.Zero(), false
                    }
                }
            }

            /* replace unused instruction assigments with zero registers */
            for _, v := range bb.Ins {
                if defs, ok = v.(IrDefinitions); ok {
                    for _, r := range defs.Definitions() {
                        if _, ok = decl[*r]; ok && r.Kind() != K_zero {
                            *r, done = r.Zero(), false
                        }
                    }
                }
            }

            /* replace unused terminator assigments with zero registers */
            if defs, ok = bb.Term.(IrDefinitions); ok {
                for _, r := range defs.Definitions() {
                    if _, ok = decl[*r]; ok && r.Kind() != K_zero {
                        *r, done = r.Zero(), false
                    }
                }
            }
        })

        /* Phase 4: Remove the entire defination if it's all zeros */
        cfg.PostOrder().ForEach(func(bb *BasicBlock) {
            phi, ins := bb.Phi, bb.Ins
            bb.Phi, bb.Ins = bb.Phi[:0], bb.Ins[:0]

            /* remove Phi nodes that don't have any effects */
            for _, v := range phi {
                for _, r := range v.Definitions() {
                    if r.Kind() != K_zero {
                        bb.Phi = append(bb.Phi, v)
                        break
                    }
                }
            }

            /* remove instructions that don't have any effects */
            for _, v := range ins {
                if _, ok := v.(IrImpure); ok {
                    bb.Ins = append(bb.Ins, v)
                } else if d, ok := v.(IrDefinitions); !ok {
                    bb.Ins = append(bb.Ins, v)
                } else {
                    for _, r := range d.Definitions() {
                        if r.Kind() != K_zero {
                            bb.Ins = append(bb.Ins, v)
                            break
                        }
                    }
                }
            }
        })

        /* no more modifications */
        if done {
            break
        }
    }
}
