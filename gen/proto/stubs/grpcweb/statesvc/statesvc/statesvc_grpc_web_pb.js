/**
 * @fileoverview gRPC-Web generated client stub for autokitteh.statesvc
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

var values_values_pb = require('../values/values_pb.js')
const proto = {};
proto.autokitteh = {};
proto.autokitteh.statesvc = require('./statesvc_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.statesvc.StateClient =
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
proto.autokitteh.statesvc.StatePromiseClient =
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
 *   !proto.autokitteh.statesvc.SetRequest,
 *   !proto.autokitteh.statesvc.SetResponse>}
 */
const methodDescriptor_State_Set = new grpc.web.MethodDescriptor(
  '/autokitteh.statesvc.State/Set',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.statesvc.SetRequest,
  proto.autokitteh.statesvc.SetResponse,
  /**
   * @param {!proto.autokitteh.statesvc.SetRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.statesvc.SetResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.statesvc.SetRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.statesvc.SetResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.statesvc.SetResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.statesvc.StateClient.prototype.set =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.statesvc.State/Set',
      request,
      metadata || {},
      methodDescriptor_State_Set,
      callback);
};


/**
 * @param {!proto.autokitteh.statesvc.SetRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.statesvc.SetResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.statesvc.StatePromiseClient.prototype.set =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.statesvc.State/Set',
      request,
      metadata || {},
      methodDescriptor_State_Set);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.statesvc.GetRequest,
 *   !proto.autokitteh.statesvc.GetResponse>}
 */
const methodDescriptor_State_Get = new grpc.web.MethodDescriptor(
  '/autokitteh.statesvc.State/Get',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.statesvc.GetRequest,
  proto.autokitteh.statesvc.GetResponse,
  /**
   * @param {!proto.autokitteh.statesvc.GetRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.statesvc.GetResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.statesvc.GetRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.statesvc.GetResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.statesvc.GetResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.statesvc.StateClient.prototype.get =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.statesvc.State/Get',
      request,
      metadata || {},
      methodDescriptor_State_Get,
      callback);
};


/**
 * @param {!proto.autokitteh.statesvc.GetRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.statesvc.GetResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.statesvc.StatePromiseClient.prototype.get =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.statesvc.State/Get',
      request,
      metadata || {},
      methodDescriptor_State_Get);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.statesvc.ListRequest,
 *   !proto.autokitteh.statesvc.ListResponse>}
 */
const methodDescriptor_State_List = new grpc.web.MethodDescriptor(
  '/autokitteh.statesvc.State/List',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.statesvc.ListRequest,
  proto.autokitteh.statesvc.ListResponse,
  /**
   * @param {!proto.autokitteh.statesvc.ListRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.statesvc.ListResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.statesvc.ListRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.statesvc.ListResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.statesvc.ListResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.statesvc.StateClient.prototype.list =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.statesvc.State/List',
      request,
      metadata || {},
      methodDescriptor_State_List,
      callback);
};


/**
 * @param {!proto.autokitteh.statesvc.ListRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.statesvc.ListResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.statesvc.StatePromiseClient.prototype.list =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.statesvc.State/List',
      request,
      metadata || {},
      methodDescriptor_State_List);
};


module.exports = proto.autokitteh.statesvc;

