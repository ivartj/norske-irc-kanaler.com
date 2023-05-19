package util

func Map[T any, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i, t := range ts {
		us[i] = f(t)
	}
	return us
}

func GroupBy[K comparable, V any](vs []V, f func(V) K) map[K][]V {
	m := map[K][]V{}
	for _, v := range vs {
		k := f(v)
		kvs, ok := m[k]
		if !ok {
			kvs = []V{v}
		} else {
			kvs = append(kvs, v)
		}
		m[k] = kvs
	}
	return m
}

func Filter[T any](ts []T, f func(t T) bool) []T {
	ts2 := make([]T, 0, len(ts))
	for _, t := range ts {
		if f(t) {
			ts2 = append(ts2, t)
		}
	}
	return ts2
}
