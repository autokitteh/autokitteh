// @generated by protoc-gen-es v1.5.1 with parameter "target=ts"
// @generated from file autokitteh/orgs/v1/svc.proto (package autokitteh.orgs.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { Org } from "./org_pb.js";

/**
 * @generated from message autokitteh.orgs.v1.CreateRequest
 */
export class CreateRequest extends Message<CreateRequest> {
  /**
   * @generated from field: autokitteh.orgs.v1.Org org = 1;
   */
  org?: Org;

  constructor(data?: PartialMessage<CreateRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.CreateRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "org", kind: "message", T: Org },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateRequest {
    return new CreateRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateRequest {
    return new CreateRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateRequest {
    return new CreateRequest().fromJsonString(jsonString, options);
  }

  static equals(a: CreateRequest | PlainMessage<CreateRequest> | undefined, b: CreateRequest | PlainMessage<CreateRequest> | undefined): boolean {
    return proto3.util.equals(CreateRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.CreateResponse
 */
export class CreateResponse extends Message<CreateResponse> {
  /**
   * @generated from field: string org_id = 1;
   */
  orgId = "";

  constructor(data?: PartialMessage<CreateResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.CreateResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "org_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateResponse {
    return new CreateResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateResponse {
    return new CreateResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateResponse {
    return new CreateResponse().fromJsonString(jsonString, options);
  }

  static equals(a: CreateResponse | PlainMessage<CreateResponse> | undefined, b: CreateResponse | PlainMessage<CreateResponse> | undefined): boolean {
    return proto3.util.equals(CreateResponse, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.GetRequest
 */
export class GetRequest extends Message<GetRequest> {
  /**
   * @generated from field: string org_id = 1;
   */
  orgId = "";

  constructor(data?: PartialMessage<GetRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.GetRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "org_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
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
 * @generated from message autokitteh.orgs.v1.GetResponse
 */
export class GetResponse extends Message<GetResponse> {
  /**
   * @generated from field: autokitteh.orgs.v1.Org org = 1;
   */
  org?: Org;

  constructor(data?: PartialMessage<GetResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.GetResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "org", kind: "message", T: Org },
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
 * @generated from message autokitteh.orgs.v1.UpdateRequest
 */
export class UpdateRequest extends Message<UpdateRequest> {
  /**
   * @generated from field: autokitteh.orgs.v1.Org org = 1;
   */
  org?: Org;

  constructor(data?: PartialMessage<UpdateRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.UpdateRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "org", kind: "message", T: Org },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpdateRequest {
    return new UpdateRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpdateRequest {
    return new UpdateRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpdateRequest {
    return new UpdateRequest().fromJsonString(jsonString, options);
  }

  static equals(a: UpdateRequest | PlainMessage<UpdateRequest> | undefined, b: UpdateRequest | PlainMessage<UpdateRequest> | undefined): boolean {
    return proto3.util.equals(UpdateRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.UpdateResponse
 */
export class UpdateResponse extends Message<UpdateResponse> {
  constructor(data?: PartialMessage<UpdateResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.UpdateResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpdateResponse {
    return new UpdateResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpdateResponse {
    return new UpdateResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpdateResponse {
    return new UpdateResponse().fromJsonString(jsonString, options);
  }

  static equals(a: UpdateResponse | PlainMessage<UpdateResponse> | undefined, b: UpdateResponse | PlainMessage<UpdateResponse> | undefined): boolean {
    return proto3.util.equals(UpdateResponse, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.AddMemberRequest
 */
export class AddMemberRequest extends Message<AddMemberRequest> {
  /**
   * @generated from field: string user_id = 1;
   */
  userId = "";

  /**
   * @generated from field: string org_id = 2;
   */
  orgId = "";

  constructor(data?: PartialMessage<AddMemberRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.AddMemberRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "user_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "org_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AddMemberRequest {
    return new AddMemberRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AddMemberRequest {
    return new AddMemberRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AddMemberRequest {
    return new AddMemberRequest().fromJsonString(jsonString, options);
  }

  static equals(a: AddMemberRequest | PlainMessage<AddMemberRequest> | undefined, b: AddMemberRequest | PlainMessage<AddMemberRequest> | undefined): boolean {
    return proto3.util.equals(AddMemberRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.AddMemberResponse
 */
export class AddMemberResponse extends Message<AddMemberResponse> {
  constructor(data?: PartialMessage<AddMemberResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.AddMemberResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AddMemberResponse {
    return new AddMemberResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AddMemberResponse {
    return new AddMemberResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AddMemberResponse {
    return new AddMemberResponse().fromJsonString(jsonString, options);
  }

  static equals(a: AddMemberResponse | PlainMessage<AddMemberResponse> | undefined, b: AddMemberResponse | PlainMessage<AddMemberResponse> | undefined): boolean {
    return proto3.util.equals(AddMemberResponse, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.RemoveMemberRequest
 */
export class RemoveMemberRequest extends Message<RemoveMemberRequest> {
  /**
   * @generated from field: string user_id = 1;
   */
  userId = "";

  /**
   * @generated from field: string org_id = 2;
   */
  orgId = "";

  constructor(data?: PartialMessage<RemoveMemberRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.RemoveMemberRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "user_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "org_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): RemoveMemberRequest {
    return new RemoveMemberRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): RemoveMemberRequest {
    return new RemoveMemberRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): RemoveMemberRequest {
    return new RemoveMemberRequest().fromJsonString(jsonString, options);
  }

  static equals(a: RemoveMemberRequest | PlainMessage<RemoveMemberRequest> | undefined, b: RemoveMemberRequest | PlainMessage<RemoveMemberRequest> | undefined): boolean {
    return proto3.util.equals(RemoveMemberRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.RemoveMemberResponse
 */
export class RemoveMemberResponse extends Message<RemoveMemberResponse> {
  constructor(data?: PartialMessage<RemoveMemberResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.RemoveMemberResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): RemoveMemberResponse {
    return new RemoveMemberResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): RemoveMemberResponse {
    return new RemoveMemberResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): RemoveMemberResponse {
    return new RemoveMemberResponse().fromJsonString(jsonString, options);
  }

  static equals(a: RemoveMemberResponse | PlainMessage<RemoveMemberResponse> | undefined, b: RemoveMemberResponse | PlainMessage<RemoveMemberResponse> | undefined): boolean {
    return proto3.util.equals(RemoveMemberResponse, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.ListMembersRequest
 */
export class ListMembersRequest extends Message<ListMembersRequest> {
  /**
   * @generated from field: string org_id = 1;
   */
  orgId = "";

  constructor(data?: PartialMessage<ListMembersRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.ListMembersRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "org_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListMembersRequest {
    return new ListMembersRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListMembersRequest {
    return new ListMembersRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListMembersRequest {
    return new ListMembersRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ListMembersRequest | PlainMessage<ListMembersRequest> | undefined, b: ListMembersRequest | PlainMessage<ListMembersRequest> | undefined): boolean {
    return proto3.util.equals(ListMembersRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.ListMembersResponse
 */
export class ListMembersResponse extends Message<ListMembersResponse> {
  /**
   * @generated from field: repeated string user_ids = 1;
   */
  userIds: string[] = [];

  constructor(data?: PartialMessage<ListMembersResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.ListMembersResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "user_ids", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListMembersResponse {
    return new ListMembersResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListMembersResponse {
    return new ListMembersResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListMembersResponse {
    return new ListMembersResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ListMembersResponse | PlainMessage<ListMembersResponse> | undefined, b: ListMembersResponse | PlainMessage<ListMembersResponse> | undefined): boolean {
    return proto3.util.equals(ListMembersResponse, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.IsMemberRequest
 */
export class IsMemberRequest extends Message<IsMemberRequest> {
  /**
   * @generated from field: string user_id = 1;
   */
  userId = "";

  /**
   * @generated from field: string org_id = 2;
   */
  orgId = "";

  constructor(data?: PartialMessage<IsMemberRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.IsMemberRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "user_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "org_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): IsMemberRequest {
    return new IsMemberRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): IsMemberRequest {
    return new IsMemberRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): IsMemberRequest {
    return new IsMemberRequest().fromJsonString(jsonString, options);
  }

  static equals(a: IsMemberRequest | PlainMessage<IsMemberRequest> | undefined, b: IsMemberRequest | PlainMessage<IsMemberRequest> | undefined): boolean {
    return proto3.util.equals(IsMemberRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.IsMemberResponse
 */
export class IsMemberResponse extends Message<IsMemberResponse> {
  /**
   * @generated from field: bool is_member = 1;
   */
  isMember = false;

  constructor(data?: PartialMessage<IsMemberResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.IsMemberResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "is_member", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): IsMemberResponse {
    return new IsMemberResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): IsMemberResponse {
    return new IsMemberResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): IsMemberResponse {
    return new IsMemberResponse().fromJsonString(jsonString, options);
  }

  static equals(a: IsMemberResponse | PlainMessage<IsMemberResponse> | undefined, b: IsMemberResponse | PlainMessage<IsMemberResponse> | undefined): boolean {
    return proto3.util.equals(IsMemberResponse, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.GetOrgsForUserRequest
 */
export class GetOrgsForUserRequest extends Message<GetOrgsForUserRequest> {
  /**
   * @generated from field: string user_id = 1;
   */
  userId = "";

  constructor(data?: PartialMessage<GetOrgsForUserRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.GetOrgsForUserRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "user_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetOrgsForUserRequest {
    return new GetOrgsForUserRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetOrgsForUserRequest {
    return new GetOrgsForUserRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetOrgsForUserRequest {
    return new GetOrgsForUserRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetOrgsForUserRequest | PlainMessage<GetOrgsForUserRequest> | undefined, b: GetOrgsForUserRequest | PlainMessage<GetOrgsForUserRequest> | undefined): boolean {
    return proto3.util.equals(GetOrgsForUserRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.GetOrgsForUserResponse
 */
export class GetOrgsForUserResponse extends Message<GetOrgsForUserResponse> {
  /**
   * @generated from field: repeated autokitteh.orgs.v1.Org orgs = 1;
   */
  orgs: Org[] = [];

  constructor(data?: PartialMessage<GetOrgsForUserResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.GetOrgsForUserResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "orgs", kind: "message", T: Org, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetOrgsForUserResponse {
    return new GetOrgsForUserResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetOrgsForUserResponse {
    return new GetOrgsForUserResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetOrgsForUserResponse {
    return new GetOrgsForUserResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetOrgsForUserResponse | PlainMessage<GetOrgsForUserResponse> | undefined, b: GetOrgsForUserResponse | PlainMessage<GetOrgsForUserResponse> | undefined): boolean {
    return proto3.util.equals(GetOrgsForUserResponse, a, b);
  }
}

