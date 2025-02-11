package static

import (
	"embed"
)

//go:embed images/* *.png *.svg *.html *.ico *.txt *.webmanifest *.css *.js
var RootWebContent embed.FS

//go:embed asana
var AsanaWebContent embed.FS

//go:embed auth0
var Auth0WebContent embed.FS

//go:embed aws
var AWSWebContent embed.FS

//go:embed chatgpt
var ChatGPTWebContent embed.FS

//go:embed confluence
var ConfluenceWebContent embed.FS

//go:embed discord
var DiscordWebContent embed.FS

//go:embed github
var GitHubWebContent embed.FS

//go:embed gmail
var GmailWebContent embed.FS

//go:embed google
var GoogleWebContent embed.FS

//go:embed googlecalendar
var GoogleCalendarWebContent embed.FS

//go:embed google
var GoogleChatWebContent embed.FS

//go:embed googledrive
var GoogleDriveWebContent embed.FS

//go:embed googleforms
var GoogleFormsWebContent embed.FS

//go:embed googlegemini
var GoogleGeminiWebContent embed.FS

//go:embed googlesheets
var GoogleSheetsWebContent embed.FS

//go:embed height
var HeightWebContent embed.FS

//go:embed hubspot
var HubSpotWebContent embed.FS

//go:embed jira
var JiraWebContent embed.FS

//go:embed linear
var LinearWebContent embed.FS

//go:embed microsoft
var MicrosoftWebContent embed.FS

//go:embed salesforce
var SalesforceWebContent embed.FS

//go:embed slack
var SlackWebContent embed.FS

//go:embed twilio
var TwilioWebContent embed.FS

//go:embed zoom
var ZoomWebContent embed.FS
