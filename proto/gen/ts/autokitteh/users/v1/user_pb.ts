// @generated by protoc-gen-es v1.5.1 with parameter "target=ts"
// @generated from file autokitteh/users/v1/user.proto (package autokitteh.users.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * `display_name` is a human-readable name for the user.
 * `name` is a globally unique symbol for the user.
 *
 * @generated from message autokitteh.users.v1.User
 */
export class User extends Message<User> {
  /**
   * @generated from field: string user_id = 1;
   */
  userId = "";

  /**
   * if email is empty, user cannot login.
   *
   * @generated from field: string email = 2;
   */
  email = "";

  /**
   * @generated from field: string display_name = 3;
   */
  displayName = "";

  /**
   * @generated from field: bool disabled = 4;
   */
  disabled = false;

  /**
   * org to use for projects, if not otherwise specified.
   *
   * @generated from field: string default_org_id = 5;
   */
  defaultOrgId = "";

  constructor(data?: PartialMessage<User>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.users.v1.User";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "user_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "email", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "display_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "disabled", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 5, name: "default_org_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): User {
    return new User().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): User {
    return new User().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): User {
    return new User().fromJsonString(jsonString, options);
  }

  static equals(a: User | PlainMessage<User> | undefined, b: User | PlainMessage<User> | undefined): boolean {
    return proto3.util.equals(User, a, b);
  }
}

