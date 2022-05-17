/**
 * @fileoverview gRPC-Web generated client stub for autokitteh.pluginsprovidersvc
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


var validate_validate_pb = require('../validate/validate_pb.js')

var plugin_desc_pb = require('../plugin/desc_pb.js')

var values_values_pb = require('../values/values_pb.js')

var program_error_pb = require('../program/error_pb.js')
const proto = {};
proto.autokitteh = {};
proto.autokitteh.pluginsprovidersvc = require('./svc_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.pluginsprovidersvc.PluginsProviderClient =
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
proto.autokitteh.pluginsprovidersvc.PluginsProviderPromiseClient =
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
 *   !proto.autokitteh.pluginsprovidersvc.ListRequest,
 *   !proto.autokitteh.pluginsprovidersvc.ListResponse>}
 */
const methodDescriptor_PluginsProvider_List = new grpc.web.MethodDescriptor(
  '/autokitteh.pluginsprovidersvc.PluginsProvider/List',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.pluginsprovidersvc.ListRequest,
  proto.autokitteh.pluginsprovidersvc.ListResponse,
  /**
   * @param {!proto.autokitteh.pluginsprovidersvc.ListRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.pluginsprovidersvc.ListResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.pluginsprovidersvc.ListRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.pluginsprovidersvc.ListResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.pluginsprovidersvc.ListResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.pluginsprovidersvc.PluginsProviderClient.prototype.list =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.pluginsprovidersvc.PluginsProvider/List',
      request,
      metadata || {},
      methodDescriptor_PluginsProvider_List,
      callback);
};


/**
 * @param {!proto.autokitteh.pluginsprovidersvc.ListRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.pluginsprovidersvc.ListResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.pluginsprovidersvc.PluginsProviderPromiseClient.prototype.list =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.pluginsprovidersvc.PluginsProvider/List',
      request,
      metadata || {},
      methodDescriptor_PluginsProvider_List);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.pluginsprovidersvc.DescribeRequest,
 *   !proto.autokitteh.pluginsprovidersvc.DescribeResponse>}
 */
const methodDescriptor_PluginsProvider_Describe = new grpc.web.MethodDescriptor(
  '/autokitteh.pluginsprovidersvc.PluginsProvider/Describe',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.pluginsprovidersvc.DescribeRequest,
  proto.autokitteh.pluginsprovidersvc.DescribeResponse,
  /**
   * @param {!proto.autokitteh.pluginsprovidersvc.DescribeRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.pluginsprovidersvc.DescribeResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.pluginsprovidersvc.DescribeRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.pluginsprovidersvc.DescribeResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.pluginsprovidersvc.DescribeResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.pluginsprovidersvc.PluginsProviderClient.prototype.describe =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.pluginsprovidersvc.PluginsProvider/Describe',
      request,
      metadata || {},
      methodDescriptor_PluginsProvider_Describe,
      callback);
};


/**
 * @param {!proto.autokitteh.pluginsprovidersvc.DescribeRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.pluginsprovidersvc.DescribeResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.pluginsprovidersvc.PluginsProviderPromiseClient.prototype.describe =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.pluginsprovidersvc.PluginsProvider/Describe',
      request,
      metadata || {},
      methodDescriptor_PluginsProvider_Describe);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.pluginsprovidersvc.GetValuesRequest,
 *   !proto.autokitteh.pluginsprovidersvc.GetValuesResponse>}
 */
const methodDescriptor_PluginsProvider_GetValues = new grpc.web.MethodDescriptor(
  '/autokitteh.pluginsprovidersvc.PluginsProvider/GetValues',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.pluginsprovidersvc.GetValuesRequest,
  proto.autokitteh.pluginsprovidersvc.GetValuesResponse,
  /**
   * @param {!proto.autokitteh.pluginsprovidersvc.GetValuesRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.pluginsprovidersvc.GetValuesResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.pluginsprovidersvc.GetValuesRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.pluginsprovidersvc.GetValuesResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.pluginsprovidersvc.GetValuesResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.pluginsprovidersvc.PluginsProviderClient.prototype.getValues =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.pluginsprovidersvc.PluginsProvider/GetValues',
      request,
      metadata || {},
      methodDescriptor_PluginsProvider_GetValues,
      callback);
};


/**
 * @param {!proto.autokitteh.pluginsprovidersvc.GetValuesRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.pluginsprovidersvc.GetValuesResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.pluginsprovidersvc.PluginsProviderPromiseClient.prototype.getValues =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.pluginsprovidersvc.PluginsProvider/GetValues',
      request,
      metadata || {},
      methodDescriptor_PluginsProvider_GetValues);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.pluginsprovidersvc.CallValueRequest,
 *   !proto.autokitteh.pluginsprovidersvc.CallValueResponse>}
 */
const methodDescriptor_PluginsProvider_CallValue = new grpc.web.MethodDescriptor(
  '/autokitteh.pluginsprovidersvc.PluginsProvider/CallValue',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.pluginsprovidersvc.CallValueRequest,
  proto.autokitteh.pluginsprovidersvc.CallValueResponse,
  /**
   * @param {!proto.autokitteh.pluginsprovidersvc.CallValueRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.pluginsprovidersvc.CallValueResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.pluginsprovidersvc.CallValueRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.pluginsprovidersvc.CallValueResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.pluginsprovidersvc.CallValueResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.pluginsprovidersvc.PluginsProviderClient.prototype.callValue =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.pluginsprovidersvc.PluginsProvider/CallValue',
      request,
      metadata || {},
      methodDescriptor_PluginsProvider_CallValue,
      callback);
};


/**
 * @param {!proto.autokitteh.pluginsprovidersvc.CallValueRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.pluginsprovidersvc.CallValueResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.pluginsprovidersvc.PluginsProviderPromiseClient.prototype.callValue =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.pluginsprovidersvc.PluginsProvider/CallValue',
      request,
      metadata || {},
      methodDescriptor_PluginsProvider_CallValue);
};


module.exports = proto.autokitteh.pluginsprovidersvc;

