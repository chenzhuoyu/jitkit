package ssa

// BlockMerge merges redundant intermediate blocks (blocks with a single
// outgoing edge which goes to another block with a single incoming edge).
type BlockMerge struct{}

func (BlockMerge) Apply(cfg *CFG) {
    for {
        var rt bool
        var nx *BasicBlock
        var it IrSuccessors
        var tr IrTerminator

        /* check every block */
        cfg.PostOrder().ForEach(func(bb *BasicBlock) {
            tr = bb.Term
            it = tr.Successors()

            /* it must have successors */
            if !it.Next() {
                return
            }

            /* it must have exactly 1 successor, and the successor must have exactly 1 predecessor */
            if nx = it.Block(); it.Next() || len(nx.Pred) != 1 {
                return
            }

            /* merge the two blocks */
            rt = true
            bb.Ins = append(bb.Ins, nx.Ins...)
            bb.Term = nx.Term

            /* must not have Phi nodes */
            if len(nx.Phi) != 0 {
                panic("invalid Phi node found in intermediate blocks")
            }

            /* get the successor iterator */
            tr = nx.Term
            it = tr.Successors()

            /* update all predecessors references */
            for it.Next() {
                rb := it.Block()
                pp := rb.Pred

                /* update predecessor list */
                for i, p := range pp {
                    if p == nx {
                        pp[i] = bb
                    }
                }

                /* update in Phi nodes */
                for _, v := range rb.Phi {
                    v.V[bb] = v.V[nx]
                    delete(v.V, nx)
                }
            }
        })

        /* rebuild the dominator tree, and retry if needed */
        if cfg.Rebuild(); !rt {
            break
        }
    }
}
