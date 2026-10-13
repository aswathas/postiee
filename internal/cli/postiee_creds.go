// Hand-written credential loading. The generated auth login reads only
// --client-id/--client-secret flags or two env vars; this adds file support so
// the secret never has to appear in shell history or a process listing.
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// postieeCredEnvVars is where the generated login already looks.
const (
	postieeClientIDEnv     = "POSTIEE_CLIENT_ID"
	postieeClientSecretEnv = "POSTIEE_CLIENT_SECRET"
)

// postieeCredentialFiles are searched in order. The first that exists wins.
func postieeCredentialFiles() []string {
	var paths []string
	if v := strings.TrimSpace(os.Getenv("POSTIEE_CREDENTIALS_FILE")); v != "" {
		paths = append(paths, v)
	}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths,
			home+"/.config/postiee/credentials.toml",
			home+"/.config/postiee/credentials.env",
		)
	}
	return paths
}

// parseTOMLScalars pulls `key = "value"` pairs out of a small TOML file. Only
// flat scalar keys are supported, which is all a credentials file needs; a real
// TOML parser would be a dependency for nothing.
func parseTOMLScalars(src string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		// Strip trailing comment and surrounding quotes.
		if i := strings.Index(v, " #"); i >= 0 {
			v = strings.TrimSpace(v[:i])
		}
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if k != "" && v != "" {
			out[k] = v
		}
	}
	return out
}

// parseEnvFile reads KEY=VALUE lines, ignoring comments and blanks.
func parseEnvFile(src string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"'`)
		if k != "" && v != "" {
			out[k] = v
		}
	}
	return out
}

// resolvePostieeCredentials fills missing client id/secret from a credentials
// file. Precedence: explicit flag > environment > file. The returned bool says
// whether the file supplied anything, so the caller can report the path without
// printing any secret.
func resolvePostieeCredentials(clientID, clientSecret string) (string, string, string) {
	if clientID != "" && clientSecret != "" {
		return clientID, clientSecret, ""
	}
	if clientID == "" {
		clientID = os.Getenv(postieeClientIDEnv)
	}
	if clientSecret == "" {
		clientSecret = os.Getenv(postieeClientSecretEnv)
	}
	if clientID != "" && clientSecret != "" {
		return clientID, clientSecret, ""
	}

	for _, path := range postieeCredentialFiles() {
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var kv map[string]string
		if strings.HasSuffix(path, ".env") {
			kv = parseEnvFile(string(b))
		} else {
			kv = parseTOMLScalars(string(b))
		}
		if clientID == "" {
			clientID = firstNonEmpty(kv["client_id"], kv["LINKEDIN_CLIENT_ID"])
		}
		if clientSecret == "" {
			clientSecret = firstNonEmpty(kv["client_secret"], kv["LINKEDIN_CLIENT_SECRET"])
		}
		if clientID != "" || clientSecret != "" {
			return clientID, clientSecret, path
		}
	}
	return clientID, clientSecret, ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// postieeLoginCmd wraps the generated login with file-backed credentials.

// postieeLoginCmd wraps the generated login with file-backed credentials.
func newPostieeLoginCmd(flags *rootFlags) *cobra.Command {
	var port int
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with LinkedIn, reading credentials from file",
		Long: "Opens your browser to approve access, then stores the token.\n\n" +
			"Credentials are read from, in order: --client-id/--client-secret flags,\n" +
			"POSTIEE_CLIENT_ID and POSTIEE_CLIENT_SECRET, then\n" +
			"~/.config/postiee/credentials.toml. Prefer the file over the secret\n" +
			"flag so the secret stays out of your shell history.",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, secret, path := resolvePostieeCredentials("", "")
			if id == "" || secret == "" {
				return fmt.Errorf(
					"LinkedIn credentials not found.\n\n" +
						"Write them to ~/.config/postiee/credentials.toml:\n\n" +
						"    client_id = \"...\"\n" +
						"    client_secret = \"...\"\n\n" +
						"Get them from linkedin.com/developers/apps -> your app -> Auth tab.\n" +
						"Your app must also list this redirect URL under Auth:\n\n" +
						"    http://localhost:%d/callback\n", port)
			}
			if path != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Using credentials from %s\n", path)
			}
			return runOAuthLogin(cmd, flags, id, secret, port, false, false)
		},
	}
	cmd.Flags().IntVar(&port, "port", 8085, "Local callback server port (must match the app's registered redirect URL)")
	return cmd
}

func init() {
	novelCommandHooks = append(novelCommandHooks, func(root *cobra.Command, flags *rootFlags) {
		// Swap the generated auth login for the file-backed one, in place, so
		// `auth login` keeps working and nothing shadows it.
		var authCmd *cobra.Command
		for _, c := range root.Commands() {
			if c.Name() == "auth" {
				authCmd = c
				break
			}
		}
		if authCmd == nil {
			return
		}
		for _, c := range authCmd.Commands() {
			if c.Name() == "login" {
				authCmd.RemoveCommand(c)
			}
		}
		authCmd.AddCommand(newPostieeLoginCmd(flags))
	})
}