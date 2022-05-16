/**
 * @fileoverview gRPC-Web generated client stub for autokitteh.langsvc
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


var google_api_annotations_pb = require('../google/api/annotations_pb.js')

var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js')

var validate_validate_pb = require('../validate/validate_pb.js')

var lang_run_pb = require('../lang/run_pb.js')

var program_error_pb = require('../program/error_pb.js')

var program_program_pb = require('../program/program_pb.js')

var values_values_pb = require('../values/values_pb.js')
const proto = {};
proto.autokitteh = {};
proto.autokitteh.langsvc = require('./langrunsvc_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.langsvc.LangRunClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.langsvc.LangRunPromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.langsvc.RunRequest,
 *   !proto.autokitteh.langsvc.RunUpdate>}
 */
const methodDescriptor_LangRun_Run = new grpc.web.MethodDescriptor(
  '/autokitteh.langsvc.LangRun/Run',
  grpc.web.MethodType.SERVER_STREAMING,
  proto.autokitteh.langsvc.RunRequest,
  proto.autokitteh.langsvc.RunUpdate,
  /**
   * @param {!proto.autokitteh.langsvc.RunRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.langsvc.RunUpdate.deserializeBinary
);


/**
 * @param {!proto.autokitteh.langsvc.RunRequest} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.langsvc.RunUpdate>}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.langsvc.LangRunClient.prototype.run =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/autokitteh.langsvc.LangRun/Run',
      request,
      metadata || {},
      methodDescriptor_LangRun_Run);
};


/**
 * @param {!proto.autokitteh.langsvc.RunRequest} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.langsvc.RunUpdate>}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.langsvc.LangRunPromiseClient.prototype.run =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/autokitteh.langsvc.LangRun/Run',
      request,
      metadata || {},
      methodDescriptor_LangRun_Run);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.langsvc.CallFunctionRequest,
 *   !proto.autokitteh.langsvc.RunUpdate>}
 */
const methodDescriptor_LangRun_CallFunction = new grpc.web.MethodDescriptor(
  '/autokitteh.langsvc.LangRun/CallFunction',
  grpc.web.MethodType.SERVER_STREAMING,
  proto.autokitteh.langsvc.CallFunctionRequest,
  proto.autokitteh.langsvc.RunUpdate,
  /**
   * @param {!proto.autokitteh.langsvc.CallFunctionRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.langsvc.RunUpdate.deserializeBinary
);


/**
 * @param {!proto.autokitteh.langsvc.CallFunctionRequest} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.langsvc.RunUpdate>}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.langsvc.LangRunClient.prototype.callFunction =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/autokitteh.langsvc.LangRun/CallFunction',
      request,
      metadata || {},
      methodDescriptor_LangRun_CallFunction);
};


/**
 * @param {!proto.autokitteh.langsvc.CallFunctionRequest} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.langsvc.RunUpdate>}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.langsvc.LangRunPromiseClient.prototype.callFunction =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/autokitteh.langsvc.LangRun/CallFunction',
      request,
      metadata || {},
      methodDescriptor_LangRun_CallFunction);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.langsvc.RunGetRequest,
 *   !proto.autokitteh.langsvc.RunGetResponse>}
 */
const methodDescriptor_LangRun_RunGet = new grpc.web.MethodDescriptor(
  '/autokitteh.langsvc.LangRun/RunGet',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.langsvc.RunGetRequest,
  proto.autokitteh.langsvc.RunGetResponse,
  /**
   * @param {!proto.autokitteh.langsvc.RunGetRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.langsvc.RunGetResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.langsvc.RunGetRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.langsvc.RunGetResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.langsvc.RunGetResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.langsvc.LangRunClient.prototype.runGet =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.langsvc.LangRun/RunGet',
      request,
      metadata || {},
      methodDescriptor_LangRun_RunGet,
      callback);
};


/**
 * @param {!proto.autokitteh.langsvc.RunGetRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.langsvc.RunGetResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.langsvc.LangRunPromiseClient.prototype.runGet =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.langsvc.LangRun/RunGet',
      request,
      metadata || {},
      methodDescriptor_LangRun_RunGet);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.langsvc.RunCallReturnRequest,
 *   !proto.autokitteh.langsvc.RunCallReturnResponse>}
 */
const methodDescriptor_LangRun_RunCallReturn = new grpc.web.MethodDescriptor(
  '/autokitteh.langsvc.LangRun/RunCallReturn',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.langsvc.RunCallReturnRequest,
  proto.autokitteh.langsvc.RunCallReturnResponse,
  /**
   * @param {!proto.autokitteh.langsvc.RunCallReturnRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.langsvc.RunCallReturnResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.langsvc.RunCallReturnRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.langsvc.RunCallReturnResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.langsvc.RunCallReturnResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.langsvc.LangRunClient.prototype.runCallReturn =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.langsvc.LangRun/RunCallReturn',
      request,
      metadata || {},
      methodDescriptor_LangRun_RunCallReturn,
      callback);
};


/**
 * @param {!proto.autokitteh.langsvc.RunCallReturnRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.langsvc.RunCallReturnResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.langsvc.LangRunPromiseClient.prototype.runCallReturn =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.langsvc.LangRun/RunCallReturn',
      request,
      metadata || {},
      methodDescriptor_LangRun_RunCallReturn);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.langsvc.RunLoadReturnRequest,
 *   !proto.autokitteh.langsvc.RunLoadReturnResponse>}
 */
const methodDescriptor_LangRun_RunLoadReturn = new grpc.web.MethodDescriptor(
  '/autokitteh.langsvc.LangRun/RunLoadReturn',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.langsvc.RunLoadReturnRequest,
  proto.autokitteh.langsvc.RunLoadReturnResponse,
  /**
   * @param {!proto.autokitteh.langsvc.RunLoadReturnRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.langsvc.RunLoadReturnResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.langsvc.RunLoadReturnRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.langsvc.RunLoadReturnResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.langsvc.RunLoadReturnResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.langsvc.LangRunClient.prototype.runLoadReturn =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.langsvc.LangRun/RunLoadReturn',
      request,
      metadata || {},
      methodDescriptor_LangRun_RunLoadReturn,
      callback);
};


/**
 * @param {!proto.autokitteh.langsvc.RunLoadReturnRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.langsvc.RunLoadReturnResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.langsvc.LangRunPromiseClient.prototype.runLoadReturn =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.langsvc.LangRun/RunLoadReturn',
      request,
      metadata || {},
      methodDescriptor_LangRun_RunLoadReturn);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.langsvc.RunCancelRequest,
 *   !proto.autokitteh.langsvc.RunCancelResponse>}
 */
const methodDescriptor_LangRun_RunCancel = new grpc.web.MethodDescriptor(
  '/autokitteh.langsvc.LangRun/RunCancel',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.langsvc.RunCancelRequest,
  proto.autokitteh.langsvc.RunCancelResponse,
  /**
   * @param {!proto.autokitteh.langsvc.RunCancelRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.langsvc.RunCancelResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.langsvc.RunCancelRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.langsvc.RunCancelResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.langsvc.RunCancelResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.langsvc.LangRunClient.prototype.runCancel =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.langsvc.LangRun/RunCancel',
      request,
      metadata || {},
      methodDescriptor_LangRun_RunCancel,
      callback);
};


/**
 * @param {!proto.autokitteh.langsvc.RunCancelRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.langsvc.RunCancelResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.langsvc.LangRunPromiseClient.prototype.runCancel =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.langsvc.LangRun/RunCancel',
      request,
      metadata || {},
      methodDescriptor_LangRun_RunCancel);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.langsvc.ListRunsRequest,
 *   !proto.autokitteh.langsvc.ListRunsResponse>}
 */
const methodDescriptor_LangRun_ListRuns = new grpc.web.MethodDescriptor(
  '/autokitteh.langsvc.LangRun/ListRuns',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.langsvc.ListRunsRequest,
  proto.autokitteh.langsvc.ListRunsResponse,
  /**
   * @param {!proto.autokitteh.langsvc.ListRunsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.langsvc.ListRunsResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.langsvc.ListRunsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.langsvc.ListRunsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.langsvc.ListRunsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.langsvc.LangRunClient.prototype.listRuns =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.langsvc.LangRun/ListRuns',
      request,
      metadata || {},
      methodDescriptor_LangRun_ListRuns,
      callback);
};


/**
 * @param {!proto.autokitteh.langsvc.ListRunsRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.langsvc.ListRunsResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.langsvc.LangRunPromiseClient.prototype.listRuns =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.langsvc.LangRun/ListRuns',
      request,
      metadata || {},
      methodDescriptor_LangRun_ListRuns);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.langsvc.RunDiscardRequest,
 *   !proto.autokitteh.langsvc.RunDiscardResponse>}
 */
const methodDescriptor_LangRun_RunDiscard = new grpc.web.MethodDescriptor(
  '/autokitteh.langsvc.LangRun/RunDiscard',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.langsvc.RunDiscardRequest,
  proto.autokitteh.langsvc.RunDiscardResponse,
  /**
   * @param {!proto.autokitteh.langsvc.RunDiscardRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.langsvc.RunDiscardResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.langsvc.RunDiscardRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.langsvc.RunDiscardResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.langsvc.RunDiscardResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.langsvc.LangRunClient.prototype.runDiscard =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.langsvc.LangRun/RunDiscard',
      request,
      metadata || {},
      methodDescriptor_LangRun_RunDiscard,
      callback);
};


/**
 * @param {!proto.autokitteh.langsvc.RunDiscardRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.langsvc.RunDiscardResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.langsvc.LangRunPromiseClient.prototype.runDiscard =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.langsvc.LangRun/RunDiscard',
      request,
      metadata || {},
      methodDescriptor_LangRun_RunDiscard);
};


module.exports = proto.autokitteh.langsvc;

