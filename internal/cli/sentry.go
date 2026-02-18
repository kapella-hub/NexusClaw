package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var sentryCmd = &cobra.Command{
	Use:   "sentry",
	Short: "Manage AI Sentry (firewall & auditing)",
}

var sentryAuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "View audit logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		data, status, err := client.get("/api/v1/sentry/audit")
		if err != nil {
			return err
		}
		if checkError(data, status) {
			return nil
		}

		var entries []struct {
			ID        string `json:"id"`
			Action    string `json:"action"`
			Resource  string `json:"resource"`
			CreatedAt string `json:"created_at"`
		}
		if err := json.Unmarshal(data, &entries); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tACTION\tRESOURCE\tCREATED AT")
		for _, e := range entries {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.ID, e.Action, e.Resource, e.CreatedAt)
		}
		return w.Flush()
	},
}

var sentryRulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Manage firewall rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		data, status, err := client.get("/api/v1/sentry/rules")
		if err != nil {
			return err
		}
		if checkError(data, status) {
			return nil
		}

		var rules []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Pattern string `json:"pattern"`
			Action  string `json:"action"`
			Enabled bool   `json:"enabled"`
		}
		if err := json.Unmarshal(data, &rules); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tPATTERN\tACTION\tENABLED")
		for _, r := range rules {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.ID, r.Name, r.Pattern, r.Action, strconv.FormatBool(r.Enabled))
		}
		return w.Flush()
	},
}

var sentryRulesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a firewall rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		pattern, _ := cmd.Flags().GetString("pattern")
		action, _ := cmd.Flags().GetString("action")

		client := newAPIClient()
		data, status, err := client.post("/api/v1/sentry/rules", map[string]any{
			"name":    name,
			"pattern": pattern,
			"action":  action,
			"enabled": true,
		})
		if err != nil {
			return err
		}
		if checkError(data, status) {
			return nil
		}

		var rule struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Pattern string `json:"pattern"`
			Action  string `json:"action"`
			Enabled bool   `json:"enabled"`
		}
		if err := json.Unmarshal(data, &rule); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		fmt.Printf("ID:      %s\nName:    %s\nPattern: %s\nAction:  %s\nEnabled: %t\n",
			rule.ID, rule.Name, rule.Pattern, rule.Action, rule.Enabled)
		return nil
	},
}

var sentryBudgetCmd = &cobra.Command{
	Use:   "budget",
	Short: "Manage token budgets",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		data, status, err := client.get("/api/v1/sentry/budget")
		if err != nil {
			return err
		}
		if checkError(data, status) {
			return nil
		}

		var budget struct {
			ID         string `json:"id"`
			Period     string `json:"period"`
			MaxTokens  int64  `json:"max_tokens"`
			UsedTokens int64  `json:"used_tokens"`
			ResetAt    string `json:"reset_at"`
		}
		if err := json.Unmarshal(data, &budget); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		fmt.Printf("Period:      %s\nMax Tokens:  %d\nUsed Tokens: %d\nReset At:    %s\n",
			budget.Period, budget.MaxTokens, budget.UsedTokens, budget.ResetAt)
		return nil
	},
}

var sentryBudgetSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set token budget limits",
	RunE: func(cmd *cobra.Command, args []string) error {
		maxTokens, _ := cmd.Flags().GetInt64("max-tokens")
		period, _ := cmd.Flags().GetString("period")

		client := newAPIClient()
		data, status, err := client.put("/api/v1/sentry/budget", map[string]any{
			"max_tokens": maxTokens,
			"period":     period,
		})
		if err != nil {
			return err
		}
		if checkError(data, status) {
			return nil
		}

		var budget struct {
			Period     string `json:"period"`
			MaxTokens  int64  `json:"max_tokens"`
			UsedTokens int64  `json:"used_tokens"`
		}
		if err := json.Unmarshal(data, &budget); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		fmt.Printf("Period:      %s\nMax Tokens:  %d\nUsed Tokens: %d\n",
			budget.Period, budget.MaxTokens, budget.UsedTokens)
		return nil
	},
}

func init() {
	sentryRulesAddCmd.Flags().String("name", "", "rule name")
	sentryRulesAddCmd.Flags().String("pattern", "", "match pattern")
	sentryRulesAddCmd.Flags().String("action", "block", "rule action (block, allow, alert)")
	sentryRulesAddCmd.MarkFlagRequired("name")
	sentryRulesAddCmd.MarkFlagRequired("pattern")

	sentryBudgetSetCmd.Flags().Int64("max-tokens", 0, "maximum token count")
	sentryBudgetSetCmd.Flags().String("period", "monthly", "budget period (daily, weekly, monthly)")
	sentryBudgetSetCmd.MarkFlagRequired("max-tokens")

	sentryRulesCmd.AddCommand(sentryRulesAddCmd)
	sentryBudgetCmd.AddCommand(sentryBudgetSetCmd)
	sentryCmd.AddCommand(sentryAuditCmd, sentryRulesCmd, sentryBudgetCmd)
	rootCmd.AddCommand(sentryCmd)
}
