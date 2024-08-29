// @generated by protoc-gen-es v1.5.1 with parameter "target=ts"
// @generated from file autokitteh/vars/v1/var.proto (package autokitteh.vars.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * @generated from message autokitteh.vars.v1.Var
 */
export class Var extends Message<Var> {
  /**
   * @generated from field: string scope_id = 1;
   */
  scopeId = "";

  /**
   * @generated from field: string name = 2;
   */
  name = "";

  /**
   * @generated from field: string value = 3;
   */
  value = "";

  /**
   * @generated from field: bool is_secret = 4;
   */
  isSecret = false;

  constructor(data?: PartialMessage<Var>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.vars.v1.Var";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "scope_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "value", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "is_secret", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Var {
    return new Var().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Var {
    return new Var().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Var {
    return new Var().fromJsonString(jsonString, options);
  }

  static equals(a: Var | PlainMessage<Var> | undefined, b: Var | PlainMessage<Var> | undefined): boolean {
    return proto3.util.equals(Var, a, b);
  }
}

