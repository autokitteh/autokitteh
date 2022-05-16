/**
 * @fileoverview gRPC-Web generated client stub for autokitteh.secretssvc
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
const proto = {};
proto.autokitteh = {};
proto.autokitteh.secretssvc = require('./svc_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.secretssvc.SecretsClient =
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
proto.autokitteh.secretssvc.SecretsPromiseClient =
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
 *   !proto.autokitteh.secretssvc.SetRequest,
 *   !proto.autokitteh.secretssvc.SetResponse>}
 */
const methodDescriptor_Secrets_Set = new grpc.web.MethodDescriptor(
  '/autokitteh.secretssvc.Secrets/Set',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.secretssvc.SetRequest,
  proto.autokitteh.secretssvc.SetResponse,
  /**
   * @param {!proto.autokitteh.secretssvc.SetRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.secretssvc.SetResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.secretssvc.SetRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.secretssvc.SetResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.secretssvc.SetResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.secretssvc.SecretsClient.prototype.set =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.secretssvc.Secrets/Set',
      request,
      metadata || {},
      methodDescriptor_Secrets_Set,
      callback);
};


/**
 * @param {!proto.autokitteh.secretssvc.SetRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.secretssvc.SetResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.secretssvc.SecretsPromiseClient.prototype.set =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.secretssvc.Secrets/Set',
      request,
      metadata || {},
      methodDescriptor_Secrets_Set);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.secretssvc.GetRequest,
 *   !proto.autokitteh.secretssvc.GetResponse>}
 */
const methodDescriptor_Secrets_Get = new grpc.web.MethodDescriptor(
  '/autokitteh.secretssvc.Secrets/Get',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.secretssvc.GetRequest,
  proto.autokitteh.secretssvc.GetResponse,
  /**
   * @param {!proto.autokitteh.secretssvc.GetRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.secretssvc.GetResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.secretssvc.GetRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.secretssvc.GetResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.secretssvc.GetResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.secretssvc.SecretsClient.prototype.get =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.secretssvc.Secrets/Get',
      request,
      metadata || {},
      methodDescriptor_Secrets_Get,
      callback);
};


/**
 * @param {!proto.autokitteh.secretssvc.GetRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.secretssvc.GetResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.secretssvc.SecretsPromiseClient.prototype.get =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.secretssvc.Secrets/Get',
      request,
      metadata || {},
      methodDescriptor_Secrets_Get);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.secretssvc.ListRequest,
 *   !proto.autokitteh.secretssvc.ListResponse>}
 */
const methodDescriptor_Secrets_List = new grpc.web.MethodDescriptor(
  '/autokitteh.secretssvc.Secrets/List',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.secretssvc.ListRequest,
  proto.autokitteh.secretssvc.ListResponse,
  /**
   * @param {!proto.autokitteh.secretssvc.ListRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.secretssvc.ListResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.secretssvc.ListRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.secretssvc.ListResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.secretssvc.ListResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.secretssvc.SecretsClient.prototype.list =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.secretssvc.Secrets/List',
      request,
      metadata || {},
      methodDescriptor_Secrets_List,
      callback);
};


/**
 * @param {!proto.autokitteh.secretssvc.ListRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.secretssvc.ListResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.secretssvc.SecretsPromiseClient.prototype.list =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.secretssvc.Secrets/List',
      request,
      metadata || {},
      methodDescriptor_Secrets_List);
};


module.exports = proto.autokitteh.secretssvc;

