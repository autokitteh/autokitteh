// @generated by protoc-gen-es v1.5.1 with parameter "target=ts"
// @generated from file autokitteh/triggers/v1/trigger.proto (package autokitteh.triggers.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { CodeLocation } from "../../program/v1/program_pb.js";
import { Value } from "../../values/v1/values_pb.js";

/**
 * @generated from message autokitteh.triggers.v1.Trigger
 */
export class Trigger extends Message<Trigger> {
  /**
   * @generated from field: string trigger_id = 1;
   */
  triggerId = "";

  /**
   * @generated from field: string name = 2;
   */
  name = "";

  /**
   * @generated from field: string connection_id = 3;
   */
  connectionId = "";

  /**
   * if empty, applies to all envs.
   *
   * @generated from field: string env_id = 4;
   */
  envId = "";

  /**
   * @generated from field: string event_type = 5;
   */
  eventType = "";

  /**
   * @generated from field: autokitteh.program.v1.CodeLocation code_location = 6;
   */
  codeLocation?: CodeLocation;

  /**
   * @generated from field: string filter = 7;
   */
  filter = "";

  /**
   * @generated from field: map<string, autokitteh.values.v1.Value> data = 8;
   */
  data: { [key: string]: Value } = {};

  constructor(data?: PartialMessage<Trigger>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.triggers.v1.Trigger";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "trigger_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "connection_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "env_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "event_type", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 6, name: "code_location", kind: "message", T: CodeLocation },
    { no: 7, name: "filter", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 8, name: "data", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "message", T: Value} },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Trigger {
    return new Trigger().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Trigger {
    return new Trigger().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Trigger {
    return new Trigger().fromJsonString(jsonString, options);
  }

  static equals(a: Trigger | PlainMessage<Trigger> | undefined, b: Trigger | PlainMessage<Trigger> | undefined): boolean {
    return proto3.util.equals(Trigger, a, b);
  }
}

