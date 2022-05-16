/**
 * @fileoverview gRPC-Web generated client stub for autokitteh.pluginsregsvc
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


var plugin_plugin_pb = require('../plugin/plugin_pb.js')

var validate_validate_pb = require('../validate/validate_pb.js')
const proto = {};
proto.autokitteh = {};
proto.autokitteh.pluginsregsvc = require('./svc_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.pluginsregsvc.PluginsRegistryClient =
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
proto.autokitteh.pluginsregsvc.PluginsRegistryPromiseClient =
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
 *   !proto.autokitteh.pluginsregsvc.ListRequest,
 *   !proto.autokitteh.pluginsregsvc.ListResponse>}
 */
const methodDescriptor_PluginsRegistry_List = new grpc.web.MethodDescriptor(
  '/autokitteh.pluginsregsvc.PluginsRegistry/List',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.pluginsregsvc.ListRequest,
  proto.autokitteh.pluginsregsvc.ListResponse,
  /**
   * @param {!proto.autokitteh.pluginsregsvc.ListRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.pluginsregsvc.ListResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.pluginsregsvc.ListRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.pluginsregsvc.ListResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.pluginsregsvc.ListResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.pluginsregsvc.PluginsRegistryClient.prototype.list =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.pluginsregsvc.PluginsRegistry/List',
      request,
      metadata || {},
      methodDescriptor_PluginsRegistry_List,
      callback);
};


/**
 * @param {!proto.autokitteh.pluginsregsvc.ListRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.pluginsregsvc.ListResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.pluginsregsvc.PluginsRegistryPromiseClient.prototype.list =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.pluginsregsvc.PluginsRegistry/List',
      request,
      metadata || {},
      methodDescriptor_PluginsRegistry_List);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.pluginsregsvc.GetRequest,
 *   !proto.autokitteh.pluginsregsvc.GetResponse>}
 */
const methodDescriptor_PluginsRegistry_Get = new grpc.web.MethodDescriptor(
  '/autokitteh.pluginsregsvc.PluginsRegistry/Get',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.pluginsregsvc.GetRequest,
  proto.autokitteh.pluginsregsvc.GetResponse,
  /**
   * @param {!proto.autokitteh.pluginsregsvc.GetRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.pluginsregsvc.GetResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.pluginsregsvc.GetRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.pluginsregsvc.GetResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.pluginsregsvc.GetResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.pluginsregsvc.PluginsRegistryClient.prototype.get =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.pluginsregsvc.PluginsRegistry/Get',
      request,
      metadata || {},
      methodDescriptor_PluginsRegistry_Get,
      callback);
};


/**
 * @param {!proto.autokitteh.pluginsregsvc.GetRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.pluginsregsvc.GetResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.pluginsregsvc.PluginsRegistryPromiseClient.prototype.get =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.pluginsregsvc.PluginsRegistry/Get',
      request,
      metadata || {},
      methodDescriptor_PluginsRegistry_Get);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.pluginsregsvc.RegisterRequest,
 *   !proto.autokitteh.pluginsregsvc.RegisterResponse>}
 */
const methodDescriptor_PluginsRegistry_Register = new grpc.web.MethodDescriptor(
  '/autokitteh.pluginsregsvc.PluginsRegistry/Register',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.pluginsregsvc.RegisterRequest,
  proto.autokitteh.pluginsregsvc.RegisterResponse,
  /**
   * @param {!proto.autokitteh.pluginsregsvc.RegisterRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.pluginsregsvc.RegisterResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.pluginsregsvc.RegisterRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.pluginsregsvc.RegisterResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.pluginsregsvc.RegisterResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.pluginsregsvc.PluginsRegistryClient.prototype.register =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.pluginsregsvc.PluginsRegistry/Register',
      request,
      metadata || {},
      methodDescriptor_PluginsRegistry_Register,
      callback);
};


/**
 * @param {!proto.autokitteh.pluginsregsvc.RegisterRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.pluginsregsvc.RegisterResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.pluginsregsvc.PluginsRegistryPromiseClient.prototype.register =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.pluginsregsvc.PluginsRegistry/Register',
      request,
      metadata || {},
      methodDescriptor_PluginsRegistry_Register);
};


module.exports = proto.autokitteh.pluginsregsvc;

