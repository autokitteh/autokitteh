package webhooks

import (
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
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

	text := extractMDText(md)

	return parseSlashCommands(text)
}

func parseSlashCommands(md string) []slashCommand {
	var commands []slashCommand
	for line := range strings.SplitSeq(md, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "/") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				cmd := slashCommand{
					Name: strings.TrimPrefix(parts[0], "/"),
					Args: parts[1:],
					Raw:  line,
				}
				commands = append(commands, cmd)
			}
		}
	}
	return commands
}

func extractMDText(md string) string {
	p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs | parser.FencedCode)

	doc := p.Parse([]byte(md))

	var text strings.Builder

	// Walk the AST and extract only text nodes (not code)
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		switch n := node.(type) {
		case *ast.Text:
			text.Write(n.Literal)
		case *ast.Softbreak, *ast.Hardbreak:
			text.WriteString("\n")
		case *ast.Code, *ast.CodeBlock:
			return ast.SkipChildren
		}

		return ast.GoToNext
	})

	return text.String()
}
