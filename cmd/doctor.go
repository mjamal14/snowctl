package cmd

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/mjamalu/snowctl/internal/client"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check connectivity and authentication with ServiceNow",
	RunE:  runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	ctx, err := cfg.ActiveContext()
	if err != nil {
		return err
	}

	fmt.Printf("Checking context: %s\n", ctx.Name)
	fmt.Printf("Instance:         %s\n\n", ctx.Instance)

	// Check 1: DNS resolution
	host := strings.TrimPrefix(ctx.Instance, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.Split(host, "/")[0]

	fmt.Printf("[1/3] DNS resolution for %s... ", host)
	addrs, err := net.LookupHost(host)
	if err != nil {
		fmt.Printf("FAIL\n      %s\n", err)
		return nil
	}
	fmt.Printf("OK (%s)\n", addrs[0])

	// Check 2: HTTPS connectivity
	fmt.Printf("[2/3] HTTPS connectivity... ")
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(ctx.Instance)
	if err != nil {
		fmt.Printf("FAIL\n      %s\n", err)
		return nil
	}
	resp.Body.Close()
	fmt.Printf("OK (HTTP %d)\n", resp.StatusCode)

	// Check 3: API authentication
	fmt.Printf("[3/3] API authentication... ")
	c, err := getClient()
	if err != nil {
		fmt.Printf("FAIL\n      %s\n", err)
		return nil
	}

	_, listErr := c.List("sys_properties", client.ListOptions{Limit: 1})
	if listErr != nil {
		fmt.Printf("FAIL\n      %s\n", listErr)
		return nil
	}
	fmt.Printf("OK\n")

	fmt.Printf("\nAll checks passed. snowctl is ready to use.\n")
	return nil
}
