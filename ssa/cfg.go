package ssa

import (
    `sync/atomic`

    `github.com/chenzhuoyu/jitkit/internal/abi`
)

type _CFGPrivate struct {
    reg   uint64
    block uint64
}

func (self *_CFGPrivate) allocreg() int {
    return int(atomic.AddUint64(&self.reg, 1)) - 1
}

func (self *_CFGPrivate) allocblock() int {
    return int(atomic.AddUint64(&self.block, 1)) - 1
}

type CFG struct {
    _CFGPrivate
    Func              FuncData
    Root              *BasicBlock
    Depth             map[int]int
    Layout            *abi.FunctionLayout
    DominatedBy       map[int]*BasicBlock
    DominatorOf       map[int][]*BasicBlock
    DominanceFrontier map[int][]*BasicBlock
}

func (self *CFG) Rebuild() {
    updateDominatorTree(self)
    updateDominatorDepth(self)
    updateDominatorFrontier(self)
}

func (self *CFG) MaxBlock() int {
    return int(self.block)
}

func (self *CFG) PostOrder() *BasicBlockIter {
    return newBasicBlockIter(self)
}

func (self *CFG) CreateBlock() (r *BasicBlock) {
    r = new(BasicBlock)
    r.Id = self.allocblock()
    return
}

func (self *CFG) CreateRegister(ptr bool) Reg {
    if i := self.allocreg(); ptr {
        return mkreg(1, K_norm, 0).Derive(i)
    } else {
        return mkreg(0, K_norm, 0).Derive(i)
    }
}

func (self *CFG) CreateUnreachable(bb *BasicBlock) (ret *BasicBlock) {
    ret      = self.CreateBlock()
    ret.Ins  = []IrNode { new(IrBreakpoint) }
    ret.Term = &IrSwitch { Ln: IrLikely(ret) }
    ret.Pred = []*BasicBlock { bb, ret }
    return
}
