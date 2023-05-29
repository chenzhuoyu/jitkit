package abi

import (
    `testing`

    `github.com/davecgh/go-spew/spew`
)

func TestABI_FunctionLayout(t *testing.T) {
    spew.Dump(ABI.FnTab)
}
