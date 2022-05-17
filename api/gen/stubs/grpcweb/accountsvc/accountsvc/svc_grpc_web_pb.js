/**
 * @fileoverview gRPC-Web generated client stub for autokitteh.accountsvc
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

var account_account_pb = require('../account/account_pb.js')
const proto = {};
proto.autokitteh = {};
proto.autokitteh.accountsvc = require('./svc_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.autokitteh.accountsvc.AccountsClient =
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
proto.autokitteh.accountsvc.AccountsPromiseClient =
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
 *   !proto.autokitteh.accountsvc.CreateAccountRequest,
 *   !proto.autokitteh.accountsvc.CreateAccountResponse>}
 */
const methodDescriptor_Accounts_CreateAccount = new grpc.web.MethodDescriptor(
  '/autokitteh.accountsvc.Accounts/CreateAccount',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.accountsvc.CreateAccountRequest,
  proto.autokitteh.accountsvc.CreateAccountResponse,
  /**
   * @param {!proto.autokitteh.accountsvc.CreateAccountRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.accountsvc.CreateAccountResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.accountsvc.CreateAccountRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.accountsvc.CreateAccountResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.accountsvc.CreateAccountResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.accountsvc.AccountsClient.prototype.createAccount =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.accountsvc.Accounts/CreateAccount',
      request,
      metadata || {},
      methodDescriptor_Accounts_CreateAccount,
      callback);
};


/**
 * @param {!proto.autokitteh.accountsvc.CreateAccountRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.accountsvc.CreateAccountResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.accountsvc.AccountsPromiseClient.prototype.createAccount =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.accountsvc.Accounts/CreateAccount',
      request,
      metadata || {},
      methodDescriptor_Accounts_CreateAccount);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.accountsvc.UpdateAccountRequest,
 *   !proto.autokitteh.accountsvc.UpdateAccountResponse>}
 */
const methodDescriptor_Accounts_UpdateAccount = new grpc.web.MethodDescriptor(
  '/autokitteh.accountsvc.Accounts/UpdateAccount',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.accountsvc.UpdateAccountRequest,
  proto.autokitteh.accountsvc.UpdateAccountResponse,
  /**
   * @param {!proto.autokitteh.accountsvc.UpdateAccountRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.accountsvc.UpdateAccountResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.accountsvc.UpdateAccountRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.accountsvc.UpdateAccountResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.accountsvc.UpdateAccountResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.accountsvc.AccountsClient.prototype.updateAccount =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.accountsvc.Accounts/UpdateAccount',
      request,
      metadata || {},
      methodDescriptor_Accounts_UpdateAccount,
      callback);
};


/**
 * @param {!proto.autokitteh.accountsvc.UpdateAccountRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.accountsvc.UpdateAccountResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.accountsvc.AccountsPromiseClient.prototype.updateAccount =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.accountsvc.Accounts/UpdateAccount',
      request,
      metadata || {},
      methodDescriptor_Accounts_UpdateAccount);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.accountsvc.GetAccountRequest,
 *   !proto.autokitteh.accountsvc.GetAccountResponse>}
 */
const methodDescriptor_Accounts_GetAccount = new grpc.web.MethodDescriptor(
  '/autokitteh.accountsvc.Accounts/GetAccount',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.accountsvc.GetAccountRequest,
  proto.autokitteh.accountsvc.GetAccountResponse,
  /**
   * @param {!proto.autokitteh.accountsvc.GetAccountRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.accountsvc.GetAccountResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.accountsvc.GetAccountRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.accountsvc.GetAccountResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.accountsvc.GetAccountResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.accountsvc.AccountsClient.prototype.getAccount =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.accountsvc.Accounts/GetAccount',
      request,
      metadata || {},
      methodDescriptor_Accounts_GetAccount,
      callback);
};


/**
 * @param {!proto.autokitteh.accountsvc.GetAccountRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.accountsvc.GetAccountResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.accountsvc.AccountsPromiseClient.prototype.getAccount =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.accountsvc.Accounts/GetAccount',
      request,
      metadata || {},
      methodDescriptor_Accounts_GetAccount);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.autokitteh.accountsvc.GetAccountsRequest,
 *   !proto.autokitteh.accountsvc.GetAccountsResponse>}
 */
const methodDescriptor_Accounts_GetAccounts = new grpc.web.MethodDescriptor(
  '/autokitteh.accountsvc.Accounts/GetAccounts',
  grpc.web.MethodType.UNARY,
  proto.autokitteh.accountsvc.GetAccountsRequest,
  proto.autokitteh.accountsvc.GetAccountsResponse,
  /**
   * @param {!proto.autokitteh.accountsvc.GetAccountsRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.autokitteh.accountsvc.GetAccountsResponse.deserializeBinary
);


/**
 * @param {!proto.autokitteh.accountsvc.GetAccountsRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.autokitteh.accountsvc.GetAccountsResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.autokitteh.accountsvc.GetAccountsResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.autokitteh.accountsvc.AccountsClient.prototype.getAccounts =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/autokitteh.accountsvc.Accounts/GetAccounts',
      request,
      metadata || {},
      methodDescriptor_Accounts_GetAccounts,
      callback);
};


/**
 * @param {!proto.autokitteh.accountsvc.GetAccountsRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.autokitteh.accountsvc.GetAccountsResponse>}
 *     Promise that resolves to the response
 */
proto.autokitteh.accountsvc.AccountsPromiseClient.prototype.getAccounts =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/autokitteh.accountsvc.Accounts/GetAccounts',
      request,
      metadata || {},
      methodDescriptor_Accounts_GetAccounts);
};


module.exports = proto.autokitteh.accountsvc;

