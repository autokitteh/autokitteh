package apiproject

import (
	"github.com/dustinkirkland/golang-petname"
)

type ProjectID string

var EmptyProjectID ProjectID

func (id ProjectID) String() string { return string(id) }

func (id *ProjectID) MaybeString() string {
	if id == nil {
		return ""
	}

	return id.String()
}

func (id ProjectID) IsEmpty() bool { return id == "" }

func NewProjectID() ProjectID { return ProjectID(petname.Generate(4, "_")) }
