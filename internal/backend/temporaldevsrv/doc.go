// Adapted from go.temporal.io/sdk@v1.33.0/testsuite.
//
// Main changes:
// - Separate download and start. Start requires prior download.
// - Default dir for download is now xdg.CacheHomeDir().
package temporaldevsrv
