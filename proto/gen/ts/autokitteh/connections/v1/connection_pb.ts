// @generated by protoc-gen-es v1.5.1 with parameter "target=ts"
// @generated from file autokitteh/connections/v1/connection.proto (package autokitteh.connections.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { Status } from "../../common/v1/status_pb.js";

/**
 * TODO(ENG-1026):
 * - The first integration ID and project ID validation
 *   checks are incorrect for scheduler (cron) connections
 * - The name validation check breaks sdktypes.NewConnection(id)
 *
 * @generated from message autokitteh.connections.v1.Connection
 */
export class Connection extends Message<Connection> {
  /**
   * @generated from field: string connection_id = 1;
   */
  connectionId = "";

  /**
   * [(buf.validate.field).string.min_len = 1];
   *
   * @generated from field: string integration_id = 2;
   */
  integrationId = "";

  /**
   * [(buf.validate.field).string.min_len = 1];
   *
   * @generated from field: string project_id = 3;
   */
  projectId = "";

  /**
   * [(buf.validate.field).string.min_len = 1];
   *
   * @generated from field: string name = 4;
   */
  name = "";

  /**
   * Read only fields that are filled by the server.
   *
   * @generated from field: autokitteh.common.v1.Status status = 5;
   */
  status?: Status;

  /**
   * @generated from field: autokitteh.connections.v1.Capabilities capabilities = 6;
   */
  capabilities?: Capabilities;

  /**
   * @generated from field: map<string, string> links = 7;
   */
  links: { [key: string]: string } = {};

  constructor(data?: PartialMessage<Connection>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.connections.v1.Connection";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "connection_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "integration_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "project_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "status", kind: "message", T: Status },
    { no: 6, name: "capabilities", kind: "message", T: Capabilities },
    { no: 7, name: "links", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 9 /* ScalarType.STRING */} },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Connection {
    return new Connection().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Connection {
    return new Connection().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Connection {
    return new Connection().fromJsonString(jsonString, options);
  }

  static equals(a: Connection | PlainMessage<Connection> | undefined, b: Connection | PlainMessage<Connection> | undefined): boolean {
    return proto3.util.equals(Connection, a, b);
  }
}

/**
 * @generated from message autokitteh.connections.v1.Capabilities
 */
export class Capabilities extends Message<Capabilities> {
  /**
   * @generated from field: bool supports_connection_test = 1;
   */
  supportsConnectionTest = false;

  /**
   * @generated from field: bool supports_connection_init = 2;
   */
  supportsConnectionInit = false;

  /**
   * @generated from field: bool requires_connection_init = 3;
   */
  requiresConnectionInit = false;

  constructor(data?: PartialMessage<Capabilities>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.connections.v1.Capabilities";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "supports_connection_test", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 2, name: "supports_connection_init", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 3, name: "requires_connection_init", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Capabilities {
    return new Capabilities().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Capabilities {
    return new Capabilities().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Capabilities {
    return new Capabilities().fromJsonString(jsonString, options);
  }

  static equals(a: Capabilities | PlainMessage<Capabilities> | undefined, b: Capabilities | PlainMessage<Capabilities> | undefined): boolean {
    return proto3.util.equals(Capabilities, a, b);
  }
}

