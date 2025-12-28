package builder

func ToAnySlice[T any](data []T) []any {
	v := make([]any, len(data))
	for i := range data {
		v[i] = data[i]
	}
	return v
}
