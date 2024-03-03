// This package contains the types used by the SDK.
//
// All the types represent data structures that are defined in the proto package.
// The underlying data is either a protobuf message or a primitive that can be
// set as a field in other messages. The main idea is that it should be
// impossible to have a object that contains a malform/unvalidated values.
// An object is either valid with valid information, or invalid with a nil
// message.
//
// There are three main types of underlying data:
//  1. Messages, which are also called "objects" here, which are structs that
//     contain other fields. These are implemented using the `object` struct.
//     Examples: Project, Session, etc.
//  2. Validated strings, which are Names and Symbols, implemented using the
//     `validatedString` struct.
//     Examples: Name and Symbol.
//  3. Enumerations, which are implemented using the `enum` struct.
//     Examples: DeployentState, SessionStateType, etc.
//
// Specific objects are defined like so:
//
//	  type Project struct { object[pb.Project, projectTraits] }
//	  type projectTraits struct {}
//	  func (projectTraits) Validate(m *pb.Project) error {
//	    // Validates all fields are correct. Each field is permitted to be empty.
//		 }
//	  func (projectTraits) StrictValidate(m *pb.Project) error {
//	    // Validates all mandatory fields are specified.
//	  }
//
// IDs and Names are kind of the same thing, but with different types. See examples
// mentioned above for more information.
package sdktypes
