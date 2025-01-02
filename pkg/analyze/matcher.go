package analyze

import (
	"k8s.io/apimachinery/pkg/labels"
)

type Match struct {
	Rule      *Rule
	Resources []interface{}
}

type Matcher interface {
	MatchRule(r *Rule) (*Match, error)
}

type ExpMatcher struct {
	ds *DataSource
}

func NewMatcher(ds *DataSource) Matcher {
	return &ExpMatcher{
		ds: ds,
	}
}

func (d *DataSource) resourcesOfKind(kind string) []interface{} {
	var (
		r []interface{}
	)
	if kind == "Pod" {
		a, err := d.PodLister.List(labels.Everything())
		for _, res := range a {
			r = append(r, res)
		}
		if err != nil {
			panic(err)
		}
	} else if kind == "ScyllaCluster" {
		a, err := d.ScyllaClusterLister.List(labels.Everything())
		for _, res := range a {
			r = append(r, res)
		}
		if err != nil {
			panic(err)
		}
	}
	return r
}

func (m *ExpMatcher) tryMatch(r *Rule, idx int, chosen *[]interface{}) bool {
	if idx >= len(*chosen) {
		// match found
		return true
	}

	for _, res := range m.ds.resourcesOfKind(r.Resources[idx].Kind) {
		if !r.Resources[idx].Condition.Matches(&ObjectFields{obj: res}) {
			continue
		}

		good := true
		for _, cond := range r.Relations {
			var (
				lhs interface{}
				rhs interface{}
			)
			// "Prefix" match
			if idx == cond.Rhs && idx > cond.Lhs {
				lhs = (*chosen)[cond.Lhs]
				rhs = res
			} else if idx > cond.Rhs && idx == cond.Lhs {
				lhs = res
				rhs = (*chosen)[cond.Lhs]
			}

			if lhs != nil && rhs != nil {
				match, err := cond.Rel.EvaluateOn(lhs, rhs)
				if err != nil {
					panic(err)
				}
				good = good && match
			}
		}

		if good {
			(*chosen)[idx] = res
			found := m.tryMatch(r, idx+1, chosen)
			if found {
				return true
			}
		}
	}
	return false
}

func (m *ExpMatcher) MatchRule(r *Rule) (*Match, error) {
	chosen := make([]interface{}, len(r.Resources))
	if m.tryMatch(r, 0, &chosen) {
		return &Match{
			Rule:      r,
			Resources: chosen,
		}, nil
	}
	return nil, nil
}
