// @generated by protoc-gen-es v1.5.1 with parameter "target=ts"
// @generated from file autokitteh/orgs/v1/svc.proto (package autokitteh.orgs.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { FieldMask, Message, proto3 } from "@bufbuild/protobuf";
import { Org } from "./org_pb.js";
import { User } from "../../users/v1/user_pb.js";

/**
 * @generated from message autokitteh.orgs.v1.CreateRequest
 */
export class CreateRequest extends Message<CreateRequest> {
  /**
   * org.org_id is ignored.
   *
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
 * @generated from message autokitteh.orgs.v1.DeleteRequest
 */
export class DeleteRequest extends Message<DeleteRequest> {
  /**
   * @generated from field: string org_id = 1;
   */
  orgId = "";

  constructor(data?: PartialMessage<DeleteRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.DeleteRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "org_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteRequest {
    return new DeleteRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteRequest {
    return new DeleteRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteRequest {
    return new DeleteRequest().fromJsonString(jsonString, options);
  }

  static equals(a: DeleteRequest | PlainMessage<DeleteRequest> | undefined, b: DeleteRequest | PlainMessage<DeleteRequest> | undefined): boolean {
    return proto3.util.equals(DeleteRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.DeleteResponse
 */
export class DeleteResponse extends Message<DeleteResponse> {
  constructor(data?: PartialMessage<DeleteResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.DeleteResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteResponse {
    return new DeleteResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteResponse {
    return new DeleteResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteResponse {
    return new DeleteResponse().fromJsonString(jsonString, options);
  }

  static equals(a: DeleteResponse | PlainMessage<DeleteResponse> | undefined, b: DeleteResponse | PlainMessage<DeleteResponse> | undefined): boolean {
    return proto3.util.equals(DeleteResponse, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.GetRequest
 */
export class GetRequest extends Message<GetRequest> {
  /**
   * org_id and name are mutually exclusive.
   *
   * @generated from field: string org_id = 1;
   */
  orgId = "";

  /**
   * @generated from field: string name = 2;
   */
  name = "";

  constructor(data?: PartialMessage<GetRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.GetRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "org_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
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
   * empty if not found.
   *
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

  /**
   * @generated from field: google.protobuf.FieldMask field_mask = 2;
   */
  fieldMask?: FieldMask;

  constructor(data?: PartialMessage<UpdateRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.UpdateRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "org", kind: "message", T: Org },
    { no: 2, name: "field_mask", kind: "message", T: FieldMask },
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
   * @generated from field: string org_id = 1;
   */
  orgId = "";

  /**
   * @generated from field: string user_id = 2;
   */
  userId = "";

  constructor(data?: PartialMessage<AddMemberRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.AddMemberRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "org_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "user_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
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
   * @generated from field: string org_id = 1;
   */
  orgId = "";

  /**
   * @generated from field: string user_id = 2;
   */
  userId = "";

  constructor(data?: PartialMessage<RemoveMemberRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.RemoveMemberRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "org_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "user_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
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
   * TODO: pagination.
   *
   * @generated from field: repeated autokitteh.users.v1.User users = 1;
   */
  users: User[] = [];

  constructor(data?: PartialMessage<ListMembersResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.ListMembersResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "users", kind: "message", T: User, repeated: true },
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
   * @generated from field: string org_id = 1;
   */
  orgId = "";

  /**
   * @generated from field: string user_id = 2;
   */
  userId = "";

  constructor(data?: PartialMessage<IsMemberRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.IsMemberRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "org_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "user_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
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
 * @generated from message autokitteh.orgs.v1.ListUserMembershipsRequest
 */
export class ListUserMembershipsRequest extends Message<ListUserMembershipsRequest> {
  /**
   * if empty, taken from auth.
   *
   * @generated from field: string user_id = 1;
   */
  userId = "";

  constructor(data?: PartialMessage<ListUserMembershipsRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.ListUserMembershipsRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "user_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListUserMembershipsRequest {
    return new ListUserMembershipsRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListUserMembershipsRequest {
    return new ListUserMembershipsRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListUserMembershipsRequest {
    return new ListUserMembershipsRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ListUserMembershipsRequest | PlainMessage<ListUserMembershipsRequest> | undefined, b: ListUserMembershipsRequest | PlainMessage<ListUserMembershipsRequest> | undefined): boolean {
    return proto3.util.equals(ListUserMembershipsRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.orgs.v1.ListUserMembershipsResponse
 */
export class ListUserMembershipsResponse extends Message<ListUserMembershipsResponse> {
  /**
   * TODO: pagination.
   *
   * @generated from field: repeated autokitteh.orgs.v1.Org orgs = 1;
   */
  orgs: Org[] = [];

  constructor(data?: PartialMessage<ListUserMembershipsResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.orgs.v1.ListUserMembershipsResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "orgs", kind: "message", T: Org, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListUserMembershipsResponse {
    return new ListUserMembershipsResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListUserMembershipsResponse {
    return new ListUserMembershipsResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListUserMembershipsResponse {
    return new ListUserMembershipsResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ListUserMembershipsResponse | PlainMessage<ListUserMembershipsResponse> | undefined, b: ListUserMembershipsResponse | PlainMessage<ListUserMembershipsResponse> | undefined): boolean {
    return proto3.util.equals(ListUserMembershipsResponse, a, b);
  }
}

