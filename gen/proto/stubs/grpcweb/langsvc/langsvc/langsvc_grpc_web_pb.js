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

var validate_validate_pb = require('../validate/validate_pb.js')

var program_program_pb = require('../program/program_pb.js')
const proto = {};
proto.autokitteh = {};
proto.autokitteh.langsvc = require('./langsvc_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.langsvc.LangClient =
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
proto.autokitteh.langsvc.LangPromiseClient =
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
 *   !proto.autokitteh.langsvc.ListLangsRequest,
 *   !proto.autokitteh.langsvc.ListLangsResponse>}
 */
const methodDescriptor_Lang_ListLangs = new grpc.web.MethodDescriptor(
  '/autokitteh.langsvc.Lang/ListLangs',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.langsvc.ListLangsRequest,
  proto.autokitteh.langsvc.ListLangsResponse,
  /**
   * @param {!proto.autokitteh.langsvc.ListLangsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.langsvc.ListLangsResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.langsvc.ListLangsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.langsvc.ListLangsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.langsvc.ListLangsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.langsvc.LangClient.prototype.listLangs =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.langsvc.Lang/ListLangs',
      request,
      metadata || {},
      methodDescriptor_Lang_ListLangs,
      callback);
};


/**
 * @param {!proto.autokitteh.langsvc.ListLangsRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.langsvc.ListLangsResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.langsvc.LangPromiseClient.prototype.listLangs =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.langsvc.Lang/ListLangs',
      request,
      metadata || {},
      methodDescriptor_Lang_ListLangs);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.langsvc.IsCompilerVersionSupportedRequest,
 *   !proto.autokitteh.langsvc.IsCompilerVersionSupportedResponse>}
 */
const methodDescriptor_Lang_IsCompilerVersionSupported = new grpc.web.MethodDescriptor(
  '/autokitteh.langsvc.Lang/IsCompilerVersionSupported',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.langsvc.IsCompilerVersionSupportedRequest,
  proto.autokitteh.langsvc.IsCompilerVersionSupportedResponse,
  /**
   * @param {!proto.autokitteh.langsvc.IsCompilerVersionSupportedRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.langsvc.IsCompilerVersionSupportedResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.langsvc.IsCompilerVersionSupportedRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.langsvc.IsCompilerVersionSupportedResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.langsvc.IsCompilerVersionSupportedResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.langsvc.LangClient.prototype.isCompilerVersionSupported =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.langsvc.Lang/IsCompilerVersionSupported',
      request,
      metadata || {},
      methodDescriptor_Lang_IsCompilerVersionSupported,
      callback);
};


/**
 * @param {!proto.autokitteh.langsvc.IsCompilerVersionSupportedRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.langsvc.IsCompilerVersionSupportedResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.langsvc.LangPromiseClient.prototype.isCompilerVersionSupported =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.langsvc.Lang/IsCompilerVersionSupported',
      request,
      metadata || {},
      methodDescriptor_Lang_IsCompilerVersionSupported);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.langsvc.GetModuleDependenciesRequest,
 *   !proto.autokitteh.langsvc.GetModuleDependenciesResponse>}
 */
const methodDescriptor_Lang_GetModuleDependencies = new grpc.web.MethodDescriptor(
  '/autokitteh.langsvc.Lang/GetModuleDependencies',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.langsvc.GetModuleDependenciesRequest,
  proto.autokitteh.langsvc.GetModuleDependenciesResponse,
  /**
   * @param {!proto.autokitteh.langsvc.GetModuleDependenciesRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.langsvc.GetModuleDependenciesResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.langsvc.GetModuleDependenciesRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.langsvc.GetModuleDependenciesResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.langsvc.GetModuleDependenciesResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.langsvc.LangClient.prototype.getModuleDependencies =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.langsvc.Lang/GetModuleDependencies',
      request,
      metadata || {},
      methodDescriptor_Lang_GetModuleDependencies,
      callback);
};


/**
 * @param {!proto.autokitteh.langsvc.GetModuleDependenciesRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.langsvc.GetModuleDependenciesResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.langsvc.LangPromiseClient.prototype.getModuleDependencies =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.langsvc.Lang/GetModuleDependencies',
      request,
      metadata || {},
      methodDescriptor_Lang_GetModuleDependencies);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.langsvc.CompileModuleRequest,
 *   !proto.autokitteh.langsvc.CompileModuleResponse>}
 */
const methodDescriptor_Lang_CompileModule = new grpc.web.MethodDescriptor(
  '/autokitteh.langsvc.Lang/CompileModule',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.langsvc.CompileModuleRequest,
  proto.autokitteh.langsvc.CompileModuleResponse,
  /**
   * @param {!proto.autokitteh.langsvc.CompileModuleRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.langsvc.CompileModuleResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.langsvc.CompileModuleRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.langsvc.CompileModuleResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.langsvc.CompileModuleResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.langsvc.LangClient.prototype.compileModule =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.langsvc.Lang/CompileModule',
      request,
      metadata || {},
      methodDescriptor_Lang_CompileModule,
      callback);
};


/**
 * @param {!proto.autokitteh.langsvc.CompileModuleRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.langsvc.CompileModuleResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.langsvc.LangPromiseClient.prototype.compileModule =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.langsvc.Lang/CompileModule',
      request,
      metadata || {},
      methodDescriptor_Lang_CompileModule);
};


module.exports = proto.autokitteh.langsvc;

