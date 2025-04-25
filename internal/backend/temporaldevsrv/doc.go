// Package temporaldevsrv provides utilities for managing a Temporal development server.
//
// This module is designed to simplify the process of downloading, configuring, starting,
// and interacting with a Temporal CLI-based development server. It is particularly useful
// for local development and testing scenarios where a lightweight Temporal server is needed.
//
// Key Features:
//   - Download Management: Handles downloading the Temporal development server binary,
//     with support for specifying versions and caching the binary in a default directory.
//   - Server Configuration: Provides options to configure the server, including host/port,
//     namespace, logging, and UI settings.
//   - Server Lifecycle: Includes functions to start, stop, and monitor the server's status.
//   - Integration: Designed to integrate seamlessly with other modules, such as the Temporal
//     client, to provide a complete development environment.
//
// This module is adapted from go.temporal.io/sdk@v1.33.0/testsuite with modifications to
// separate the download and start processes, and to use xdg.CacheHomeDir() as the default
// directory for downloads.

package temporaldevsrv
