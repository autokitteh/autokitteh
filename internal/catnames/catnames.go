package catnames

import (
	"fmt"
	"math/rand"
	"time"
)

func NewGenerator(pick func(int) int) func() string {
	if pick == nil {
		pick = rand.New(rand.NewSource(time.Now().UnixNano())).Intn
	}

	return func() string {
		return fmt.Sprintf(
			"%s %s",
			adjectives[pick(len(adjectives))],
			names[pick(len(names))],
		)
	}
}

var Generate = NewGenerator(nil)
