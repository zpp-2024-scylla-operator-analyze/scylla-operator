package selectors

type predicate struct {
	function[bool]
}

func newPredicate(label string, f any) *predicate {
	function := newFunction[bool]([]string{label}, f)

	if function == nil {
		return nil
	}

	return &predicate{function: *function}
}

func (p *predicate) Check(label string, value any) bool {
	return p.Call(map[string]any{label: value})
}
