/**
 * @fileoverview gRPC-Web generated client stub for autokitteh.slackeventsrc
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
proto.autokitteh.slackeventsrc = require('./src_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.slackeventsrc.SlackEventSourceClient =
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
proto.autokitteh.slackeventsrc.SlackEventSourcePromiseClient =
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
 *   !proto.autokitteh.slackeventsrc.BindRequest,
 *   !proto.autokitteh.slackeventsrc.BindResponse>}
 */
const methodDescriptor_SlackEventSource_Bind = new grpc.web.MethodDescriptor(
  '/autokitteh.slackeventsrc.SlackEventSource/Bind',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.slackeventsrc.BindRequest,
  proto.autokitteh.slackeventsrc.BindResponse,
  /**
   * @param {!proto.autokitteh.slackeventsrc.BindRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.slackeventsrc.BindResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.slackeventsrc.BindRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.slackeventsrc.BindResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.slackeventsrc.BindResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.slackeventsrc.SlackEventSourceClient.prototype.bind =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.slackeventsrc.SlackEventSource/Bind',
      request,
      metadata || {},
      methodDescriptor_SlackEventSource_Bind,
      callback);
};


/**
 * @param {!proto.autokitteh.slackeventsrc.BindRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.slackeventsrc.BindResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.slackeventsrc.SlackEventSourcePromiseClient.prototype.bind =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.slackeventsrc.SlackEventSource/Bind',
      request,
      metadata || {},
      methodDescriptor_SlackEventSource_Bind);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.slackeventsrc.UnbindRequest,
 *   !proto.autokitteh.slackeventsrc.UnbindResponse>}
 */
const methodDescriptor_SlackEventSource_Unbind = new grpc.web.MethodDescriptor(
  '/autokitteh.slackeventsrc.SlackEventSource/Unbind',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.slackeventsrc.UnbindRequest,
  proto.autokitteh.slackeventsrc.UnbindResponse,
  /**
   * @param {!proto.autokitteh.slackeventsrc.UnbindRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.slackeventsrc.UnbindResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.slackeventsrc.UnbindRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.slackeventsrc.UnbindResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.slackeventsrc.UnbindResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.slackeventsrc.SlackEventSourceClient.prototype.unbind =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.slackeventsrc.SlackEventSource/Unbind',
      request,
      metadata || {},
      methodDescriptor_SlackEventSource_Unbind,
      callback);
};


/**
 * @param {!proto.autokitteh.slackeventsrc.UnbindRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.slackeventsrc.UnbindResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.slackeventsrc.SlackEventSourcePromiseClient.prototype.unbind =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.slackeventsrc.SlackEventSource/Unbind',
      request,
      metadata || {},
      methodDescriptor_SlackEventSource_Unbind);
};


module.exports = proto.autokitteh.slackeventsrc;

