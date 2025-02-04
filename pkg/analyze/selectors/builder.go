package selectors

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
)

type builder struct {
	// TODO
}

func New() *builder {
	return &builder{}
}

func (q *builder) Select(name string, resource string) *builder {
	// TODO: Define field of name `name` of resource type `resource`

	return q
}

func (q *builder) Where(name string, lambda /* func (resource) bool */ string) *builder {
	// TODO: Assert that for all results `lambda` given field `name` returns true

	return q
}

func (q *builder) Join(lhs string, rhs string, lambda /* func (lhs, rhs) bool */ string) *builder {
	// TODO: Check if `labda` for fields named `lhs` and `rhs` respectively returns true

	return q
}

func (q *builder) Any() func(*sources.DataSource) (bool, error) {
	// TODO: Build a lambda that will take k8s state and evaluate query
	// TODO: Maybe return an instance of an interface that is evaluatable

	return func(ds *sources.DataSource) (bool, error) {
		return true, nil
	}
}
