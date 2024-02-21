package actions

type Action interface {
	GetKey() string

	Type() string

	isAction()
}
