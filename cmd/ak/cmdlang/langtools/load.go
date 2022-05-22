package langtools

import (
	"io/ioutil"

	"google.golang.org/protobuf/proto"

	pbprogram "github.com/autokitteh/autokitteh/api/gen/stubs/go/program"

	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
)

func Load(path string) (*apiprogram.Module, error) {
	compiled, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pbmod pbprogram.Module

	if err := proto.Unmarshal(compiled, &pbmod); err != nil {
		return nil, err
	}

	return apiprogram.ModuleFromProto(&pbmod)
}
