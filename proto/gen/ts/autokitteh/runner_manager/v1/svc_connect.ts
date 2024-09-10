// @generated by protoc-gen-connect-es v1.1.4 with parameter "target=ts"
// @generated from file autokitteh/runner_manager/v1/svc.proto (package autokitteh.runner_manager.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { HealthRequest, HealthResponse, RunnerHealthRequest, RunnerHealthResponse, StartRequest, StartResponse, StopRequest, StopResponse } from "./svc_pb.js";
import { MethodKind } from "@bufbuild/protobuf";

/**
 * @generated from service autokitteh.runner_manager.v1.RunnerManagerService
 */
export const RunnerManagerService = {
  typeName: "autokitteh.runner_manager.v1.RunnerManagerService",
  methods: {
    /**
     * @generated from rpc autokitteh.runner_manager.v1.RunnerManagerService.Start
     */
    start: {
      name: "Start",
      I: StartRequest,
      O: StartResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc autokitteh.runner_manager.v1.RunnerManagerService.RunnerHealth
     */
    runnerHealth: {
      name: "RunnerHealth",
      I: RunnerHealthRequest,
      O: RunnerHealthResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc autokitteh.runner_manager.v1.RunnerManagerService.Stop
     */
    stop: {
      name: "Stop",
      I: StopRequest,
      O: StopResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc autokitteh.runner_manager.v1.RunnerManagerService.Health
     */
    health: {
      name: "Health",
      I: HealthRequest,
      O: HealthResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;
