/**
 * @fileoverview gRPC-Web generated client stub for autokitteh.fseventsrc
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
proto.autokitteh.fseventsrc = require('./src_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.fseventsrc.FSEventSourceClient =
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
proto.autokitteh.fseventsrc.FSEventSourcePromiseClient =
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
 *   !proto.autokitteh.fseventsrc.BindRequest,
 *   !proto.autokitteh.fseventsrc.BindResponse>}
 */
const methodDescriptor_FSEventSource_Bind = new grpc.web.MethodDescriptor(
  '/autokitteh.fseventsrc.FSEventSource/Bind',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.fseventsrc.BindRequest,
  proto.autokitteh.fseventsrc.BindResponse,
  /**
   * @param {!proto.autokitteh.fseventsrc.BindRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.fseventsrc.BindResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.fseventsrc.BindRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.fseventsrc.BindResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.fseventsrc.BindResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.fseventsrc.FSEventSourceClient.prototype.bind =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.fseventsrc.FSEventSource/Bind',
      request,
      metadata || {},
      methodDescriptor_FSEventSource_Bind,
      callback);
};


/**
 * @param {!proto.autokitteh.fseventsrc.BindRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.fseventsrc.BindResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.fseventsrc.FSEventSourcePromiseClient.prototype.bind =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.fseventsrc.FSEventSource/Bind',
      request,
      metadata || {},
      methodDescriptor_FSEventSource_Bind);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.fseventsrc.UnbindRequest,
 *   !proto.autokitteh.fseventsrc.UnbindResponse>}
 */
const methodDescriptor_FSEventSource_Unbind = new grpc.web.MethodDescriptor(
  '/autokitteh.fseventsrc.FSEventSource/Unbind',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.fseventsrc.UnbindRequest,
  proto.autokitteh.fseventsrc.UnbindResponse,
  /**
   * @param {!proto.autokitteh.fseventsrc.UnbindRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.fseventsrc.UnbindResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.fseventsrc.UnbindRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.fseventsrc.UnbindResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.fseventsrc.UnbindResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.fseventsrc.FSEventSourceClient.prototype.unbind =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.fseventsrc.FSEventSource/Unbind',
      request,
      metadata || {},
      methodDescriptor_FSEventSource_Unbind,
      callback);
};


/**
 * @param {!proto.autokitteh.fseventsrc.UnbindRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.fseventsrc.UnbindResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.fseventsrc.FSEventSourcePromiseClient.prototype.unbind =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.fseventsrc.FSEventSource/Unbind',
      request,
      metadata || {},
      methodDescriptor_FSEventSource_Unbind);
};


module.exports = proto.autokitteh.fseventsrc;

