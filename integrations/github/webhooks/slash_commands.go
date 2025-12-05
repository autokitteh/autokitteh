package webhooks

import (
	"strings"
	"unicode"

	"github.com/google/go-github/v52/github"
)

type slashCommand struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
	Raw  string   `json:"raw"`
}

func extractSlashCommands(event any) []slashCommand {
	var md string

	switch e := event.(type) {
	case *github.IssueCommentEvent:
		md = e.GetComment().GetBody()
	case *github.PullRequestReviewCommentEvent:
		md = e.GetComment().GetBody()
	case *github.PullRequestReviewEvent:
		md = e.GetReview().GetBody()
	case *github.PullRequestEvent:
		md = e.GetPullRequest().GetBody()
	case *github.IssuesEvent:
		md = e.GetIssue().GetBody()
	case *github.DiscussionEvent:
		md = e.GetDiscussion().GetBody()
	case *github.DiscussionCommentEvent:
		md = e.GetComment().GetBody()
	case *github.CommitCommentEvent:
		md = e.GetComment().GetBody()
	default:
		return nil
	}

	return extractSlashCommandsFromMD(md)
}

func extractSlashCommandsFromMD(md string) (commands []slashCommand) {
	for line := range strings.SplitSeq(md, "\n") {
		if !strings.HasPrefix(line, "/") {
			continue
		}

		line = strings.TrimRightFunc(line, unicode.IsSpace)

		if parts := strings.Fields(line); len(parts) > 0 {
			if parts[0] == "/" {
				continue
			}

			cmd := slashCommand{
				Name: parts[0][1:],
				Args: parts[1:],
				Raw:  line,
			}
			commands = append(commands, cmd)
		}
	}

	return
}
