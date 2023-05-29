package ssa

// Reduce combines CSE, PhiElim, CopyElim and TDCE.
type Reduce struct{}

var Reductions = []PassDescriptor {
    { Name: "Common Sub-expression Elimination" , Pass: new(CSE)      },
    { Name: "Phi Elimination"                   , Pass: new(PhiElim)  },
    { Name: "Copy Elimination"                  , Pass: new(CopyElim) },
    { Name: "Trivial Dead Code Elimination"     , Pass: new(TDCE)     },
}

func (Reduce) Apply(cfg *CFG) {
    for _, r := range Reductions {
        r.Pass.Apply(cfg)
    }
}
