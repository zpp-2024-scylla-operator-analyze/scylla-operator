package analyze

import (
	"k8s.io/apimachinery/pkg/fields"
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

func (m *ExpMatcher) requirementMatches(obj interface{}, r fields.Requirement) bool {
	it := GetFieldValueIterator(r.Field, obj)
	val := it()
	for val != nil && val.String() != r.Value {
		val = it()
	}
	return val != nil
}

func (m *ExpMatcher) relationsMatch(target interface{}, r *Rule, idx int, chosen *[]interface{}) bool {
	for _, cond := range r.Relations {
		var (
			lhs interface{}
			rhs interface{}
		)
		// "Prefix" match
		prefixOk := false
		if idx == cond.Rhs && idx >= cond.Lhs {
			lhs = (*chosen)[cond.Lhs]
			rhs = target
			prefixOk = true
		} else if idx >= cond.Rhs && idx == cond.Lhs {
			lhs = target
			rhs = (*chosen)[cond.Lhs]
			prefixOk = true
		}

		if prefixOk {
			match, err := cond.Rel.EvaluateOn(lhs, rhs)
			if err != nil {
				panic(err)
			}
			if !match {
				return false
			}
		}
	}
	return true
}

func (m *ExpMatcher) tryMatch(r *Rule, idx int, chosen *[]interface{}) bool {
	if idx >= len(*chosen) {
		// match found
		return true
	}

	resources := m.ds.resourcesOfKind(r.Resources[idx].Kind)
	success := false
	if len(resources) > 0 {
		// Go through possible resources
		for _, res := range resources {
			reqs := r.Resources[idx].Condition.Requirements()
			requirementsOk := true
			for _, req := range reqs {
				if !m.requirementMatches(res, req) {
					requirementsOk = false
					break
				}
			}
			if !requirementsOk {
				continue
			}
			if m.relationsMatch(res, r, idx, chosen) {
				(*chosen)[idx] = res
				found := m.tryMatch(r, idx+1, chosen)
				if found {
					success = true
				}
			}
		}
	} else {
		if m.relationsMatch(nil, r, idx, chosen) {
			(*chosen)[idx] = nil
			success = m.tryMatch(r, idx+1, chosen)
		}
	}

	return success
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
