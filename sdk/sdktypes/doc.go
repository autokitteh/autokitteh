// Exposes types that are used in the SDK.
// Those types are mostly an immutable boxes for proto defined types.
// The idea behind hiding behind a box is twofold:
// 1. Make the data immutable.
// 2. Make invalid data unrepresantble.
// There must be no way to instansiate an object or an id outside
// of the SDK, other than other via SDK methods, such as <Object>FromProto or New<Object>ID.
//
// Strict* versions of object functions do not allow for missing required fields.
// They do allow, though, for missing the entire object entirely (nil input).
//
// Strict* versions of ids and handles (non-objects) do not allow for missing values at all.
package sdktypes

// TODO: Strict by default, non-strict on request?
