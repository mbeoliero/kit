package repox

type IList[T any] interface {
	List() []func(*T)
}

func NewOptions[T any](opts ...IList[T]) *T {
	o := new(T)
	for _, opt := range opts {
		for _, setArgs := range opt.List() {
			if setArgs == nil {
				continue
			}

			setArgs(o)
		}
	}
	return o
}
