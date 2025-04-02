// @generated by protoc-gen-connect-es v1.1.4 with parameter "target=ts"
// @generated from file autokitteh/runner_manager/v1/runner_manager_svc.proto (package autokitteh.runner_manager.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { HealthRequest, HealthResponse, RunnerHealthRequest, RunnerHealthResponse, StartRunnerRequest, StartRunnerResponse, StopRunnerRequest, StopRunnerResponse } from "./runner_manager_svc_pb";
import { MethodKind } from "@bufbuild/protobuf";

/**
 * @generated from service autokitteh.runner_manager.v1.RunnerManagerService
 */
export const RunnerManagerService = {
  typeName: "autokitteh.runner_manager.v1.RunnerManagerService",
  methods: {
    /**
     * @generated from rpc autokitteh.runner_manager.v1.RunnerManagerService.StartRunner
     */
    startRunner: {
      name: "StartRunner",
      I: StartRunnerRequest,
      O: StartRunnerResponse,
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
     * @generated from rpc autokitteh.runner_manager.v1.RunnerManagerService.StopRunner
     */
    stopRunner: {
      name: "StopRunner",
      I: StopRunnerRequest,
      O: StopRunnerResponse,
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

