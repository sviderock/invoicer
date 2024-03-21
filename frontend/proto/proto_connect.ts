// @generated by protoc-gen-connect-es v1.4.0 with parameter "target=ts,import_extension=.ts"
// @generated from file proto.proto (package proto, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { FailRequest, FailResponse, PingRequest, PingResponse } from "./proto_pb.ts";
import { MethodIdempotency, MethodKind } from "@bufbuild/protobuf";

/**
 * @generated from service proto.PingService
 */
export const PingService = {
  typeName: "proto.PingService",
  methods: {
    /**
     * @generated from rpc proto.PingService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * @generated from rpc proto.PingService.Fail
     */
    fail: {
      name: "Fail",
      I: FailRequest,
      O: FailResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;
