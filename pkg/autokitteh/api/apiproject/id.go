package apiproject

import (
	"fmt"
	"strings"

	"github.com/dustinkirkland/golang-petname"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiaccount"
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

func (id ProjectID) AccountName() apiaccount.AccountName {
	a, _, _ := strings.Cut(id.String(), ".")
	return apiaccount.AccountName(a)
}

func NewProjectID(aname apiaccount.AccountName) ProjectID {
	return ProjectID(fmt.Sprintf("%v.%s", aname, petname.Generate(4, "_")))
}
