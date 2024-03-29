// @generated by protoc-gen-connect-es v1.1.4 with parameter "target=ts"
// @generated from file autokitteh/envs/v1/svc.proto (package autokitteh.envs.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { CreateRequest, CreateResponse, GetRequest, GetResponse, GetVarsRequest, GetVarsResponse, ListRequest, ListResponse, RemoveRequest, RemoveResponse, RemoveVarRequest, RemoveVarResponse, RevealVarRequest, RevealVarResponse, SetVarRequest, SetVarResponse, UpdateRequest, UpdateResponse } from "./svc_pb.js";
import { MethodKind } from "@bufbuild/protobuf";

/**
 * @generated from service autokitteh.envs.v1.EnvsService
 */
export const EnvsService = {
  typeName: "autokitteh.envs.v1.EnvsService",
  methods: {
    /**
     * @generated from rpc autokitteh.envs.v1.EnvsService.List
     */
    list: {
      name: "List",
      I: ListRequest,
      O: ListResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc autokitteh.envs.v1.EnvsService.Create
     */
    create: {
      name: "Create",
      I: CreateRequest,
      O: CreateResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc autokitteh.envs.v1.EnvsService.Get
     */
    get: {
      name: "Get",
      I: GetRequest,
      O: GetResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc autokitteh.envs.v1.EnvsService.Remove
     */
    remove: {
      name: "Remove",
      I: RemoveRequest,
      O: RemoveResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc autokitteh.envs.v1.EnvsService.Update
     */
    update: {
      name: "Update",
      I: UpdateRequest,
      O: UpdateResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc autokitteh.envs.v1.EnvsService.SetVar
     */
    setVar: {
      name: "SetVar",
      I: SetVarRequest,
      O: SetVarResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc autokitteh.envs.v1.EnvsService.RemoveVar
     */
    removeVar: {
      name: "RemoveVar",
      I: RemoveVarRequest,
      O: RemoveVarResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc autokitteh.envs.v1.EnvsService.GetVars
     */
    getVars: {
      name: "GetVars",
      I: GetVarsRequest,
      O: GetVarsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc autokitteh.envs.v1.EnvsService.RevealVar
     */
    revealVar: {
      name: "RevealVar",
      I: RevealVarRequest,
      O: RevealVarResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

