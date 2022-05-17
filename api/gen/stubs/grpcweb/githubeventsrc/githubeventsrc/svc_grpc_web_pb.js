/**
 * @fileoverview gRPC-Web generated client stub for autokitteh.githubeventsrc
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
proto.autokitteh.githubeventsrc = require('./svc_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.githubeventsrc.GithubEventSourceClient =
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
proto.autokitteh.githubeventsrc.GithubEventSourcePromiseClient =
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
 *   !proto.autokitteh.githubeventsrc.BindRequest,
 *   !proto.autokitteh.githubeventsrc.BindResponse>}
 */
const methodDescriptor_GithubEventSource_Bind = new grpc.web.MethodDescriptor(
  '/autokitteh.githubeventsrc.GithubEventSource/Bind',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.githubeventsrc.BindRequest,
  proto.autokitteh.githubeventsrc.BindResponse,
  /**
   * @param {!proto.autokitteh.githubeventsrc.BindRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.githubeventsrc.BindResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.githubeventsrc.BindRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.githubeventsrc.BindResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.githubeventsrc.BindResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.githubeventsrc.GithubEventSourceClient.prototype.bind =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.githubeventsrc.GithubEventSource/Bind',
      request,
      metadata || {},
      methodDescriptor_GithubEventSource_Bind,
      callback);
};


/**
 * @param {!proto.autokitteh.githubeventsrc.BindRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.githubeventsrc.BindResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.githubeventsrc.GithubEventSourcePromiseClient.prototype.bind =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.githubeventsrc.GithubEventSource/Bind',
      request,
      metadata || {},
      methodDescriptor_GithubEventSource_Bind);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.githubeventsrc.UnbindRequest,
 *   !proto.autokitteh.githubeventsrc.UnbindResponse>}
 */
const methodDescriptor_GithubEventSource_Unbind = new grpc.web.MethodDescriptor(
  '/autokitteh.githubeventsrc.GithubEventSource/Unbind',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.githubeventsrc.UnbindRequest,
  proto.autokitteh.githubeventsrc.UnbindResponse,
  /**
   * @param {!proto.autokitteh.githubeventsrc.UnbindRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.githubeventsrc.UnbindResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.githubeventsrc.UnbindRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.githubeventsrc.UnbindResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.githubeventsrc.UnbindResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.githubeventsrc.GithubEventSourceClient.prototype.unbind =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.githubeventsrc.GithubEventSource/Unbind',
      request,
      metadata || {},
      methodDescriptor_GithubEventSource_Unbind,
      callback);
};


/**
 * @param {!proto.autokitteh.githubeventsrc.UnbindRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.githubeventsrc.UnbindResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.githubeventsrc.GithubEventSourcePromiseClient.prototype.unbind =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.githubeventsrc.GithubEventSource/Unbind',
      request,
      metadata || {},
      methodDescriptor_GithubEventSource_Unbind);
};


module.exports = proto.autokitteh.githubeventsrc;

