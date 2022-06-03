package bottom

func myMap[A, B any](f func(a A) B, xs []A) []B {
	var out []B
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}
