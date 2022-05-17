/**
 * @fileoverview gRPC-Web generated client stub for autokitteh.httpeventsrc
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
proto.autokitteh.httpeventsrc = require('./src_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.httpeventsrc.HTTPEventSourceClient =
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
proto.autokitteh.httpeventsrc.HTTPEventSourcePromiseClient =
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
 *   !proto.autokitteh.httpeventsrc.BindRequest,
 *   !proto.autokitteh.httpeventsrc.BindResponse>}
 */
const methodDescriptor_HTTPEventSource_Bind = new grpc.web.MethodDescriptor(
  '/autokitteh.httpeventsrc.HTTPEventSource/Bind',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.httpeventsrc.BindRequest,
  proto.autokitteh.httpeventsrc.BindResponse,
  /**
   * @param {!proto.autokitteh.httpeventsrc.BindRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.httpeventsrc.BindResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.httpeventsrc.BindRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.httpeventsrc.BindResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.httpeventsrc.BindResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.httpeventsrc.HTTPEventSourceClient.prototype.bind =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.httpeventsrc.HTTPEventSource/Bind',
      request,
      metadata || {},
      methodDescriptor_HTTPEventSource_Bind,
      callback);
};


/**
 * @param {!proto.autokitteh.httpeventsrc.BindRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.httpeventsrc.BindResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.httpeventsrc.HTTPEventSourcePromiseClient.prototype.bind =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.httpeventsrc.HTTPEventSource/Bind',
      request,
      metadata || {},
      methodDescriptor_HTTPEventSource_Bind);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.httpeventsrc.UnbindRequest,
 *   !proto.autokitteh.httpeventsrc.UnbindResponse>}
 */
const methodDescriptor_HTTPEventSource_Unbind = new grpc.web.MethodDescriptor(
  '/autokitteh.httpeventsrc.HTTPEventSource/Unbind',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.httpeventsrc.UnbindRequest,
  proto.autokitteh.httpeventsrc.UnbindResponse,
  /**
   * @param {!proto.autokitteh.httpeventsrc.UnbindRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.httpeventsrc.UnbindResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.httpeventsrc.UnbindRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.httpeventsrc.UnbindResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.httpeventsrc.UnbindResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.httpeventsrc.HTTPEventSourceClient.prototype.unbind =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.httpeventsrc.HTTPEventSource/Unbind',
      request,
      metadata || {},
      methodDescriptor_HTTPEventSource_Unbind,
      callback);
};


/**
 * @param {!proto.autokitteh.httpeventsrc.UnbindRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.httpeventsrc.UnbindResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.httpeventsrc.HTTPEventSourcePromiseClient.prototype.unbind =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.httpeventsrc.HTTPEventSource/Unbind',
      request,
      metadata || {},
      methodDescriptor_HTTPEventSource_Unbind);
};


module.exports = proto.autokitteh.httpeventsrc;

