// @generated by protoc-gen-connect-es v1.4.0 with parameter "target=ts,import_extension=.ts"
// @generated from file file_upload/v1/file_upload.proto (package file_upload.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { FileUploadRequest, FileUploadResponse } from "./file_upload_pb.ts";
import { MethodKind } from "@bufbuild/protobuf";

/**
 * @generated from service file_upload.v1.FileUploadService
 */
export const FileUploadService = {
  typeName: "file_upload.v1.FileUploadService",
  methods: {
    /**
     * @generated from rpc file_upload.v1.FileUploadService.UploadFile
     */
    uploadFile: {
      name: "UploadFile",
      I: FileUploadRequest,
      O: FileUploadResponse,
      kind: MethodKind.ClientStreaming,
    },
  }
} as const;
