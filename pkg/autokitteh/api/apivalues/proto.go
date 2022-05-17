package apivalues

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbvalues "github.com/autokitteh/autokitteh/api/gen/stubs/go/values"
)

var ErrUnknownType = errors.New("(proto) unknown type")

func ValuesListToProto(vs []*Value) []*pbvalues.Value {
	pbs := make([]*pbvalues.Value, len(vs))
	for i, v := range vs {
		pbs[i] = v.PB()
	}
	return pbs
}

func MustValuesListFromProto(pbs []*pbvalues.Value) []*Value {
	vs, err := ValuesListFromProto(pbs)
	if err != nil {
		panic(err)
	}
	return vs
}

func ValuesListFromProto(pbs []*pbvalues.Value) ([]*Value, error) {
	vs := make([]*Value, len(pbs))
	for i, pb := range pbs {
		var err error
		if vs[i], err = ValueFromProto(pb); err != nil {
			return nil, fmt.Errorf("index %d: %w", i, err)
		}
	}
	return vs, nil
}

func StringValueMapToProto(m map[string]*Value) map[string]*pbvalues.Value {
	r := make(map[string]*pbvalues.Value, len(m))
	for k, v := range m {
		r[k] = v.PB()
	}
	return r
}

func MustStringValueMapFromProto(m map[string]*pbvalues.Value) map[string]*Value {
	vs, err := StringValueMapFromProto(m)
	if err != nil {
		panic(err)
	}
	return vs
}

func StringValueMapFromProto(m map[string]*pbvalues.Value) (map[string]*Value, error) {
	r := make(map[string]*Value, len(m))
	for k, v := range m {
		var err error
		if r[k], err = ValueFromProto(v); err != nil {
			return nil, fmt.Errorf("key %s: %w", k, err)
		}
	}
	return r, nil
}

// return value is not validated.
func toProto(v value) *pbvalues.Value {
	var pb pbvalues.Value

	switch vv := v.(type) {
	case NoneValue:
		pb.Type = &pbvalues.Value_None{None: &pbvalues.None{}}
	case StringValue:
		pb.Type = &pbvalues.Value_String_{String_: &pbvalues.String{V: vv.String()}}
	case SymbolValue:
		pb.Type = &pbvalues.Value_Symbol{Symbol: &pbvalues.Symbol{Name: vv.String()}}
	case IntegerValue:
		pb.Type = &pbvalues.Value_Integer{Integer: &pbvalues.Integer{V: (int64)(vv)}}
	case BooleanValue:
		pb.Type = &pbvalues.Value_Boolean{Boolean: &pbvalues.Boolean{V: (bool)(vv)}}
	case FloatValue:
		pb.Type = &pbvalues.Value_Float{Float: &pbvalues.Float{V: (float32)(vv)}}
	case BytesValue:
		pb.Type = &pbvalues.Value_Bytes{Bytes: &pbvalues.Bytes{V: ([]byte)(vv)}}
	case TimeValue:
		pb.Type = &pbvalues.Value_Time{Time: &pbvalues.Time{T: timestamppb.New(time.Time(vv))}}
	case DurationValue:
		pb.Type = &pbvalues.Value_Duration{Duration: &pbvalues.Duration{D: durationpb.New(time.Duration(vv))}}
	case CallValue:
		pb.Type = &pbvalues.Value_Call{Call: &pbvalues.Call{Id: vv.ID, Name: vv.Name, Flags: vv.Flags, Issuer: vv.Issuer}}
	case FunctionValue:
		var pbsig *pbvalues.FunctionSignature

		if sig := vv.Signature; sig != nil {
			pbsig = &pbvalues.FunctionSignature{
				Name:        sig.Name,
				Doc:         sig.Doc,
				NArgs:       sig.NumArgs,
				NKwonlyargs: sig.NumKWOnlyArgs,
				ArgsNames:   sig.ArgsNames,
				HasKwargs:   sig.HasKWArgs,
				HasVarargs:  sig.HasVarargs,
			}
		}

		pb.Type = &pbvalues.Value_Function{
			Function: &pbvalues.Function{
				Lang:   vv.Lang,
				FuncId: vv.FuncID,
				Scope:  vv.Scope,
				Sig:    pbsig,
			},
		}
	case ListValue:
		pb.Type = &pbvalues.Value_List{List: listValueToProto(vv)}
	case SetValue:
		pb.Type = &pbvalues.Value_Set{Set: setValueToProto(vv)}
	case DictValue:
		pb.Type = &pbvalues.Value_Dict{Dict: dictValueToProto(vv)}
	case StructValue:
		pb.Type = &pbvalues.Value_Struct{Struct: &pbvalues.Struct{Ctor: vv.Ctor.PB(), Fields: StringValueMapToProto(vv.Fields)}}
	case ModuleValue:
		pb.Type = &pbvalues.Value_Module{Module: &pbvalues.Module{Name: vv.Name, Members: StringValueMapToProto(vv.Members)}}
	default:
		panic(fmt.Errorf("(to) %w: %v", ErrUnknownType, reflect.TypeOf(v)))
	}

	return &pb
}

func dictValueToProto(d DictValue) *pbvalues.Dict {
	vs := &pbvalues.Dict{Items: make([]*pbvalues.DictItem, len(d))}
	for i, v := range d {
		vs.Items[i] = &pbvalues.DictItem{K: v.K.PB(), V: v.V.PB()}
	}
	return vs
}

func setValueToProto(s SetValue) *pbvalues.Set {
	vs := &pbvalues.Set{Vs: make([]*pbvalues.Value, len(s))}
	for i, v := range s {
		vs.Vs[i] = v.PB()
	}
	return vs
}

func listValueToProto(l ListValue) *pbvalues.List {
	vl := &pbvalues.List{Vs: make([]*pbvalues.Value, len(l))}
	for i, v := range l {
		vl.Vs[i] = v.PB()
	}
	return vl
}

func fromProto(pb interface{}) (value, error) {
	switch v := pb.(type) {
	case *pbvalues.Value_None:
		return NoneValue{}, nil
	case *pbvalues.Value_String_:
		return StringValue(v.String_.V), nil
	case *pbvalues.Value_Symbol:
		return SymbolValue(v.Symbol.Name), nil
	case *pbvalues.Value_Integer:
		return IntegerValue(v.Integer.V), nil
	case *pbvalues.Value_Boolean:
		return BooleanValue(v.Boolean.V), nil
	case *pbvalues.Value_Float:
		return FloatValue(v.Float.V), nil
	case *pbvalues.Value_List:
		return listValueFromProto(v.List.Vs)
	case *pbvalues.Value_Dict:
		return dictValueFromProto(v.Dict.Items)
	case *pbvalues.Value_Bytes:
		return BytesValue(v.Bytes.V), nil
	case *pbvalues.Value_Set:
		return setValueFromProto(v.Set.Vs)
	case *pbvalues.Value_Call:
		return CallValue{ID: v.Call.Id, Name: v.Call.Name, Flags: v.Call.Flags, Issuer: v.Call.Issuer}, nil
	case *pbvalues.Value_Time:
		return TimeValue(v.Time.T.AsTime()), nil
	case *pbvalues.Value_Duration:
		return DurationValue(v.Duration.D.AsDuration()), nil
	case *pbvalues.Value_Function:
		var sig *FunctionSignature

		if pbsig := v.Function.Sig; pbsig != nil {
			sig = &FunctionSignature{
				Name:          pbsig.Name,
				Doc:           pbsig.Doc,
				NumArgs:       pbsig.NArgs,
				NumKWOnlyArgs: pbsig.NKwonlyargs,
				ArgsNames:     pbsig.ArgsNames,
				HasKWArgs:     pbsig.HasKwargs,
				HasVarargs:    pbsig.HasVarargs,
			}
		}

		return FunctionValue{
			Lang:      v.Function.Lang,
			FuncID:    v.Function.FuncId,
			Scope:     v.Function.Scope,
			Signature: sig,
		}, nil
	case *pbvalues.Value_Struct:
		ctor, err := ValueFromProto(v.Struct.Ctor)
		if err != nil {
			return nil, fmt.Errorf("ctor: %w", err)
		}

		fs, err := StringValueMapFromProto(v.Struct.Fields)
		if err != nil {
			return nil, fmt.Errorf("fields: %w", err)
		}

		return StructValue{Ctor: ctor, Fields: fs}, nil
	case *pbvalues.Value_Module:
		ms, err := StringValueMapFromProto(v.Module.Members)
		if err != nil {
			return nil, fmt.Errorf("members: %w", err)
		}

		return ModuleValue{Name: v.Module.Name, Members: ms}, nil
	default:
		return nil, fmt.Errorf("(from) %w: %v", ErrUnknownType, reflect.TypeOf(v))
	}
}

func setValueFromProto(vs []*pbvalues.Value) (s SetValue, err error) {
	s = SetValue(make([]*Value, len(vs)))

	for i, v := range vs {
		s[i], err = ValueFromProto(v)
		if err != nil {
			return nil, fmt.Errorf("item %d: %w", i, err)
		}
	}

	return
}

func listValueFromProto(vs []*pbvalues.Value) (l ListValue, err error) {
	l = ListValue(make([]*Value, len(vs)))

	for i, v := range vs {
		l[i], err = ValueFromProto(v)
		if err != nil {
			return nil, fmt.Errorf("item %d: %w", i, err)
		}
	}

	return
}

func dictValueFromProto(vs []*pbvalues.DictItem) (DictValue, error) {
	d := DictValue(make([]*DictItem, len(vs)))

	for i, v := range vs {
		k, err := ValueFromProto(v.K)
		if err != nil {
			return nil, fmt.Errorf("item %d key: %w", i, err)
		}

		v, err := ValueFromProto(v.V)
		if err != nil {
			return nil, fmt.Errorf("item %d value: %w", i, err)
		}

		d[i] = &DictItem{K: k, V: v}
	}

	return d, nil
}
