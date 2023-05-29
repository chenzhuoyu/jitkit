package ssa

import (
    `reflect`

    `github.com/chenzhuoyu/jitkit/hir`
    `github.com/chenzhuoyu/jitkit/internal/abi`
)

type Pass interface {
    Apply(*CFG)
}

type PassDescriptor struct {
    Pass Pass
    Name string
}

var Passes = []PassDescriptor {
    { Name: "Early Constant Propagation" , Pass: new(ConstProp)     },
    { Name: "Early Reduction"            , Pass: new(Reduce)        },
    { Name: "Branch Elimination"         , Pass: new(BranchElim)    },
    { Name: "Return Spreading"           , Pass: new(ReturnSpread)  },
    { Name: "Value Reordering"           , Pass: new(Reorder)       },
    { Name: "Late Constant Propagation"  , Pass: new(ConstProp)     },
    { Name: "Late Reduction"             , Pass: new(Reduce)        },
    { Name: "Machine Dependent Lowering" , Pass: new(Lowering)      },
    { Name: "Zero Register Substitution" , Pass: new(ZeroReg)       },
    { Name: "Write Barrier Insertion"    , Pass: new(WriteBarrier)  },
    { Name: "ABI-Specific Lowering"      , Pass: new(ABILowering)   },
    { Name: "Instruction Fusion"         , Pass: new(Fusion)        },
    { Name: "Instruction Compaction"     , Pass: new(Compaction)    },
    { Name: "Block Merging"              , Pass: new(BlockMerge)    },
    { Name: "Critical Edge Splitting"    , Pass: new(SplitCritical) },
    { Name: "Phi Propagation"            , Pass: new(PhiProp)       },
    { Name: "Operand Allocation"         , Pass: new(OperandAlloc)  },
    { Name: "Constant Rematerialize"     , Pass: new(Rematerialize) },
    { Name: "Pre-allocation TDCE"        , Pass: new(TDCE)          },
    { Name: "Register Allocation"        , Pass: new(RegAlloc)      },
    { Name: "Stack Liveness Analysis"    , Pass: new(StackLiveness) },
    { Name: "Function Layout"            , Pass: new(Layout)        },
}

func toFuncType(fn interface{}) reflect.Type {
    if vt := reflect.TypeOf(fn); vt.Kind() != reflect.Func {
        panic("ssa: fn must be a function prototype")
    } else {
        return vt
    }
}

func executeSSAPasses(cfg *CFG) {
    for _, p := range Passes {
        p.Pass.Apply(cfg)
    }
}

func Compile(p hir.Program, fn interface{}) (cfg *CFG) {
    cfg = newGraphBuilder().build(p)
    cfg.Layout = abi.ABI.LayoutFunc(-1, toFuncType(fn))
    insertPhiNodes(cfg)
    renameRegisters(cfg)
    executeSSAPasses(cfg)
    return
}
