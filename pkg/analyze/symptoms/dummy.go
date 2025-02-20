package symptoms

var DummySymptoms = NewSymptomSet("dummy", []*SymptomSet{
	buildDummySymptoms(),
})

func buildDummySymptoms() *SymptomSet {
	//query := selectors.
	//	Select(selectors.Resource[scyllav1.ScyllaCluster]("cluster")).
	//	Select(selectors.Resource[v1.Pod]("pod")).
	//	Filter("pod", func(_ *v1.Pod) (bool, error) {
	//		return false, nil
	//	}).
	//	Relate("cluster", "pod", func(_ *scyllav1.ScyllaCluster, _ *v1.Pod) (bool, error) {
	//		return false, nil
	//	}).
	//	Collect([]string{"a"}, func(_ string, _ *v1.Pod) {})
	//
	//query := selectors.Builder().
	//	New("cluster", selectors.Type[scyllav1.ScyllaCluster]()).
	//	New("pod", selectors.Type[v1.Pod]()).
	//	Join(&selectors.FuncRelation[*scyllav1.ScyllaCluster, *v1.Pod]{
	//		Lhs: "cluster",
	//		Rhs: "pod",
	//		Lambda: func(_ *scyllav1.ScyllaCluster, _ *v1.Pod) (bool, error) {
	//			return false, nil
	//		},
	//	}).
	//	Where(&selectors.FuncConstraint[*v1.Pod]{
	//		Resource: "pod",
	//		Lambda: func(_ *v1.Pod) (bool, error) {
	//			return false, nil
	//		},
	//	}).
	//	Any()
	return nil
}
