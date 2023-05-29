package ssa

import (
    `fmt`
    `sort`
    `strings`
)

type (
    SlotSet map[IrSpillSlot]struct{}
)

func (self SlotSet) add(r IrSpillSlot) bool {
    if _, ok := self[r]; ok {
        return false
    } else {
        self[r] = struct{}{}
        return true
    }
}

func (self SlotSet) clone() (rs SlotSet) {
    rs = make(SlotSet, len(self))
    for r := range self { rs.add(r) }
    return
}

func (self SlotSet) remove(r IrSpillSlot) bool {
    if _, ok := self[r]; !ok {
        return false
    } else {
        delete(self, r)
        return true
    }
}

func (self SlotSet) String() string {
    nb := len(self)
    rs := make([]string, 0, nb)
    rr := make([]IrSpillSlot, 0, nb)

    /* extract all slot */
    for r := range self {
        rr = append(rr, r)
    }

    /* sort by slot ID */
    sort.Slice(rr, func(i int, j int) bool {
        return rr[i] < rr[j]
    })

    /* convert every slot */
    for _, r := range rr {
        rs = append(rs, r.String())
    }

    /* join them together */
    return fmt.Sprintf(
        "{%s}",
        strings.Join(rs, ", "),
    )
}
