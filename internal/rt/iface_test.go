package rt

import (
    `testing`
)

type FooIface interface {
    FooMethod()
}

type FooValue struct {
    X int
}

func (self FooValue) FooMethod() {
    println(self.X)
}

type FooPointer struct {
    X int
}

func (self *FooPointer) FooMethod() {
    println(self.X)
}

func TestIface_Invoke(t *testing.T) {
    var v FooIface
    v = FooValue{X: 100}
    v.FooMethod()
    v = &FooPointer{X: 200}
    v.FooMethod()
}
