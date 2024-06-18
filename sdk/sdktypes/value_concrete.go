package sdktypes

type concreteValue interface {
	Object

	IsTrue() bool

	isConcreteValue()
}

func IsConcreteValue(x any) bool { _, ok := x.(concreteValue); return ok }
