package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var passCmd = &cobra.Command{
	Use:   "pass",
	Short: "Manage Nexus Pass (credential vault)",
}

var passLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate and obtain a session token",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		client := newAPIClient()
		data, status, err := client.post("/api/v1/pass/sessions", map[string]string{
			"email":    email,
			"password": password,
		})
		if err != nil {
			return err
		}
		if checkError(data, status) {
			return nil
		}

		var session struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal(data, &session); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		fmt.Println(session.Token)
		return nil
	},
}

var passRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user account",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		client := newAPIClient()
		data, status, err := client.post("/api/v1/pass/register", map[string]string{
			"email":    email,
			"password": password,
		})
		if err != nil {
			return err
		}
		if checkError(data, status) {
			return nil
		}

		var user struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}
		if err := json.Unmarshal(data, &user); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		fmt.Printf("ID:    %s\nEmail: %s\n", user.ID, user.Email)
		return nil
	},
}

var passListCmd = &cobra.Command{
	Use:   "list",
	Short: "List stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		data, status, err := client.get("/api/v1/pass/vault")
		if err != nil {
			return err
		}
		if checkError(data, status) {
			return nil
		}

		var entries []struct {
			ID        string `json:"id"`
			Provider  string `json:"provider"`
			CreatedAt string `json:"created_at"`
		}
		if err := json.Unmarshal(data, &entries); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tPROVIDER\tCREATED AT")
		for _, e := range entries {
			fmt.Fprintf(w, "%s\t%s\t%s\n", e.ID, e.Provider, e.CreatedAt)
		}
		return w.Flush()
	},
}

var passAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a credential to the vault",
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, _ := cmd.Flags().GetString("provider")
		accessToken, _ := cmd.Flags().GetString("access-token")
		secret, _ := cmd.Flags().GetString("secret")

		client := newAPIClient()
		data, status, err := client.post("/api/v1/pass/vault", map[string]string{
			"provider":     provider,
			"access_token": accessToken,
			"secret":       secret,
		})
		if err != nil {
			return err
		}
		if checkError(data, status) {
			return nil
		}

		var entry struct {
			ID        string `json:"id"`
			Provider  string `json:"provider"`
			CreatedAt string `json:"created_at"`
		}
		if err := json.Unmarshal(data, &entry); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		fmt.Printf("ID:       %s\nProvider: %s\nCreated:  %s\n", entry.ID, entry.Provider, entry.CreatedAt)
		return nil
	},
}

var passRemoveCmd = &cobra.Command{
	Use:   "remove [id]",
	Short: "Remove a credential from the vault",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		data, status, err := client.delete("/api/v1/pass/vault/" + args[0])
		if err != nil {
			return err
		}
		if checkError(data, status) {
			return nil
		}

		fmt.Println("Removed")
		return nil
	},
}

func init() {
	passLoginCmd.Flags().String("email", "", "user email address")
	passLoginCmd.Flags().String("password", "", "user password")
	passLoginCmd.MarkFlagRequired("email")
	passLoginCmd.MarkFlagRequired("password")

	passRegisterCmd.Flags().String("email", "", "user email address")
	passRegisterCmd.Flags().String("password", "", "user password")
	passRegisterCmd.MarkFlagRequired("email")
	passRegisterCmd.MarkFlagRequired("password")

	passAddCmd.Flags().String("provider", "", "provider name")
	passAddCmd.Flags().String("access-token", "", "access token")
	passAddCmd.Flags().String("secret", "", "secret key")
	passAddCmd.MarkFlagRequired("provider")

	passCmd.AddCommand(passLoginCmd, passRegisterCmd, passListCmd, passAddCmd, passRemoveCmd)
	rootCmd.AddCommand(passCmd)
}
