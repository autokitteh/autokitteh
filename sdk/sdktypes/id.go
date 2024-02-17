package sdktypes

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

const idDelim = ":"

var validIDRe = regexp.MustCompile(`^\w+:[0-9a-f]+$`)

// IsID returns if s might be an ID, not a valid ID.
//
// In the CLI, UI, or even some routes we might want to get a ref by either a
// handle or an ID. For that, we need to understand what the user intended to
// put, and if it's an invalid ID - still treat it as an ID and say "this is an
// invalid ID".
func IsID(s string) bool { return strings.Contains(s, idDelim) }

func IsValidID(s string) bool { return validIDRe.MatchString(s) }

type ID interface {
	fmt.Stringer
	json.Marshaler
	json.Unmarshaler

	Kind() string
	Value() string

	isID()
}

// idTraits define the expected format for the ID.
type idTraits interface {
	Kind() string
	ValidateValue(string) error
}

type id[T idTraits] struct{ kind, value string }

func (id *id[T]) isID() {}

func (id *id[T]) String() string {
	if id == nil {
		return ""
	}

	return fmt.Sprintf("%s%s%s", id.kind, idDelim, id.value)
}

func (id *id[T]) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, id.String())), nil
}

func (id *id[T]) Kind() string {
	if id == nil {
		return ""
	}
	return id.kind
}

func (id *id[T]) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}

	id1, err := strictParseTypedID[T](text)
	if err != nil {
		return err
	}

	*id = *id1
	return nil
}

func (id *id[T]) Value() string {
	if id == nil {
		return ""
	}
	return id.value
}

func SplitRawID(raw string) (kind, data string, ok bool) {
	kind, data, ok = strings.Cut(raw, idDelim)
	return
}

// If raw == "", returns nil, nil.
func parseTypedID[T idTraits](raw string) (*id[T], error) {
	if raw == "" {
		return nil, nil
	}

	return strictParseTypedID[T](raw)
}

func strictParseTypedID[T idTraits](raw string) (*id[T], error) {
	kind, value, ok := SplitRawID(raw)
	if !ok {
		return nil, fmt.Errorf("%w: no delimiter", sdkerrors.ErrInvalidArgument)
	}

	var t T

	if kind != t.Kind() {
		return nil, fmt.Errorf("%q != expected %q: %w", kind, t.Kind(), sdkerrors.ErrInvalidArgument)
	}

	if err := t.ValidateValue(value); err != nil {
		err = errors.Join(sdkerrors.ErrInvalidArgument, err)
		return nil, fmt.Errorf("id value (%v): %w", value, err)
	}

	return &id[T]{
		kind:  t.Kind(),
		value: value,
	}, nil
}

func parseIDOrName[T idTraits](raw string) (h Name, id *id[T], err error) {
	if raw == "" {
		return nil, nil, fmt.Errorf("must not be empty: %w", sdkerrors.ErrInvalidArgument)
	}

	if _, _, ok := SplitRawID(raw); ok {
		id, err = parseTypedID[T](raw)
		return
	}

	h, err = ParseName(raw)
	return
}

func StrictParseAnyID(raw string) (ID, error) {
	kind, _, ok := SplitRawID(raw)
	if !ok {
		return nil, sdkerrors.ErrInvalidArgument
	}

	switch kind {
	case ProjectIDKind:
		return ParseProjectID(raw)
	case EnvIDKind:
		return ParseEnvID(raw)
	case DeploymentIDKind:
		return ParseDeploymentID(raw)
	case SessionIDKind:
		return ParseSessionID(raw)
	case BuildIDKind:
		return ParseBuildID(raw)
	case ConnectionIDKind:
		return ParseConnectionID(raw)
	case EventIDKind:
		return ParseEventID(raw)
	case MappingIDKind:
		return ParseMappingID(raw)
	case IntegrationIDKind:
		return ParseIntegrationID(raw)

	default:
		return nil, fmt.Errorf("unrecognized kind %q: %w", kind, sdkerrors.ErrInvalidArgument)
	}
}

func ParseAnyID(raw string) (ID, error) {
	if raw == "" {
		return nil, nil
	}

	return StrictParseAnyID(raw)
}
