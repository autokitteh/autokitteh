package sdktypes

type concreteValue interface {
	Object

	isConcreteValue()
}

func IsConcreteValue(x any) bool { _, ok := x.(concreteValue); return ok }
