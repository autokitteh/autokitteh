// @generated by protoc-gen-es v1.5.1 with parameter "target=ts"
// @generated from file autokitteh/runtimes/v1/build.proto (package autokitteh.runtimes.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { CodeLocation } from "../../program/v1/program_pb.js";

/**
 * @generated from message autokitteh.runtimes.v1.Artifact
 */
export class Artifact extends Message<Artifact> {
  /**
   * @generated from field: repeated autokitteh.runtimes.v1.Requirement requirements = 1;
   */
  requirements: Requirement[] = [];

  /**
   * @generated from field: repeated autokitteh.runtimes.v1.Export exports = 2;
   */
  exports: Export[] = [];

  /**
   * Runtime specific build output. This essentially the "executable".
   * Map structure for convenience. Intended to use as a filesystem -
   * each entry will be stored as a different file in a persistent store.
   * This means that each key must be a relative path, no '..' or '.' allowed.
   *
   * @generated from field: map<string, bytes> compiled_data = 3;
   */
  compiledData: { [key: string]: Uint8Array } = {};

  constructor(data?: PartialMessage<Artifact>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.runtimes.v1.Artifact";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "requirements", kind: "message", T: Requirement, repeated: true },
    { no: 2, name: "exports", kind: "message", T: Export, repeated: true },
    { no: 3, name: "compiled_data", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 12 /* ScalarType.BYTES */} },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Artifact {
    return new Artifact().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Artifact {
    return new Artifact().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Artifact {
    return new Artifact().fromJsonString(jsonString, options);
  }

  static equals(a: Artifact | PlainMessage<Artifact> | undefined, b: Artifact | PlainMessage<Artifact> | undefined): boolean {
    return proto3.util.equals(Artifact, a, b);
  }
}

/**
 * @generated from message autokitteh.runtimes.v1.Requirement
 */
export class Requirement extends Message<Requirement> {
  /**
   * where the requirement is coming from.
   *
   * @generated from field: autokitteh.program.v1.CodeLocation location = 1;
   */
  location?: CodeLocation;

  /**
   * @generated from field: string url = 2;
   */
  url = "";

  /**
   * @generated from field: string symbol = 3;
   */
  symbol = "";

  constructor(data?: PartialMessage<Requirement>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.runtimes.v1.Requirement";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "location", kind: "message", T: CodeLocation },
    { no: 2, name: "url", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "symbol", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Requirement {
    return new Requirement().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Requirement {
    return new Requirement().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Requirement {
    return new Requirement().fromJsonString(jsonString, options);
  }

  static equals(a: Requirement | PlainMessage<Requirement> | undefined, b: Requirement | PlainMessage<Requirement> | undefined): boolean {
    return proto3.util.equals(Requirement, a, b);
  }
}

/**
 * @generated from message autokitteh.runtimes.v1.Export
 */
export class Export extends Message<Export> {
  /**
   * where the export is coming from.
   *
   * @generated from field: autokitteh.program.v1.CodeLocation location = 1;
   */
  location?: CodeLocation;

  /**
   * @generated from field: string symbol = 2;
   */
  symbol = "";

  constructor(data?: PartialMessage<Export>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.runtimes.v1.Export";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "location", kind: "message", T: CodeLocation },
    { no: 2, name: "symbol", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Export {
    return new Export().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Export {
    return new Export().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Export {
    return new Export().fromJsonString(jsonString, options);
  }

  static equals(a: Export | PlainMessage<Export> | undefined, b: Export | PlainMessage<Export> | undefined): boolean {
    return proto3.util.equals(Export, a, b);
  }
}

