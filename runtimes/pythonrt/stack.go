package pythonrt

type Stack[T any] []T

func (s *Stack[T]) Push(val T) {
	*s = append(*s, val)
}

func (s *Stack[T]) Len() int {
	return len(*s)
}

func (s *Stack[T]) Pop() T {
	s.panicEmpty()

	v := (*s)[s.Len()-1]
	*s = (*s)[:s.Len()-1]

	return v
}

func (s *Stack[T]) Top() T {
	s.panicEmpty()

	return (*s)[s.Len()-1]
}

func (s *Stack[T]) panicEmpty() {
	if s.Len() == 0 {
		panic("empty stack")
	}
}
