// Package events implements handlers for Slack's Events API:
// see https://api.slack.com/apis/connections/events-api and
// https://api.slack.com/events?ref=apis&filter=Events.
//
// Each event handler is required to return within 3 seconds, even if its
// response is an empty acknowledgement (an HTTP 200 status with no payload)
// which continues behind the scenes with further asynchronous processing.
// see https://api.slack.com/apis/connections/events-api#responding.
package events
