package selectors

type constraint struct {
	function[bool]
}

func newConstraint(label string, f any) *constraint {
	function := newFunction[bool]([]string{label}, f)

	if function == nil {
		return nil
	}

	return &constraint{function: *function}
}

func (c *constraint) Check(label string, value any) bool {
	return c.Call(map[string]any{label: value})
}
