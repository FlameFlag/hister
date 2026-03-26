package completion

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNushellPositionalArgs(t *testing.T) {
	tests := []struct {
		use  string
		want []string
	}{
		{"cmd", nil},
		{"cmd URL [URL...]", []string{"...url: string"}},
		{"cmd [search terms]", []string{"...terms: string"}},
		{"cmd BROWSER_TYPE DB_PATH", []string{"browser_type: string", "db_path: string"}},
		{"cmd [FILENAME]", []string{"filename?: string"}},
		{"cmd REQUIRED [OPTIONAL]", []string{"required: string", "optional?: string"}},
		{"cmd ARG [ARG...]", []string{"...arg: string"}},
	}
	for _, tt := range tests {
		t.Run(tt.use, func(t *testing.T) {
			cmd := &cobra.Command{Use: tt.use}
			got := nushellPositionalArgs(cmd)
			if len(got) != len(tt.want) {
				t.Fatalf("nushellPositionalArgs(%q) = %v, want %v", tt.use, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("nushellPositionalArgs(%q)[%d] = %q, want %q", tt.use, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestNushellCmdName(t *testing.T) {
	root := &cobra.Command{Use: "app"}
	sub := &cobra.Command{Use: "sub"}
	root.AddCommand(sub)

	if got := nushellCmdName(root); got != "app" {
		t.Fatalf("nushellCmdName(root) = %q, want %q", got, "app")
	}
	if got := nushellCmdName(sub); got != "app sub" {
		t.Fatalf("nushellCmdName(sub) = %q, want %q", got, "app sub")
	}
}

func TestNushellFlagDecl(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("output", "o", "", "output file")
	cmd.Flags().Bool("verbose", false, "enable verbose output")
	cmd.Flags().Int("count", 0, "item count")

	tests := []struct {
		flag string
		want string
	}{
		{"output", `--output(-o): string  # output file`},
		{"verbose", `--verbose  # enable verbose output`},
		{"count", `--count: int  # item count`},
	}
	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			f := cmd.Flags().Lookup(tt.flag)
			got := nushellFlagDecl(cmd, f)
			if got != tt.want {
				t.Fatalf("nushellFlagDecl(%q) = %q, want %q", tt.flag, got, tt.want)
			}
		})
	}
}

func TestNushellFlagDeclWithCompletion(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("format", "", "output format")
	_ = cmd.RegisterFlagCompletionFunc("format", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "csv"}, cobra.ShellCompDirectiveNoFileComp
	})

	f := cmd.Flags().Lookup("format")
	got := nushellFlagDecl(cmd, f)
	want := `--format: string@"nu-complete test format"  # output format`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGenNushell(t *testing.T) {
	root := &cobra.Command{Use: "myapp", Short: "My application", Run: func(*cobra.Command, []string) {}}
	sub := &cobra.Command{Use: "serve", Short: "Start server", Run: func(*cobra.Command, []string) {}}
	sub.Flags().StringP("port", "p", "8080", "listen port")
	root.AddCommand(sub)
	root.InitDefaultHelpCmd()
	root.PersistentFlags().String("config", "", "config file")

	var buf bytes.Buffer
	GenNushell(root, &buf)
	out := buf.String()

	// Should contain the header
	if !strings.Contains(out, "# Nushell completions for myapp") {
		t.Fatal("missing header")
	}
	// Should contain dynamic completer using cobra's __complete
	if !strings.Contains(out, "nu-complete myapp dynamic") {
		t.Fatal("missing dynamic completer")
	}
	if !strings.Contains(out, "^myapp __complete") {
		t.Fatal("missing __complete invocation in dynamic completer")
	}
	// Should contain subcommand completer
	if !strings.Contains(out, "nu-complete myapp subcommands") {
		t.Fatal("missing subcommand completer")
	}
	// Should contain extern for root
	if !strings.Contains(out, `export extern "myapp"`) {
		t.Fatal("missing root extern")
	}
	// Should contain extern for subcommand
	if !strings.Contains(out, `export extern "myapp serve"`) {
		t.Fatal("missing subcommand extern")
	}
	// Should contain the port flag on the serve command
	if !strings.Contains(out, "--port(-p)") {
		t.Fatal("missing --port flag")
	}
	// Should contain inherited config flag on subcommand
	if !strings.Contains(out, "--config") {
		t.Fatal("missing inherited --config flag")
	}
}

func TestNushellHelp(t *testing.T) {
	help := NushellHelp("myapp")
	if !strings.Contains(help, "myapp completion nushell") {
		t.Fatal("help text should use the provided binary name")
	}
	if !strings.Contains(help, "myapp-completions.nu") {
		t.Fatal("help text should use the provided binary name for the output file")
	}
	if strings.Contains(help, "hister") {
		t.Fatal("help text should not contain hardcoded 'hister'")
	}
}

func TestIsVisibleCommand(t *testing.T) {
	root := &cobra.Command{Use: "app", Run: func(*cobra.Command, []string) {}}
	visible := &cobra.Command{Use: "serve", Short: "Start server", Run: func(*cobra.Command, []string) {}}
	hidden := &cobra.Command{Use: "internal", Hidden: true, Run: func(*cobra.Command, []string) {}}
	root.AddCommand(visible, hidden)

	// Initialize the command tree (cobra needs this for IsAvailableCommand)
	root.InitDefaultHelpCmd()

	if !isVisibleCommand(visible) {
		t.Fatal("expected 'serve' to be visible")
	}
	if isVisibleCommand(hidden) {
		t.Fatal("expected hidden command to not be visible")
	}
}
