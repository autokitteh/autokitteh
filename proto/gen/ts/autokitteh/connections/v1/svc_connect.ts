// @generated by protoc-gen-connect-es v1.1.4 with parameter "target=ts"
// @generated from file autokitteh/connections/v1/svc.proto (package autokitteh.connections.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { CreateRequest, CreateResponse, DeleteRequest, DeleteResponse, GetRequest, GetResponse, ListRequest, ListResponse, RefreshStatusRequest, RefreshStatusResponse, TestRequest, TestResponse, UpdateRequest, UpdateResponse } from "./svc_pb.js";
import { MethodKind } from "@bufbuild/protobuf";

/**
 * Implemented by the autokitteh server.
 *
 * @generated from service autokitteh.connections.v1.ConnectionsService
 */
export const ConnectionsService = {
  typeName: "autokitteh.connections.v1.ConnectionsService",
  methods: {
    /**
     * Initiated indirectly by an autokitteh user, based on an registered integration.
     *
     * @generated from rpc autokitteh.connections.v1.ConnectionsService.Create
     */
    create: {
      name: "Create",
      I: CreateRequest,
      O: CreateResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc autokitteh.connections.v1.ConnectionsService.Delete
     */
    delete: {
      name: "Delete",
      I: DeleteRequest,
      O: DeleteResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc autokitteh.connections.v1.ConnectionsService.Update
     */
    update: {
      name: "Update",
      I: UpdateRequest,
      O: UpdateResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc autokitteh.connections.v1.ConnectionsService.Get
     */
    get: {
      name: "Get",
      I: GetRequest,
      O: GetResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc autokitteh.connections.v1.ConnectionsService.List
     */
    list: {
      name: "List",
      I: ListRequest,
      O: ListResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Test actively performs an integration test using a connection's configuration.
     * (This in turn calls Integration.TestConnection).
     *
     * @generated from rpc autokitteh.connections.v1.ConnectionsService.Test
     */
    test: {
      name: "Test",
      I: TestRequest,
      O: TestResponse,
      kind: MethodKind.Unary,
    },
    /**
     * RefreshStatus makes the connection query the integration regarding the
     * current connection status. This checks that the connection is configured correctly,
     * but does not perform any actual data transfer.
     * (This in turn calls Integration.GetConnectionStatus).
     *
     * @generated from rpc autokitteh.connections.v1.ConnectionsService.RefreshStatus
     */
    refreshStatus: {
      name: "RefreshStatus",
      I: RefreshStatusRequest,
      O: RefreshStatusResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

