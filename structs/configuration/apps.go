package configuration

import "fmt"

type App struct {
	Name       string       `yaml:"name"`
	Bin        string       `yaml:"bin"`
	Ports      []int        `yaml:"ports"`
	ProxyRules []*ProxyRule `yaml:"proxy_rules"`
}

type ProxyRule struct {
	Type    string `yaml:"type"`
	Matcher string `yaml:"matcher"`
	Target  string `yaml:"target"`
}

func (r ProxyRule) Hash() string {
	// TODO this is not a foolproof way of uniquely identifying rules, need to fix
	return fmt.Sprintf("type-%s-matcher-%s-target-%s", r.Type, r.Matcher, r.Target)
}

type ProxyRuleType string

const (
	ProxyRuleFileServer = "file"
	ProxyRuleRedirect   = "redirect"
)
