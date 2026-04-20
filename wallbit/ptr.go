package wallbit

// Ptr returns a pointer to its argument. It exists because request bodies
// for the Wallbit API use pointer fields to distinguish "unset" from
// "explicit zero value".
// Without a helper, callers have
// to introduce a throwaway local for every optional field:
//
//	name := "Jeremy"
//	req := UpdateProfile{Name: &name}
//
// which scales poorly when a request has several optional fields. With
// the generic helper the same code collapses to:
//
//	req := UpdateProfile{Name: wallbit.Ptr("Jeremy"), Age: wallbit.Ptr(28)}
func Ptr[T any](v T) *T {
	return &v
}
