package catnames

import (
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var caser = cases.Title(language.English)

func NewGenerator(pick func(int) int) func() string {
	if pick == nil {
		pick = rand.New(rand.NewSource(time.Now().UnixNano())).Intn
	}

	return func() string {
		i := pick(len(names))
		j := pick(len(adjectives))

		return caser.String(fmt.Sprintf("%s the %s", names[i], adjectives[j]))
	}
}

var Generate = NewGenerator(nil)
