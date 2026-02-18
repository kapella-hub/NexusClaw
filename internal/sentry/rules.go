package sentry

import (
	"context"
	"regexp"
)

// RuleEngine evaluates firewall rules against requests.
type RuleEngine interface {
	Evaluate(ctx context.Context, action, resource string) (bool, error)
}

type ruleEngine struct {
	repo Repository
}

// NewRuleEngine creates a new DB-backed rule engine.
func NewRuleEngine(repo Repository) RuleEngine {
	return &ruleEngine{repo: repo}
}

// Evaluate loads all enabled rules and checks them against the given action and
// resource. Returns false if any blocking rule matches, true otherwise.
func (re *ruleEngine) Evaluate(ctx context.Context, action, resource string) (bool, error) {
	rules, err := re.repo.ListRules(ctx)
	if err != nil {
		return false, err
	}

	subject := action + ":" + resource

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		matched, err := regexp.MatchString(rule.Pattern, subject)
		if err != nil {
			// Invalid pattern; skip this rule rather than blocking everything.
			continue
		}
		if matched && rule.Action == "block" {
			return false, nil
		}
	}

	return true, nil
}
