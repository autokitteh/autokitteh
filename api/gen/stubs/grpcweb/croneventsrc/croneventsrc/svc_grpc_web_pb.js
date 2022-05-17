/**
 * @fileoverview gRPC-Web generated client stub for autokitteh.croneventsrc
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
proto.autokitteh.croneventsrc = require('./svc_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.croneventsrc.CronEventSourceClient =
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
proto.autokitteh.croneventsrc.CronEventSourcePromiseClient =
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
 *   !proto.autokitteh.croneventsrc.TickRequest,
 *   !proto.autokitteh.croneventsrc.TickResponse>}
 */
const methodDescriptor_CronEventSource_Tick = new grpc.web.MethodDescriptor(
  '/autokitteh.croneventsrc.CronEventSource/Tick',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.croneventsrc.TickRequest,
  proto.autokitteh.croneventsrc.TickResponse,
  /**
   * @param {!proto.autokitteh.croneventsrc.TickRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.croneventsrc.TickResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.croneventsrc.TickRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.croneventsrc.TickResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.croneventsrc.TickResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.croneventsrc.CronEventSourceClient.prototype.tick =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.croneventsrc.CronEventSource/Tick',
      request,
      metadata || {},
      methodDescriptor_CronEventSource_Tick,
      callback);
};


/**
 * @param {!proto.autokitteh.croneventsrc.TickRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.croneventsrc.TickResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.croneventsrc.CronEventSourcePromiseClient.prototype.tick =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.croneventsrc.CronEventSource/Tick',
      request,
      metadata || {},
      methodDescriptor_CronEventSource_Tick);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.croneventsrc.BindRequest,
 *   !proto.autokitteh.croneventsrc.BindResponse>}
 */
const methodDescriptor_CronEventSource_Bind = new grpc.web.MethodDescriptor(
  '/autokitteh.croneventsrc.CronEventSource/Bind',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.croneventsrc.BindRequest,
  proto.autokitteh.croneventsrc.BindResponse,
  /**
   * @param {!proto.autokitteh.croneventsrc.BindRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.croneventsrc.BindResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.croneventsrc.BindRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.croneventsrc.BindResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.croneventsrc.BindResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.croneventsrc.CronEventSourceClient.prototype.bind =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.croneventsrc.CronEventSource/Bind',
      request,
      metadata || {},
      methodDescriptor_CronEventSource_Bind,
      callback);
};


/**
 * @param {!proto.autokitteh.croneventsrc.BindRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.croneventsrc.BindResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.croneventsrc.CronEventSourcePromiseClient.prototype.bind =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.croneventsrc.CronEventSource/Bind',
      request,
      metadata || {},
      methodDescriptor_CronEventSource_Bind);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.croneventsrc.UnbindRequest,
 *   !proto.autokitteh.croneventsrc.UnbindResponse>}
 */
const methodDescriptor_CronEventSource_Unbind = new grpc.web.MethodDescriptor(
  '/autokitteh.croneventsrc.CronEventSource/Unbind',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.croneventsrc.UnbindRequest,
  proto.autokitteh.croneventsrc.UnbindResponse,
  /**
   * @param {!proto.autokitteh.croneventsrc.UnbindRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.croneventsrc.UnbindResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.croneventsrc.UnbindRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.croneventsrc.UnbindResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.croneventsrc.UnbindResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.croneventsrc.CronEventSourceClient.prototype.unbind =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.croneventsrc.CronEventSource/Unbind',
      request,
      metadata || {},
      methodDescriptor_CronEventSource_Unbind,
      callback);
};


/**
 * @param {!proto.autokitteh.croneventsrc.UnbindRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.croneventsrc.UnbindResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.croneventsrc.CronEventSourcePromiseClient.prototype.unbind =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.croneventsrc.CronEventSource/Unbind',
      request,
      metadata || {},
      methodDescriptor_CronEventSource_Unbind);
};


module.exports = proto.autokitteh.croneventsrc;

