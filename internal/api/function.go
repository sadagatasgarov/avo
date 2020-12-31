package api

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/mmcloughlin/avo/internal/inst"
)

// Function represents a function that constructs some collection of
// instruction forms.
type Function struct {
	Instruction inst.Instruction
	Suffixes    []string
	inst.Forms
}

// Name returns the function name.
func (f *Function) Name() string {
	return f.opcodesuffix("_")
}

// Opcode returns the full Go opcode of the instruction built by this function. Includes any suffixes.
func (f *Function) Opcode() string {
	return f.opcodesuffix(".")
}

func (f *Function) opcodesuffix(sep string) string {
	n := f.Instruction.Opcode
	for _, suffix := range f.Suffixes {
		n += sep
		n += suffix
	}
	return n
}

// Doc returns the function document comment as a list of lines.
func (f *Function) Doc() []string {
	lines := []string{
		fmt.Sprintf("%s: %s.", f.Name(), f.Instruction.Summary),
		"",
		"Forms:",
		"",
	}

	// Write a table of instruction forms.
	buf := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)
	for _, form := range f.Forms {
		row := f.Opcode() + "\t" + strings.Join(form.Signature(), "\t") + "\n"
		fmt.Fprint(w, row)
	}
	w.Flush()

	tbl := strings.TrimSpace(buf.String())
	for _, line := range strings.Split(tbl, "\n") {
		lines = append(lines, "\t"+line)
	}

	return lines
}

// Signature of the function. Derived from the instruction forms generated by this function.
func (f *Function) Signature() Signature {
	// Handle the case of forms with multiple arities.
	switch {
	case f.IsVariadic():
		return variadic{name: "ops"}
	case f.IsNiladic():
		return niladic{}
	}

	// Generate nice-looking variable names.
	n := f.Arity()
	ops := make([]string, n)
	count := map[string]int{}
	for j := 0; j < n; j++ {
		// Collect unique lowercase bytes from first characters of operand types.
		s := map[byte]bool{}
		for _, form := range f.Forms {
			c := form.Operands[j].Type[0]
			if 'a' <= c && c <= 'z' {
				s[c] = true
			}
		}

		// Operand name is the sorted bytes.
		var b []byte
		for c := range s {
			b = append(b, c)
		}
		sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })
		name := string(b)

		// Append a counter if we've seen it already.
		m := count[name]
		count[name]++
		if m > 0 {
			name += strconv.Itoa(m)
		}
		ops[j] = name
	}

	return argslist(ops)
}

// InstructionFunctions builds the list of all functions for a given
// instruction.
func InstructionFunctions(i inst.Instruction) []*Function {
	// One function for each possible suffix combination.
	bysuffix := map[string]*Function{}
	for _, f := range i.Forms {
		for _, suffixes := range f.SupportedSuffixes() {
			k := strings.Join(suffixes, ".")
			if _, ok := bysuffix[k]; !ok {
				bysuffix[k] = &Function{
					Instruction: i,
					Suffixes:    suffixes,
				}
			}
			bysuffix[k].Forms = append(bysuffix[k].Forms, f)
		}
	}

	// Convert to a sorted slice.
	var fns []*Function
	for _, fn := range bysuffix {
		fns = append(fns, fn)
	}

	SortFunctions(fns)

	return fns
}

// InstructionsFunctions builds all functions for a list of instructions.
func InstructionsFunctions(is []inst.Instruction) []*Function {
	var all []*Function
	for _, i := range is {
		fns := InstructionFunctions(i)
		all = append(all, fns...)
	}

	SortFunctions(all)

	return all
}

// SortFunctions sorts a list of functions by name.
func SortFunctions(fns []*Function) {
	sort.Slice(fns, func(i, j int) bool {
		return fns[i].Name() < fns[j].Name()
	})
}
