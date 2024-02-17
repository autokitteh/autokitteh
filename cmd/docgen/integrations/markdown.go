package integrations

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	integrationsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1"
)

type markdownGenerator struct {
	outputDir string
}

func (g markdownGenerator) Output() string {
	return fmt.Sprintf("Markdown documents (%s)", g.outputDir)
}

func NewMarkdownGenerator(rootDir string) Generator {
	outputDir := filepath.Join(rootDir, "int/mdx")
	if err := resetDir(outputDir); err != nil {
		log.Fatal(err)
	}

	return markdownGenerator{outputDir: outputDir}
}

//go:embed templates/mdx_doc.tmpl
var mdxTemplate string

func (g markdownGenerator) Generate(akURL string, n int, i *integrationsv1.Integration) {
	// Template notes:
	// - Why specify a sidebar position in the front matter? To order the
	//   integrations in the sidebar based on display name, not unique ID
	// - Why use tabs? Multi-language support ("snake_case" vs. "camelCase")
	// - Reminder: string-key maps are visited in sorted key order
	// - TODO: Unused for now:
	//   - $func.Examples
	//   - $input|output.Optional
	//   - $input|output.Kwarg
	//   - $input|output.Examples
	funcMap := template.FuncMap{
		"trimSpace": strings.TrimSpace,
		"refLabel":  stripNumPrefix,
		"funcName":  capitalize,
	}
	t, err := template.New("mdx_doc").Funcs(funcMap).Parse(mdxTemplate)
	if err != nil {
		log.Fatal(err)
	}

	data := struct {
		Pos int
		Int *integrationsv1.Integration
	}{
		Pos: n,
		Int: i,
	}
	b := new(bytes.Buffer)
	if err := t.Execute(b, data); err != nil {
		log.Fatal(err)
	}

	path := filepath.Join(g.outputDir, i.UniqueName+".mdx")
	if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
}

// Use numbers to sort reference link maps, but don't display them.
func stripNumPrefix(s string) string {
	return strings.TrimSpace(regexp.MustCompile(`^\d+\s*`).ReplaceAllString(s, ""))
}

// Convert "snake_case" function names to "Title Case" headers.
func capitalize(s string) string {
	caser := cases.Title(language.AmericanEnglish)
	s = strings.TrimSpace(caser.String(strings.ReplaceAll(s, "_", " ")))

	for from, to := range map[string]string{
		// Special handling for prepositions in AWS function names.
		" By ":   " by ",
		" For ":  " for ",
		" From ": " from ",
		" To ":   " to ",

		// Special handling for acronyms in AWS function names.
		"Acl":         "ACL",
		"Api":         "API",
		"Aws":         "AWS",
		"Byoasn":      "BYOASN",
		"Byoip":       "BYOIP",
		"Cidr":        "CIDR",
		"Coip":        "CoIP",
		"Cors":        "CORS",
		"Db":          "DB",
		"Dhcp":        "DHCP",
		"Ebs":         "EBS",
		"Ec2":         "EC2",
		"Eventbridge": "EventBridge",
		"Fpga":        "FPGA",
		"Iam":         "IAM",
		"Ip":          "IP",
		"IPam":        "IPAM",
		"Kms":         "KMS",
		"Mfa":         "MFA",
		"Nat":         "NAT",
		"Rds":         "RDS",
		"Saml":        "SAML",
		"Sms":         "SMS",
		"Sns":         "SNS",
		"Sql":         "SQL",
		"Sqs":         "SQS",
		"Ssh":         "SSH",
		"Vgw":         "VGW",
		"Vpc":         "VPC",
		"Vpn":         "VPN",
	} {
		s = strings.ReplaceAll(s, from, to)
	}
	return s
}
