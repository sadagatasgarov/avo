package operand

import (
	"fmt"

	"github.com/mmcloughlin/avo/reg"
)

type Op interface {
	Asm() string
}

type Symbol struct {
	Name   string
	Static bool
}

func (s Symbol) String() string {
	n := s.Name
	if s.Static {
		n += "<>"
	}
	return n
}

type Mem struct {
	Symbol Symbol
	Disp   int
	Base   reg.Register
	Index  reg.Register
	Scale  uint8
}

// NewParamAddr is a convenience to build a Mem operand pointing to a function
// parameter, which is a named offset from the frame pointer pseudo register.
func NewParamAddr(name string, offset int) Mem {
	return Mem{
		Symbol: Symbol{
			Name:   name,
			Static: false,
		},
		Disp: offset,
		Base: reg.FramePointer,
	}
}

func (m Mem) Asm() string {
	a := m.Symbol.String()
	if m.Disp != 0 {
		if a == "" {
			a += fmt.Sprintf("%d", m.Disp)
		} else {
			a += fmt.Sprintf("%+d", m.Disp)
		}
	}
	if m.Base != nil {
		a += fmt.Sprintf("(%s)", m.Base.Asm())
	}
	if m.Index != nil && m.Scale != 0 {
		a += fmt.Sprintf("(%s*%d)", m.Index.Asm(), m.Scale)
	}
	return a
}

type Imm uint64

func (i Imm) Asm() string {
	return fmt.Sprintf("$%#x", i)
}

// Rel is an offset relative to the instruction pointer.
type Rel int32

func (r Rel) Asm() string {
	return fmt.Sprintf(".%+d", r)
}

// LabelRef is a reference to a label.
type LabelRef string

func (l LabelRef) Asm() string {
	return string(l)
}

// Registers returns the list of all operands involved in the given operand.
func Registers(op Op) []reg.Register {
	switch op := op.(type) {
	case reg.Register:
		return []reg.Register{op}
	case Mem:
		var r []reg.Register
		if op.Base != nil {
			r = append(r, op.Base)
		}
		if op.Index != nil {
			r = append(r, op.Index)
		}
		return r
	case Imm, Rel, LabelRef:
		return nil
	}
	panic("unknown operand type")
}

// ApplyAllocation returns an operand with allocated registers replaced. Registers missing from the allocation are left alone.
func ApplyAllocation(op Op, a reg.Allocation) Op {
	switch op := op.(type) {
	case reg.Register:
		return a.LookupDefault(op)
	case Mem:
		op.Base = a.LookupDefault(op.Base)
		op.Index = a.LookupDefault(op.Index)
		return op
	}
	return op
}
