// Copyright 2023 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// @generated by protoc-gen-es v2.2.3 with parameter "target=ts"
// @generated from file buf/validate/priv/private.proto (package buf.validate.priv, syntax proto3)
/* eslint-disable */

import type { GenExtension, GenFile, GenMessage } from "@bufbuild/protobuf/codegenv1";
import { extDesc, fileDesc, messageDesc } from "@bufbuild/protobuf/codegenv1";
import type { FieldOptions } from "@bufbuild/protobuf/wkt";
import { file_google_protobuf_descriptor } from "@bufbuild/protobuf/wkt";
import type { Message } from "@bufbuild/protobuf";

/**
 * Describes the file buf/validate/priv/private.proto.
 */
export const file_buf_validate_priv_private: GenFile = /*@__PURE__*/
  fileDesc("Ch9idWYvdmFsaWRhdGUvcHJpdi9wcml2YXRlLnByb3RvEhFidWYudmFsaWRhdGUucHJpdiI+ChBGaWVsZENvbnN0cmFpbnRzEioKA2NlbBgBIAMoCzIdLmJ1Zi52YWxpZGF0ZS5wcml2LkNvbnN0cmFpbnQiPQoKQ29uc3RyYWludBIKCgJpZBgBIAEoCRIPCgdtZXNzYWdlGAIgASgJEhIKCmV4cHJlc3Npb24YAyABKAk6XQoFZmllbGQSHS5nb29nbGUucHJvdG9idWYuRmllbGRPcHRpb25zGICPAyABKAsyIy5idWYudmFsaWRhdGUucHJpdi5GaWVsZENvbnN0cmFpbnRzUgVmaWVsZIgBAULZAQoVY29tLmJ1Zi52YWxpZGF0ZS5wcml2QgxQcml2YXRlUHJvdG9QAVpMYnVmLmJ1aWxkL2dlbi9nby9idWZidWlsZC9wcm90b3ZhbGlkYXRlL3Byb3RvY29sYnVmZmVycy9nby9idWYvdmFsaWRhdGUvcHJpdqICA0JWUKoCEUJ1Zi5WYWxpZGF0ZS5Qcml2ygIRQnVmXFZhbGlkYXRlXFByaXbiAh1CdWZcVmFsaWRhdGVcUHJpdlxHUEJNZXRhZGF0YeoCE0J1Zjo6VmFsaWRhdGU6OlByaXZiBnByb3RvMw", [file_google_protobuf_descriptor]);

/**
 * Do not use. Internal to protovalidate library
 *
 * @generated from message buf.validate.priv.FieldConstraints
 */
export type FieldConstraints = Message<"buf.validate.priv.FieldConstraints"> & {
  /**
   * @generated from field: repeated buf.validate.priv.Constraint cel = 1;
   */
  cel: Constraint[];
};

/**
 * Describes the message buf.validate.priv.FieldConstraints.
 * Use `create(FieldConstraintsSchema)` to create a new message.
 */
export const FieldConstraintsSchema: GenMessage<FieldConstraints> = /*@__PURE__*/
  messageDesc(file_buf_validate_priv_private, 0);

/**
 * Do not use. Internal to protovalidate library
 *
 * @generated from message buf.validate.priv.Constraint
 */
export type Constraint = Message<"buf.validate.priv.Constraint"> & {
  /**
   * @generated from field: string id = 1;
   */
  id: string;

  /**
   * @generated from field: string message = 2;
   */
  message: string;

  /**
   * @generated from field: string expression = 3;
   */
  expression: string;
};

/**
 * Describes the message buf.validate.priv.Constraint.
 * Use `create(ConstraintSchema)` to create a new message.
 */
export const ConstraintSchema: GenMessage<Constraint> = /*@__PURE__*/
  messageDesc(file_buf_validate_priv_private, 1);

/**
 * Do not use. Internal to protovalidate library
 *
 * @generated from extension: optional buf.validate.priv.FieldConstraints field = 51072;
 */
export const field: GenExtension<FieldOptions, FieldConstraints> = /*@__PURE__*/
  extDesc(file_buf_validate_priv_private, 0);

