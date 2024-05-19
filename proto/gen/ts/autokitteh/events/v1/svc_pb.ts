// @generated by protoc-gen-es v1.5.1 with parameter "target=ts"
// @generated from file autokitteh/events/v1/svc.proto (package autokitteh.events.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { Event, EventRecord, EventState } from "./event_pb.js";

/**
 * @generated from message autokitteh.events.v1.SaveRequest
 */
export class SaveRequest extends Message<SaveRequest> {
  /**
   * @generated from field: autokitteh.events.v1.Event event = 1;
   */
  event?: Event;

  constructor(data?: PartialMessage<SaveRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.events.v1.SaveRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "event", kind: "message", T: Event },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SaveRequest {
    return new SaveRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SaveRequest {
    return new SaveRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SaveRequest {
    return new SaveRequest().fromJsonString(jsonString, options);
  }

  static equals(a: SaveRequest | PlainMessage<SaveRequest> | undefined, b: SaveRequest | PlainMessage<SaveRequest> | undefined): boolean {
    return proto3.util.equals(SaveRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.events.v1.SaveResponse
 */
export class SaveResponse extends Message<SaveResponse> {
  /**
   * @generated from field: string event_id = 1;
   */
  eventId = "";

  constructor(data?: PartialMessage<SaveResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.events.v1.SaveResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "event_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SaveResponse {
    return new SaveResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SaveResponse {
    return new SaveResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SaveResponse {
    return new SaveResponse().fromJsonString(jsonString, options);
  }

  static equals(a: SaveResponse | PlainMessage<SaveResponse> | undefined, b: SaveResponse | PlainMessage<SaveResponse> | undefined): boolean {
    return proto3.util.equals(SaveResponse, a, b);
  }
}

/**
 * @generated from message autokitteh.events.v1.GetRequest
 */
export class GetRequest extends Message<GetRequest> {
  /**
   * @generated from field: string event_id = 1;
   */
  eventId = "";

  constructor(data?: PartialMessage<GetRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.events.v1.GetRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "event_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetRequest {
    return new GetRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetRequest {
    return new GetRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetRequest {
    return new GetRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetRequest | PlainMessage<GetRequest> | undefined, b: GetRequest | PlainMessage<GetRequest> | undefined): boolean {
    return proto3.util.equals(GetRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.events.v1.GetResponse
 */
export class GetResponse extends Message<GetResponse> {
  /**
   * @generated from field: autokitteh.events.v1.Event event = 1;
   */
  event?: Event;

  constructor(data?: PartialMessage<GetResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.events.v1.GetResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "event", kind: "message", T: Event },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetResponse {
    return new GetResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetResponse {
    return new GetResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetResponse {
    return new GetResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetResponse | PlainMessage<GetResponse> | undefined, b: GetResponse | PlainMessage<GetResponse> | undefined): boolean {
    return proto3.util.equals(GetResponse, a, b);
  }
}

/**
 * @generated from message autokitteh.events.v1.ListRequest
 */
export class ListRequest extends Message<ListRequest> {
  /**
   * @generated from field: string integration_id = 1;
   */
  integrationId = "";

  /**
   * @generated from field: string connection_id = 2;
   */
  connectionId = "";

  /**
   * @generated from field: string event_type = 3;
   */
  eventType = "";

  /**
   * @generated from field: uint32 max_results = 4;
   */
  maxResults = 0;

  /**
   * @generated from field: string order = 5;
   */
  order = "";

  constructor(data?: PartialMessage<ListRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.events.v1.ListRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "integration_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "connection_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "event_type", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "max_results", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 5, name: "order", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListRequest {
    return new ListRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListRequest {
    return new ListRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListRequest {
    return new ListRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ListRequest | PlainMessage<ListRequest> | undefined, b: ListRequest | PlainMessage<ListRequest> | undefined): boolean {
    return proto3.util.equals(ListRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.events.v1.ListResponse
 */
export class ListResponse extends Message<ListResponse> {
  /**
   * @generated from field: repeated autokitteh.events.v1.Event events = 1;
   */
  events: Event[] = [];

  constructor(data?: PartialMessage<ListResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.events.v1.ListResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "events", kind: "message", T: Event, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListResponse {
    return new ListResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListResponse {
    return new ListResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListResponse {
    return new ListResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ListResponse | PlainMessage<ListResponse> | undefined, b: ListResponse | PlainMessage<ListResponse> | undefined): boolean {
    return proto3.util.equals(ListResponse, a, b);
  }
}

/**
 * @generated from message autokitteh.events.v1.ListEventRecordsRequest
 */
export class ListEventRecordsRequest extends Message<ListEventRecordsRequest> {
  /**
   * @generated from field: string event_id = 1;
   */
  eventId = "";

  /**
   * @generated from field: autokitteh.events.v1.EventState state = 3;
   */
  state = EventState.UNSPECIFIED;

  constructor(data?: PartialMessage<ListEventRecordsRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.events.v1.ListEventRecordsRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "event_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "state", kind: "enum", T: proto3.getEnumType(EventState) },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListEventRecordsRequest {
    return new ListEventRecordsRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListEventRecordsRequest {
    return new ListEventRecordsRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListEventRecordsRequest {
    return new ListEventRecordsRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ListEventRecordsRequest | PlainMessage<ListEventRecordsRequest> | undefined, b: ListEventRecordsRequest | PlainMessage<ListEventRecordsRequest> | undefined): boolean {
    return proto3.util.equals(ListEventRecordsRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.events.v1.ListEventRecordsResponse
 */
export class ListEventRecordsResponse extends Message<ListEventRecordsResponse> {
  /**
   * @generated from field: repeated autokitteh.events.v1.EventRecord records = 1;
   */
  records: EventRecord[] = [];

  constructor(data?: PartialMessage<ListEventRecordsResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.events.v1.ListEventRecordsResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "records", kind: "message", T: EventRecord, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListEventRecordsResponse {
    return new ListEventRecordsResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListEventRecordsResponse {
    return new ListEventRecordsResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListEventRecordsResponse {
    return new ListEventRecordsResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ListEventRecordsResponse | PlainMessage<ListEventRecordsResponse> | undefined, b: ListEventRecordsResponse | PlainMessage<ListEventRecordsResponse> | undefined): boolean {
    return proto3.util.equals(ListEventRecordsResponse, a, b);
  }
}

/**
 * @generated from message autokitteh.events.v1.AddEventRecordRequest
 */
export class AddEventRecordRequest extends Message<AddEventRecordRequest> {
  /**
   * @generated from field: autokitteh.events.v1.EventRecord record = 1;
   */
  record?: EventRecord;

  constructor(data?: PartialMessage<AddEventRecordRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.events.v1.AddEventRecordRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "record", kind: "message", T: EventRecord },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AddEventRecordRequest {
    return new AddEventRecordRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AddEventRecordRequest {
    return new AddEventRecordRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AddEventRecordRequest {
    return new AddEventRecordRequest().fromJsonString(jsonString, options);
  }

  static equals(a: AddEventRecordRequest | PlainMessage<AddEventRecordRequest> | undefined, b: AddEventRecordRequest | PlainMessage<AddEventRecordRequest> | undefined): boolean {
    return proto3.util.equals(AddEventRecordRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.events.v1.AddEventRecordResponse
 */
export class AddEventRecordResponse extends Message<AddEventRecordResponse> {
  constructor(data?: PartialMessage<AddEventRecordResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.events.v1.AddEventRecordResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AddEventRecordResponse {
    return new AddEventRecordResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AddEventRecordResponse {
    return new AddEventRecordResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AddEventRecordResponse {
    return new AddEventRecordResponse().fromJsonString(jsonString, options);
  }

  static equals(a: AddEventRecordResponse | PlainMessage<AddEventRecordResponse> | undefined, b: AddEventRecordResponse | PlainMessage<AddEventRecordResponse> | undefined): boolean {
    return proto3.util.equals(AddEventRecordResponse, a, b);
  }
}

