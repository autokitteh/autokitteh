# Life of a Connection

A connection is a way for a project to access an integration with a specific configuration.

## Flow

1. A connection is created using the _Connections_ service with the following data specified:
   - **Name**, this is used to reference the connection in code and configuration
   - **Integration** that this connection is representing
   - **Project** that this connection serves
2. Most connections need to be configured. The configuration is persisted as **Connection Variables**. These are very much like environment variables, but resided in a scope of a specific connection. This is done using the _Vars_ service. The variables can be set either automatically by the integration, using the UI or manually by the user via APIs. One example of such configuration is OAuth flow, described below.
3. Now the connection is ready to be used. The user reference the connection from their code. Once a new session is created, the respective integration is contacted using the `Integrations.Configure` call, and given the connection ID. The integration then looks up the connection variables for this connection, and returns values that are set up for this specific connection.
4. When a connection function call is made, the integration receives the function value, decodes the function's `Data` field (which usually contains just the connection ID, but this doesn't have to be always true), and potentially accesses the relevant connection variables as well. Then once it has the data it requires.

## OAuth

Some connections require OAuth authentication in order to function. Autokitteh provides an _OAuth_ service that knows the steps of the OAuth dance.

Every such integration has a web page that the user can use to start the OAuth flow. Autokitteh's OAuth service then does the actual OAuth dance, and once done, redirects to an integration specific page where the integration can process the OAuth result, which might contain some extra parameters from the OAuth provider. As a result of this processing, the integration generates a set of connection variables that contain the data it needs to persist. The integration then redirects again to Autokitteh's integration finalization stage where the variables are set.
