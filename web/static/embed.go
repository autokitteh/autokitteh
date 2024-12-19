package static

import (
	"embed"
)

//go:embed images/* *.png *.svg *.html *.ico *.txt *.webmanifest *.css *.js
var RootWebContent embed.FS

//go:embed asana/connect
var AsanaWebContent embed.FS

//go:embed aws/connect
var AWSWebContent embed.FS

//go:embed chatgpt/connect
var ChatGPTWebContent embed.FS

//go:embed confluence/connect
var ConfluenceWebContent embed.FS

//go:embed discord/connect
var DiscordWebContent embed.FS

//go:embed github/connect
var GitHubWebContent embed.FS

//go:embed gmail/connect
var GmailWebContent embed.FS

//go:embed google/connect
var GoogleWebContent embed.FS

//go:embed googlecalendar/connect
var GoogleCalendarWebContent embed.FS

//go:embed google/connect
var GoogleChatWebContent embed.FS

//go:embed googledrive/connect
var GoogleDriveWebContent embed.FS

//go:embed googleforms/connect
var GoogleFormsWebContent embed.FS

//go:embed googlegemini/connect
var GoogleGeminiWebContent embed.FS

//go:embed googlesheets/connect
var GoogleSheetsWebContent embed.FS

//go:embed i/http/connect
var HTTPWebContent embed.FS

//go:embed hubspot/connect
var HubSpotWebContent embed.FS

//go:embed jira/connect
var JiraWebContent embed.FS

//go:embed auth0/connect
var Auth0WebContent embed.FS

//go:embed slack/connect
var SlackWebContent embed.FS

//go:embed twilio/connect
var TwilioWebContent embed.FS
