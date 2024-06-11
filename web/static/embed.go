package static

import (
	"embed"
)

//go:embed images/* *.png *.svg *.html *.ico *.txt *.webmanifest *.css *.js
var RootWebContent embed.FS

//go:embed aws/connect
var AWSWebContent embed.FS

//go:embed chatgpt/connect
var ChatGPTWebContent embed.FS

//go:embed github/connect
var GitHubWebContent embed.FS

//go:embed gmail/connect
var GmailWebContent embed.FS

//go:embed google/connect
var GoogleWebContent embed.FS

//go:embed google/connect
var GoogleCalendarWebContent embed.FS

//go:embed google/connect
var GoogleChatWebContent embed.FS

//go:embed google/connect
var GoogleDriveWebContent embed.FS

//go:embed google/connect
var GoogleFormsWebContent embed.FS

//go:embed googlesheets/connect
var GoogleSheetsWebContent embed.FS

//go:embed i/http/connect
var HTTPWebContent embed.FS

//go:embed jira/connect
var JiraWebContent embed.FS

//go:embed slack/connect
var SlackWebContent embed.FS

//go:embed twilio/connect
var TwilioWebContent embed.FS
