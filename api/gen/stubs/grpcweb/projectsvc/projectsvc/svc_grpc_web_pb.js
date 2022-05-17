/**
 * @fileoverview gRPC-Web generated client stub for autokitteh.projectsvc
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

var project_project_pb = require('../project/project_pb.js')
const proto = {};
proto.autokitteh = {};
proto.autokitteh.projectsvc = require('./svc_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.projectsvc.ProjectsClient =
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
proto.autokitteh.projectsvc.ProjectsPromiseClient =
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
 *   !proto.autokitteh.projectsvc.CreateProjectRequest,
 *   !proto.autokitteh.projectsvc.CreateProjectResponse>}
 */
const methodDescriptor_Projects_CreateProject = new grpc.web.MethodDescriptor(
  '/autokitteh.projectsvc.Projects/CreateProject',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.projectsvc.CreateProjectRequest,
  proto.autokitteh.projectsvc.CreateProjectResponse,
  /**
   * @param {!proto.autokitteh.projectsvc.CreateProjectRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.projectsvc.CreateProjectResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.projectsvc.CreateProjectRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.projectsvc.CreateProjectResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.projectsvc.CreateProjectResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.projectsvc.ProjectsClient.prototype.createProject =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.projectsvc.Projects/CreateProject',
      request,
      metadata || {},
      methodDescriptor_Projects_CreateProject,
      callback);
};


/**
 * @param {!proto.autokitteh.projectsvc.CreateProjectRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.projectsvc.CreateProjectResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.projectsvc.ProjectsPromiseClient.prototype.createProject =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.projectsvc.Projects/CreateProject',
      request,
      metadata || {},
      methodDescriptor_Projects_CreateProject);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.projectsvc.UpdateProjectRequest,
 *   !proto.autokitteh.projectsvc.UpdateProjectResponse>}
 */
const methodDescriptor_Projects_UpdateProject = new grpc.web.MethodDescriptor(
  '/autokitteh.projectsvc.Projects/UpdateProject',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.projectsvc.UpdateProjectRequest,
  proto.autokitteh.projectsvc.UpdateProjectResponse,
  /**
   * @param {!proto.autokitteh.projectsvc.UpdateProjectRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.projectsvc.UpdateProjectResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.projectsvc.UpdateProjectRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.projectsvc.UpdateProjectResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.projectsvc.UpdateProjectResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.projectsvc.ProjectsClient.prototype.updateProject =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.projectsvc.Projects/UpdateProject',
      request,
      metadata || {},
      methodDescriptor_Projects_UpdateProject,
      callback);
};


/**
 * @param {!proto.autokitteh.projectsvc.UpdateProjectRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.projectsvc.UpdateProjectResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.projectsvc.ProjectsPromiseClient.prototype.updateProject =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.projectsvc.Projects/UpdateProject',
      request,
      metadata || {},
      methodDescriptor_Projects_UpdateProject);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.projectsvc.GetProjectRequest,
 *   !proto.autokitteh.projectsvc.GetProjectResponse>}
 */
const methodDescriptor_Projects_GetProject = new grpc.web.MethodDescriptor(
  '/autokitteh.projectsvc.Projects/GetProject',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.projectsvc.GetProjectRequest,
  proto.autokitteh.projectsvc.GetProjectResponse,
  /**
   * @param {!proto.autokitteh.projectsvc.GetProjectRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.projectsvc.GetProjectResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.projectsvc.GetProjectRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.projectsvc.GetProjectResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.projectsvc.GetProjectResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.projectsvc.ProjectsClient.prototype.getProject =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.projectsvc.Projects/GetProject',
      request,
      metadata || {},
      methodDescriptor_Projects_GetProject,
      callback);
};


/**
 * @param {!proto.autokitteh.projectsvc.GetProjectRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.projectsvc.GetProjectResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.projectsvc.ProjectsPromiseClient.prototype.getProject =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.projectsvc.Projects/GetProject',
      request,
      metadata || {},
      methodDescriptor_Projects_GetProject);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.projectsvc.GetProjectsRequest,
 *   !proto.autokitteh.projectsvc.GetProjectsResponse>}
 */
const methodDescriptor_Projects_GetProjects = new grpc.web.MethodDescriptor(
  '/autokitteh.projectsvc.Projects/GetProjects',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.projectsvc.GetProjectsRequest,
  proto.autokitteh.projectsvc.GetProjectsResponse,
  /**
   * @param {!proto.autokitteh.projectsvc.GetProjectsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.projectsvc.GetProjectsResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.projectsvc.GetProjectsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.projectsvc.GetProjectsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.projectsvc.GetProjectsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.projectsvc.ProjectsClient.prototype.getProjects =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.projectsvc.Projects/GetProjects',
      request,
      metadata || {},
      methodDescriptor_Projects_GetProjects,
      callback);
};


/**
 * @param {!proto.autokitteh.projectsvc.GetProjectsRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.projectsvc.GetProjectsResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.projectsvc.ProjectsPromiseClient.prototype.getProjects =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.projectsvc.Projects/GetProjects',
      request,
      metadata || {},
      methodDescriptor_Projects_GetProjects);
};


module.exports = proto.autokitteh.projectsvc;

