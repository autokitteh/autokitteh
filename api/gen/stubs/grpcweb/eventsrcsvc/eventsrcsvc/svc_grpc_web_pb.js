/**
 * @fileoverview gRPC-Web generated client stub for autokitteh.eventsrcsvc
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


var validate_validate_pb = require('../validate/validate_pb.js')

var eventsrc_src_pb = require('../eventsrc/src_pb.js')
const proto = {};
proto.autokitteh = {};
proto.autokitteh.eventsrcsvc = require('./svc_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.eventsrcsvc.EventSourcesClient =
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
proto.autokitteh.eventsrcsvc.EventSourcesPromiseClient =
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
 *   !proto.autokitteh.eventsrcsvc.AddEventSourceRequest,
 *   !proto.autokitteh.eventsrcsvc.AddEventSourceResponse>}
 */
const methodDescriptor_EventSources_AddEventSource = new grpc.web.MethodDescriptor(
  '/autokitteh.eventsrcsvc.EventSources/AddEventSource',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.eventsrcsvc.AddEventSourceRequest,
  proto.autokitteh.eventsrcsvc.AddEventSourceResponse,
  /**
   * @param {!proto.autokitteh.eventsrcsvc.AddEventSourceRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.eventsrcsvc.AddEventSourceResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.eventsrcsvc.AddEventSourceRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.eventsrcsvc.AddEventSourceResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.eventsrcsvc.AddEventSourceResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.eventsrcsvc.EventSourcesClient.prototype.addEventSource =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.eventsrcsvc.EventSources/AddEventSource',
      request,
      metadata || {},
      methodDescriptor_EventSources_AddEventSource,
      callback);
};


/**
 * @param {!proto.autokitteh.eventsrcsvc.AddEventSourceRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.eventsrcsvc.AddEventSourceResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.eventsrcsvc.EventSourcesPromiseClient.prototype.addEventSource =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.eventsrcsvc.EventSources/AddEventSource',
      request,
      metadata || {},
      methodDescriptor_EventSources_AddEventSource);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.eventsrcsvc.UpdateEventSourceRequest,
 *   !proto.autokitteh.eventsrcsvc.UpdateEventSourceResponse>}
 */
const methodDescriptor_EventSources_UpdateEventSource = new grpc.web.MethodDescriptor(
  '/autokitteh.eventsrcsvc.EventSources/UpdateEventSource',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.eventsrcsvc.UpdateEventSourceRequest,
  proto.autokitteh.eventsrcsvc.UpdateEventSourceResponse,
  /**
   * @param {!proto.autokitteh.eventsrcsvc.UpdateEventSourceRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.eventsrcsvc.UpdateEventSourceResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.eventsrcsvc.UpdateEventSourceRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.eventsrcsvc.UpdateEventSourceResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.eventsrcsvc.UpdateEventSourceResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.eventsrcsvc.EventSourcesClient.prototype.updateEventSource =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.eventsrcsvc.EventSources/UpdateEventSource',
      request,
      metadata || {},
      methodDescriptor_EventSources_UpdateEventSource,
      callback);
};


/**
 * @param {!proto.autokitteh.eventsrcsvc.UpdateEventSourceRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.eventsrcsvc.UpdateEventSourceResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.eventsrcsvc.EventSourcesPromiseClient.prototype.updateEventSource =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.eventsrcsvc.EventSources/UpdateEventSource',
      request,
      metadata || {},
      methodDescriptor_EventSources_UpdateEventSource);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.eventsrcsvc.GetEventSourceRequest,
 *   !proto.autokitteh.eventsrcsvc.GetEventSourceResponse>}
 */
const methodDescriptor_EventSources_GetEventSource = new grpc.web.MethodDescriptor(
  '/autokitteh.eventsrcsvc.EventSources/GetEventSource',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.eventsrcsvc.GetEventSourceRequest,
  proto.autokitteh.eventsrcsvc.GetEventSourceResponse,
  /**
   * @param {!proto.autokitteh.eventsrcsvc.GetEventSourceRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.eventsrcsvc.GetEventSourceResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.eventsrcsvc.GetEventSourceRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.eventsrcsvc.GetEventSourceResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.eventsrcsvc.GetEventSourceResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.eventsrcsvc.EventSourcesClient.prototype.getEventSource =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.eventsrcsvc.EventSources/GetEventSource',
      request,
      metadata || {},
      methodDescriptor_EventSources_GetEventSource,
      callback);
};


/**
 * @param {!proto.autokitteh.eventsrcsvc.GetEventSourceRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.eventsrcsvc.GetEventSourceResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.eventsrcsvc.EventSourcesPromiseClient.prototype.getEventSource =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.eventsrcsvc.EventSources/GetEventSource',
      request,
      metadata || {},
      methodDescriptor_EventSources_GetEventSource);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.eventsrcsvc.AddEventSourceProjectBindingRequest,
 *   !proto.autokitteh.eventsrcsvc.AddEventSourceProjectBindingResponse>}
 */
const methodDescriptor_EventSources_AddEventSourceProjectBinding = new grpc.web.MethodDescriptor(
  '/autokitteh.eventsrcsvc.EventSources/AddEventSourceProjectBinding',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.eventsrcsvc.AddEventSourceProjectBindingRequest,
  proto.autokitteh.eventsrcsvc.AddEventSourceProjectBindingResponse,
  /**
   * @param {!proto.autokitteh.eventsrcsvc.AddEventSourceProjectBindingRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.eventsrcsvc.AddEventSourceProjectBindingResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.eventsrcsvc.AddEventSourceProjectBindingRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.eventsrcsvc.AddEventSourceProjectBindingResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.eventsrcsvc.AddEventSourceProjectBindingResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.eventsrcsvc.EventSourcesClient.prototype.addEventSourceProjectBinding =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.eventsrcsvc.EventSources/AddEventSourceProjectBinding',
      request,
      metadata || {},
      methodDescriptor_EventSources_AddEventSourceProjectBinding,
      callback);
};


/**
 * @param {!proto.autokitteh.eventsrcsvc.AddEventSourceProjectBindingRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.eventsrcsvc.AddEventSourceProjectBindingResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.eventsrcsvc.EventSourcesPromiseClient.prototype.addEventSourceProjectBinding =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.eventsrcsvc.EventSources/AddEventSourceProjectBinding',
      request,
      metadata || {},
      methodDescriptor_EventSources_AddEventSourceProjectBinding);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.eventsrcsvc.UpdateEventSourceProjectBindingRequest,
 *   !proto.autokitteh.eventsrcsvc.UpdateEventSourceProjectBindingResponse>}
 */
const methodDescriptor_EventSources_UpdateEventSourceProjectBinding = new grpc.web.MethodDescriptor(
  '/autokitteh.eventsrcsvc.EventSources/UpdateEventSourceProjectBinding',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.eventsrcsvc.UpdateEventSourceProjectBindingRequest,
  proto.autokitteh.eventsrcsvc.UpdateEventSourceProjectBindingResponse,
  /**
   * @param {!proto.autokitteh.eventsrcsvc.UpdateEventSourceProjectBindingRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.eventsrcsvc.UpdateEventSourceProjectBindingResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.eventsrcsvc.UpdateEventSourceProjectBindingRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.eventsrcsvc.UpdateEventSourceProjectBindingResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.eventsrcsvc.UpdateEventSourceProjectBindingResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.eventsrcsvc.EventSourcesClient.prototype.updateEventSourceProjectBinding =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.eventsrcsvc.EventSources/UpdateEventSourceProjectBinding',
      request,
      metadata || {},
      methodDescriptor_EventSources_UpdateEventSourceProjectBinding,
      callback);
};


/**
 * @param {!proto.autokitteh.eventsrcsvc.UpdateEventSourceProjectBindingRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.eventsrcsvc.UpdateEventSourceProjectBindingResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.eventsrcsvc.EventSourcesPromiseClient.prototype.updateEventSourceProjectBinding =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.eventsrcsvc.EventSources/UpdateEventSourceProjectBinding',
      request,
      metadata || {},
      methodDescriptor_EventSources_UpdateEventSourceProjectBinding);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.eventsrcsvc.GetEventSourceProjectBindingsRequest,
 *   !proto.autokitteh.eventsrcsvc.GetEventSourceProjectBindingsResponse>}
 */
const methodDescriptor_EventSources_GetEventSourceProjectBindings = new grpc.web.MethodDescriptor(
  '/autokitteh.eventsrcsvc.EventSources/GetEventSourceProjectBindings',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.eventsrcsvc.GetEventSourceProjectBindingsRequest,
  proto.autokitteh.eventsrcsvc.GetEventSourceProjectBindingsResponse,
  /**
   * @param {!proto.autokitteh.eventsrcsvc.GetEventSourceProjectBindingsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.eventsrcsvc.GetEventSourceProjectBindingsResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.eventsrcsvc.GetEventSourceProjectBindingsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.eventsrcsvc.GetEventSourceProjectBindingsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.eventsrcsvc.GetEventSourceProjectBindingsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.eventsrcsvc.EventSourcesClient.prototype.getEventSourceProjectBindings =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.eventsrcsvc.EventSources/GetEventSourceProjectBindings',
      request,
      metadata || {},
      methodDescriptor_EventSources_GetEventSourceProjectBindings,
      callback);
};


/**
 * @param {!proto.autokitteh.eventsrcsvc.GetEventSourceProjectBindingsRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.eventsrcsvc.GetEventSourceProjectBindingsResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.eventsrcsvc.EventSourcesPromiseClient.prototype.getEventSourceProjectBindings =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.eventsrcsvc.EventSources/GetEventSourceProjectBindings',
      request,
      metadata || {},
      methodDescriptor_EventSources_GetEventSourceProjectBindings);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.eventsrcsvc.ListEventSourcesRequest,
 *   !proto.autokitteh.eventsrcsvc.ListEventSourcesResponse>}
 */
const methodDescriptor_EventSources_ListEventSources = new grpc.web.MethodDescriptor(
  '/autokitteh.eventsrcsvc.EventSources/ListEventSources',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.eventsrcsvc.ListEventSourcesRequest,
  proto.autokitteh.eventsrcsvc.ListEventSourcesResponse,
  /**
   * @param {!proto.autokitteh.eventsrcsvc.ListEventSourcesRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.eventsrcsvc.ListEventSourcesResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.eventsrcsvc.ListEventSourcesRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.eventsrcsvc.ListEventSourcesResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.eventsrcsvc.ListEventSourcesResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.eventsrcsvc.EventSourcesClient.prototype.listEventSources =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.eventsrcsvc.EventSources/ListEventSources',
      request,
      metadata || {},
      methodDescriptor_EventSources_ListEventSources,
      callback);
};


/**
 * @param {!proto.autokitteh.eventsrcsvc.ListEventSourcesRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.eventsrcsvc.ListEventSourcesResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.eventsrcsvc.EventSourcesPromiseClient.prototype.listEventSources =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.eventsrcsvc.EventSources/ListEventSources',
      request,
      metadata || {},
      methodDescriptor_EventSources_ListEventSources);
};


module.exports = proto.autokitteh.eventsrcsvc;

