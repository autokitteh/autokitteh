/**
 * @fileoverview gRPC-Web generated client stub for autokitteh.eventsvc
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

var event_event_pb = require('../event/event_pb.js')

var event_event_state_pb = require('../event/event_state_pb.js')

var event_project_state_pb = require('../event/project_state_pb.js')

var values_values_pb = require('../values/values_pb.js')
const proto = {};
proto.autokitteh = {};
proto.autokitteh.eventsvc = require('./svc_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.eventsvc.EventsClient =
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
proto.autokitteh.eventsvc.EventsPromiseClient =
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
 *   !proto.autokitteh.eventsvc.IngestEventRequest,
 *   !proto.autokitteh.eventsvc.IngestEventResponse>}
 */
const methodDescriptor_Events_IngestEvent = new grpc.web.MethodDescriptor(
  '/autokitteh.eventsvc.Events/IngestEvent',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.eventsvc.IngestEventRequest,
  proto.autokitteh.eventsvc.IngestEventResponse,
  /**
   * @param {!proto.autokitteh.eventsvc.IngestEventRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.eventsvc.IngestEventResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.eventsvc.IngestEventRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.eventsvc.IngestEventResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.eventsvc.IngestEventResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.eventsvc.EventsClient.prototype.ingestEvent =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/IngestEvent',
      request,
      metadata || {},
      methodDescriptor_Events_IngestEvent,
      callback);
};


/**
 * @param {!proto.autokitteh.eventsvc.IngestEventRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.eventsvc.IngestEventResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.eventsvc.EventsPromiseClient.prototype.ingestEvent =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/IngestEvent',
      request,
      metadata || {},
      methodDescriptor_Events_IngestEvent);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.eventsvc.GetEventRequest,
 *   !proto.autokitteh.eventsvc.GetEventResponse>}
 */
const methodDescriptor_Events_GetEvent = new grpc.web.MethodDescriptor(
  '/autokitteh.eventsvc.Events/GetEvent',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.eventsvc.GetEventRequest,
  proto.autokitteh.eventsvc.GetEventResponse,
  /**
   * @param {!proto.autokitteh.eventsvc.GetEventRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.eventsvc.GetEventResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.eventsvc.GetEventRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.eventsvc.GetEventResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.eventsvc.GetEventResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.eventsvc.EventsClient.prototype.getEvent =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/GetEvent',
      request,
      metadata || {},
      methodDescriptor_Events_GetEvent,
      callback);
};


/**
 * @param {!proto.autokitteh.eventsvc.GetEventRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.eventsvc.GetEventResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.eventsvc.EventsPromiseClient.prototype.getEvent =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/GetEvent',
      request,
      metadata || {},
      methodDescriptor_Events_GetEvent);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.eventsvc.GetEventStateRequest,
 *   !proto.autokitteh.eventsvc.GetEventStateResponse>}
 */
const methodDescriptor_Events_GetEventState = new grpc.web.MethodDescriptor(
  '/autokitteh.eventsvc.Events/GetEventState',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.eventsvc.GetEventStateRequest,
  proto.autokitteh.eventsvc.GetEventStateResponse,
  /**
   * @param {!proto.autokitteh.eventsvc.GetEventStateRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.eventsvc.GetEventStateResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.eventsvc.GetEventStateRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.eventsvc.GetEventStateResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.eventsvc.GetEventStateResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.eventsvc.EventsClient.prototype.getEventState =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/GetEventState',
      request,
      metadata || {},
      methodDescriptor_Events_GetEventState,
      callback);
};


/**
 * @param {!proto.autokitteh.eventsvc.GetEventStateRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.eventsvc.GetEventStateResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.eventsvc.EventsPromiseClient.prototype.getEventState =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/GetEventState',
      request,
      metadata || {},
      methodDescriptor_Events_GetEventState);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.eventsvc.UpdateEventStateRequest,
 *   !proto.autokitteh.eventsvc.UpdateEventStateResponse>}
 */
const methodDescriptor_Events_UpdateEventState = new grpc.web.MethodDescriptor(
  '/autokitteh.eventsvc.Events/UpdateEventState',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.eventsvc.UpdateEventStateRequest,
  proto.autokitteh.eventsvc.UpdateEventStateResponse,
  /**
   * @param {!proto.autokitteh.eventsvc.UpdateEventStateRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.eventsvc.UpdateEventStateResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.eventsvc.UpdateEventStateRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.eventsvc.UpdateEventStateResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.eventsvc.UpdateEventStateResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.eventsvc.EventsClient.prototype.updateEventState =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/UpdateEventState',
      request,
      metadata || {},
      methodDescriptor_Events_UpdateEventState,
      callback);
};


/**
 * @param {!proto.autokitteh.eventsvc.UpdateEventStateRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.eventsvc.UpdateEventStateResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.eventsvc.EventsPromiseClient.prototype.updateEventState =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/UpdateEventState',
      request,
      metadata || {},
      methodDescriptor_Events_UpdateEventState);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.eventsvc.ListEventsRequest,
 *   !proto.autokitteh.eventsvc.ListEventsResponse>}
 */
const methodDescriptor_Events_ListEvents = new grpc.web.MethodDescriptor(
  '/autokitteh.eventsvc.Events/ListEvents',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.eventsvc.ListEventsRequest,
  proto.autokitteh.eventsvc.ListEventsResponse,
  /**
   * @param {!proto.autokitteh.eventsvc.ListEventsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.eventsvc.ListEventsResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.eventsvc.ListEventsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.eventsvc.ListEventsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.eventsvc.ListEventsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.eventsvc.EventsClient.prototype.listEvents =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/ListEvents',
      request,
      metadata || {},
      methodDescriptor_Events_ListEvents,
      callback);
};


/**
 * @param {!proto.autokitteh.eventsvc.ListEventsRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.eventsvc.ListEventsResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.eventsvc.EventsPromiseClient.prototype.listEvents =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/ListEvents',
      request,
      metadata || {},
      methodDescriptor_Events_ListEvents);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.eventsvc.GetEventStateForProjectRequest,
 *   !proto.autokitteh.eventsvc.GetEventStateForProjectResponse>}
 */
const methodDescriptor_Events_GetEventStateForProject = new grpc.web.MethodDescriptor(
  '/autokitteh.eventsvc.Events/GetEventStateForProject',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.eventsvc.GetEventStateForProjectRequest,
  proto.autokitteh.eventsvc.GetEventStateForProjectResponse,
  /**
   * @param {!proto.autokitteh.eventsvc.GetEventStateForProjectRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.eventsvc.GetEventStateForProjectResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.eventsvc.GetEventStateForProjectRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.eventsvc.GetEventStateForProjectResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.eventsvc.GetEventStateForProjectResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.eventsvc.EventsClient.prototype.getEventStateForProject =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/GetEventStateForProject',
      request,
      metadata || {},
      methodDescriptor_Events_GetEventStateForProject,
      callback);
};


/**
 * @param {!proto.autokitteh.eventsvc.GetEventStateForProjectRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.eventsvc.GetEventStateForProjectResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.eventsvc.EventsPromiseClient.prototype.getEventStateForProject =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/GetEventStateForProject',
      request,
      metadata || {},
      methodDescriptor_Events_GetEventStateForProject);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.eventsvc.UpdateEventStateForProjectRequest,
 *   !proto.autokitteh.eventsvc.UpdateEventStateForProjectResponse>}
 */
const methodDescriptor_Events_UpdateEventStateForProject = new grpc.web.MethodDescriptor(
  '/autokitteh.eventsvc.Events/UpdateEventStateForProject',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.eventsvc.UpdateEventStateForProjectRequest,
  proto.autokitteh.eventsvc.UpdateEventStateForProjectResponse,
  /**
   * @param {!proto.autokitteh.eventsvc.UpdateEventStateForProjectRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.eventsvc.UpdateEventStateForProjectResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.eventsvc.UpdateEventStateForProjectRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.eventsvc.UpdateEventStateForProjectResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.eventsvc.UpdateEventStateForProjectResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.eventsvc.EventsClient.prototype.updateEventStateForProject =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/UpdateEventStateForProject',
      request,
      metadata || {},
      methodDescriptor_Events_UpdateEventStateForProject,
      callback);
};


/**
 * @param {!proto.autokitteh.eventsvc.UpdateEventStateForProjectRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.eventsvc.UpdateEventStateForProjectResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.eventsvc.EventsPromiseClient.prototype.updateEventStateForProject =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/UpdateEventStateForProject',
      request,
      metadata || {},
      methodDescriptor_Events_UpdateEventStateForProject);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.eventsvc.GetProjectWaitingEventsRequest,
 *   !proto.autokitteh.eventsvc.GetProjectWaitingEventsResponse>}
 */
const methodDescriptor_Events_GetProjectWaitingEvents = new grpc.web.MethodDescriptor(
  '/autokitteh.eventsvc.Events/GetProjectWaitingEvents',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.eventsvc.GetProjectWaitingEventsRequest,
  proto.autokitteh.eventsvc.GetProjectWaitingEventsResponse,
  /**
   * @param {!proto.autokitteh.eventsvc.GetProjectWaitingEventsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.eventsvc.GetProjectWaitingEventsResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.eventsvc.GetProjectWaitingEventsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.eventsvc.GetProjectWaitingEventsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.eventsvc.GetProjectWaitingEventsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.eventsvc.EventsClient.prototype.getProjectWaitingEvents =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/GetProjectWaitingEvents',
      request,
      metadata || {},
      methodDescriptor_Events_GetProjectWaitingEvents,
      callback);
};


/**
 * @param {!proto.autokitteh.eventsvc.GetProjectWaitingEventsRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.eventsvc.GetProjectWaitingEventsResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.eventsvc.EventsPromiseClient.prototype.getProjectWaitingEvents =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.eventsvc.Events/GetProjectWaitingEvents',
      request,
      metadata || {},
      methodDescriptor_Events_GetProjectWaitingEvents);
};


module.exports = proto.autokitteh.eventsvc;

