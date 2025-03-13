// @generated by protoc-gen-es v1.5.1 with parameter "target=ts"
// @generated from file autokitteh/oauth/v1/svc.proto (package autokitteh.oauth.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, protoInt64 } from "@bufbuild/protobuf";

/**
 * @generated from message autokitteh.oauth.v1.GetRequest
 */
export class GetRequest extends Message<GetRequest> {
  /**
   * @generated from field: string id = 1;
   */
  id = "";

  constructor(data?: PartialMessage<GetRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.oauth.v1.GetRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
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
 * @generated from message autokitteh.oauth.v1.GetResponse
 */
export class GetResponse extends Message<GetResponse> {
  /**
   * @generated from field: autokitteh.oauth.v1.OAuthConfig config = 1;
   */
  config?: OAuthConfig;

  constructor(data?: PartialMessage<GetResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.oauth.v1.GetResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "config", kind: "message", T: OAuthConfig },
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
 * @generated from message autokitteh.oauth.v1.StartFlowRequest
 */
export class StartFlowRequest extends Message<StartFlowRequest> {
  /**
   * @generated from field: string integration = 1;
   */
  integration = "";

  /**
   * @generated from field: string connection_id = 2;
   */
  connectionId = "";

  /**
   * @generated from field: string origin = 3;
   */
  origin = "";

  constructor(data?: PartialMessage<StartFlowRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.oauth.v1.StartFlowRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "integration", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "connection_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "origin", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): StartFlowRequest {
    return new StartFlowRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): StartFlowRequest {
    return new StartFlowRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): StartFlowRequest {
    return new StartFlowRequest().fromJsonString(jsonString, options);
  }

  static equals(a: StartFlowRequest | PlainMessage<StartFlowRequest> | undefined, b: StartFlowRequest | PlainMessage<StartFlowRequest> | undefined): boolean {
    return proto3.util.equals(StartFlowRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.oauth.v1.StartFlowResponse
 */
export class StartFlowResponse extends Message<StartFlowResponse> {
  /**
   * @generated from field: string url = 1;
   */
  url = "";

  constructor(data?: PartialMessage<StartFlowResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.oauth.v1.StartFlowResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "url", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): StartFlowResponse {
    return new StartFlowResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): StartFlowResponse {
    return new StartFlowResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): StartFlowResponse {
    return new StartFlowResponse().fromJsonString(jsonString, options);
  }

  static equals(a: StartFlowResponse | PlainMessage<StartFlowResponse> | undefined, b: StartFlowResponse | PlainMessage<StartFlowResponse> | undefined): boolean {
    return proto3.util.equals(StartFlowResponse, a, b);
  }
}

/**
 * @generated from message autokitteh.oauth.v1.ExchangeRequest
 */
export class ExchangeRequest extends Message<ExchangeRequest> {
  /**
   * @generated from field: string integration = 1;
   */
  integration = "";

  /**
   * @generated from field: string connection_id = 2;
   */
  connectionId = "";

  /**
   * @generated from field: string code = 3;
   */
  code = "";

  constructor(data?: PartialMessage<ExchangeRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.oauth.v1.ExchangeRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "integration", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "connection_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "code", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ExchangeRequest {
    return new ExchangeRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ExchangeRequest {
    return new ExchangeRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ExchangeRequest {
    return new ExchangeRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ExchangeRequest | PlainMessage<ExchangeRequest> | undefined, b: ExchangeRequest | PlainMessage<ExchangeRequest> | undefined): boolean {
    return proto3.util.equals(ExchangeRequest, a, b);
  }
}

/**
 * @generated from message autokitteh.oauth.v1.ExchangeResponse
 */
export class ExchangeResponse extends Message<ExchangeResponse> {
  /**
   * @generated from field: string access_token = 1;
   */
  accessToken = "";

  /**
   * @generated from field: string refresh_token = 2;
   */
  refreshToken = "";

  /**
   * @generated from field: string token_type = 3;
   */
  tokenType = "";

  /**
   * @generated from field: int64 expiry = 4;
   */
  expiry = protoInt64.zero;

  constructor(data?: PartialMessage<ExchangeResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.oauth.v1.ExchangeResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "access_token", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "refresh_token", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "token_type", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "expiry", kind: "scalar", T: 3 /* ScalarType.INT64 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ExchangeResponse {
    return new ExchangeResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ExchangeResponse {
    return new ExchangeResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ExchangeResponse {
    return new ExchangeResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ExchangeResponse | PlainMessage<ExchangeResponse> | undefined, b: ExchangeResponse | PlainMessage<ExchangeResponse> | undefined): boolean {
    return proto3.util.equals(ExchangeResponse, a, b);
  }
}

/**
 * @generated from message autokitteh.oauth.v1.OAuthConfig
 */
export class OAuthConfig extends Message<OAuthConfig> {
  /**
   * @generated from field: string client_id = 1;
   */
  clientId = "";

  /**
   * @generated from field: string client_secret = 2;
   */
  clientSecret = "";

  /**
   * @generated from field: string auth_url = 3;
   */
  authUrl = "";

  /**
   * @generated from field: string device_auth_url = 4;
   */
  deviceAuthUrl = "";

  /**
   * @generated from field: string token_url = 5;
   */
  tokenUrl = "";

  /**
   * @generated from field: string redirect_url = 6;
   */
  redirectUrl = "";

  /**
   * https://pkg.go.dev/golang.org/x/oauth2#AuthStyle
   *
   * @generated from field: int32 auth_style = 7;
   */
  authStyle = 0;

  /**
   * https://pkg.go.dev/golang.org/x/oauth2#AuthCodeOption
   *
   * @generated from field: map<string, string> options = 8;
   */
  options: { [key: string]: string } = {};

  /**
   * @generated from field: repeated string scopes = 9;
   */
  scopes: string[] = [];

  constructor(data?: PartialMessage<OAuthConfig>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "autokitteh.oauth.v1.OAuthConfig";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "client_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "client_secret", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "auth_url", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "device_auth_url", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "token_url", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 6, name: "redirect_url", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 7, name: "auth_style", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 8, name: "options", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 9 /* ScalarType.STRING */} },
    { no: 9, name: "scopes", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): OAuthConfig {
    return new OAuthConfig().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): OAuthConfig {
    return new OAuthConfig().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): OAuthConfig {
    return new OAuthConfig().fromJsonString(jsonString, options);
  }

  static equals(a: OAuthConfig | PlainMessage<OAuthConfig> | undefined, b: OAuthConfig | PlainMessage<OAuthConfig> | undefined): boolean {
    return proto3.util.equals(OAuthConfig, a, b);
  }
}

