package selectors

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
)

func ExampleConstraint() {
	c := newConstraint("pod", func(_ v1.Pod) bool {
		return true
	})

	fmt.Println(c)

	c.Call(map[string]any{"pod": v1.Pod{}})
}
