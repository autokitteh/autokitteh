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
			"%s the %s",
			names[pick(len(names))],
			adjectives[pick(len(adjectives))],
		)
	}
}

var Generate = NewGenerator(nil)
