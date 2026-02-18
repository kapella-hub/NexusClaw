package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage MCP server nodes",
}

var nodeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered MCP servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		data, status, err := client.get("/api/v1/nodes")
		if err != nil {
			return err
		}
		if checkError(data, status) {
			return nil
		}

		var servers []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Image  string `json:"image"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal(data, &servers); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tIMAGE\tSTATUS")
		for _, s := range servers {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.ID, s.Name, s.Image, s.Status)
		}
		return w.Flush()
	},
}

var nodeRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new MCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		image, _ := cmd.Flags().GetString("image")

		client := newAPIClient()
		data, status, err := client.post("/api/v1/nodes", map[string]string{
			"name":  name,
			"image": image,
		})
		if err != nil {
			return err
		}
		if checkError(data, status) {
			return nil
		}

		var server struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Image  string `json:"image"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal(data, &server); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		fmt.Printf("ID:     %s\nName:   %s\nImage:  %s\nStatus: %s\n", server.ID, server.Name, server.Image, server.Status)
		return nil
	},
}

var nodeStartCmd = &cobra.Command{
	Use:   "start [id]",
	Short: "Start an MCP server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		data, status, err := client.post("/api/v1/nodes/"+args[0]+"/start", nil)
		if err != nil {
			return err
		}
		if checkError(data, status) {
			return nil
		}

		var resp struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		fmt.Printf("Status: %s\n", resp.Status)
		return nil
	},
}

var nodeStopCmd = &cobra.Command{
	Use:   "stop [id]",
	Short: "Stop an MCP server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		data, status, err := client.post("/api/v1/nodes/"+args[0]+"/stop", nil)
		if err != nil {
			return err
		}
		if checkError(data, status) {
			return nil
		}

		var resp struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		fmt.Printf("Status: %s\n", resp.Status)
		return nil
	},
}

var nodeRemoveCmd = &cobra.Command{
	Use:   "remove [id]",
	Short: "Remove an MCP server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		data, status, err := client.delete("/api/v1/nodes/" + args[0])
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
	nodeRegisterCmd.Flags().String("name", "", "server name")
	nodeRegisterCmd.Flags().String("image", "", "container image")
	nodeRegisterCmd.MarkFlagRequired("name")
	nodeRegisterCmd.MarkFlagRequired("image")

	nodeCmd.AddCommand(nodeListCmd, nodeRegisterCmd, nodeStartCmd, nodeStopCmd, nodeRemoveCmd)
	rootCmd.AddCommand(nodeCmd)
}
