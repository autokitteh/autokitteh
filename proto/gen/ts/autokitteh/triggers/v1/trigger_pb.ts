// @generated by protoc-gen-es v1.5.1 with parameter "target=ts"
// @generated from file autokitteh/triggers/v1/trigger.proto (package autokitteh.triggers.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { CodeLocation } from "../../program/v1/program_pb.js";

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
   * @generated from field: autokitteh.triggers.v1.Trigger.SourceType source_type = 3;
   */
  sourceType = Trigger_SourceType.UNSPECIFIED;

  /**
   * @generated from field: string project_id = 4;
   */
  projectId = "";

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
   * if source_type == CONNECTION.
   *
   * @generated from field: string connection_id = 50;
   */
  connectionId = "";

  /**
   * if source_type == SCHEDULE.
   *
   * @generated from field: string schedule = 51;
   */
  schedule = "";

  /**
   * read only.
   *
   * if source_type == WEBHOOK, after creation.
   *
   * @generated from field: string webhook_slug = 100;
   */
  webhookSlug = "";

  constructor(data?: PartialMessage<Trigger>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.triggers.v1.Trigger";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "trigger_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "source_type", kind: "enum", T: proto3.getEnumType(Trigger_SourceType) },
    { no: 4, name: "project_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "event_type", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 6, name: "code_location", kind: "message", T: CodeLocation },
    { no: 7, name: "filter", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 50, name: "connection_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 51, name: "schedule", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 100, name: "webhook_slug", kind: "scalar", T: 9 /* ScalarType.STRING */ },
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

/**
 * @generated from enum autokitteh.triggers.v1.Trigger.SourceType
 */
export enum Trigger_SourceType {
  /**
   * @generated from enum value: SOURCE_TYPE_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * @generated from enum value: SOURCE_TYPE_CONNECTION = 1;
   */
  CONNECTION = 1,

  /**
   * @generated from enum value: SOURCE_TYPE_WEBHOOK = 2;
   */
  WEBHOOK = 2,

  /**
   * @generated from enum value: SOURCE_TYPE_SCHEDULE = 3;
   */
  SCHEDULE = 3,
}
// Retrieve enum metadata with: proto3.getEnumType(Trigger_SourceType)
proto3.util.setEnumType(Trigger_SourceType, "autokitteh.triggers.v1.Trigger.SourceType", [
  { no: 0, name: "SOURCE_TYPE_UNSPECIFIED" },
  { no: 1, name: "SOURCE_TYPE_CONNECTION" },
  { no: 2, name: "SOURCE_TYPE_WEBHOOK" },
  { no: 3, name: "SOURCE_TYPE_SCHEDULE" },
]);

