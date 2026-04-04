package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/handiism/infinita/internal/application/port/input"
	"github.com/handiism/infinita/internal/application/validation"
	domainerror "github.com/handiism/infinita/internal/domain/error"
	"github.com/handiism/infinita/internal/domain/valueobject"
	transportclient "github.com/handiism/infinita/internal/transport/client"
)

// helpTemplate is a custom help template that ensures the help command
// always appears last in the available commands list.
const helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if and .IsAvailableCommand (ne .Name "help") (ne .Name "completion")}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{range .Commands}}{{if and .IsAvailableCommand (eq .Name "completion")}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

// App is the CLI application.
type App struct {
	txnUseCase      input.TransactionUseCase
	categoryUseCase input.CategoryUseCase
	budgetUseCase   input.BudgetUseCase
	reportUseCase   input.ReportUseCase
	configPath      string
	stdout          io.Writer
	stderr          io.Writer
}

// NewApp creates a new CLI application.
func NewApp(
	txnUseCase input.TransactionUseCase,
	categoryUseCase input.CategoryUseCase,
	budgetUseCase input.BudgetUseCase,
	reportUseCase input.ReportUseCase,
	configPath string,
	stdout io.Writer,
	stderr io.Writer,
) *App {
	return &App{
		txnUseCase:      txnUseCase,
		categoryUseCase: categoryUseCase,
		budgetUseCase:   budgetUseCase,
		reportUseCase:   reportUseCase,
		configPath:      configPath,
		stdout:          stdout,
		stderr:          stderr,
	}
}

// Command returns the root cobra command.
func (a *App) Command() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "infinita",
		Short:         "Personal finance CLI",
		Long:          "infinita - Personal finance CLI",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// If --config was provided on the root command, use it to override settings file
			if configFlag := cmd.Root().Flags().Lookup("config"); configFlag != nil {
				if configPath := configFlag.Value.String(); configPath != "" {
					a.configPath = configPath
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}

	rootCmd.Flags().String("config", "", "Path to settings YAML file")
	rootCmd.PersistentFlags().String("mode", "", "Storage mode (local or remote)")
	rootCmd.PersistentFlags().String("server-url", "", "Server URL for remote mode")
	rootCmd.PersistentFlags().String("api-key", "", "API key for remote mode")

	rootCmd.AddCommand(
		a.addCommand(),
		a.listCommand(),
		a.categoryCommand(),
		a.budgetCommand(),
		a.reportCommand(),
	)

	// Set custom help template to ensure help appears last
	rootCmd.SetHelpTemplate(helpTemplate)
	// Disable the default help command, users should use --help flag instead
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	return rootCmd
}

func (a *App) addCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a transaction",
		Long:  "Add a transaction",
		RunE: func(cmd *cobra.Command, args []string) error {
			amountText, err := requiredString(cmd, "amount")
			if err != nil {
				return cliExitError(err, 2)
			}
			amount, err := validation.ParseAmount(amountText, false)
			if err != nil {
				return cliExitError(err, 2)
			}
			entryType, err := requiredString(cmd, "type")
			if err != nil {
				return cliExitError(err, 2)
			}
			category, err := requiredString(cmd, "category")
			if err != nil {
				return cliExitError(err, 2)
			}
			date, err := requiredString(cmd, "date")
			if err != nil {
				return cliExitError(err, 2)
			}
			description, _ := cmd.Flags().GetString("description")
			if _, err := a.txnUseCase.AddTransaction(cmd.Context(), entryType, amount, category, date, description); err != nil {
				return cliExitError(err, exitCode(err))
			}
			_, _ = fmt.Fprintln(a.stdout, "Transaction recorded.")
			return nil
		},
	}

	cmd.Flags().String("type", "", "Transaction type (income or expense)")
	cmd.Flags().String("amount", "", "Transaction amount")
	cmd.Flags().String("category", "", "Category name")
	cmd.Flags().String("date", "", "Transaction date (YYYY-MM-DD)")
	cmd.Flags().String("description", "", "Transaction description")

	return cmd
}

func (a *App) listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transactions",
		Long:  "List transactions",
		RunE: func(cmd *cobra.Command, args []string) error {
			category, _ := cmd.Flags().GetString("category")
			limit, _ := cmd.Flags().GetInt("limit")
			offset, _ := cmd.Flags().GetInt("offset")

			result, err := a.txnUseCase.ListTransactions(cmd.Context(), optionalString(category), limit, offset)
			if err != nil {
				return cliExitError(err, exitCode(err))
			}
			_, _ = fmt.Fprintln(a.stdout, "ID       Date       Type     Category  Amount    Description")
			for _, txn := range result.Transactions {
				shortID := txn.ID
				if len(shortID) > 8 {
					shortID = shortID[:8]
				}
				_, _ = fmt.Fprintf(a.stdout, "%-8s %-10s %-8s %-9s %-9d %s\n", shortID, txn.Date, txn.Type, txn.CategoryNameSnapshot, txn.AmountMinor, txn.Description)
			}
			return nil
		},
	}

	cmd.Flags().String("category", "", "Filter by category")
	cmd.Flags().Int("limit", valueobject.DefaultTransactionLimit, "Maximum number of transactions to list")
	cmd.Flags().Int("offset", 0, "Number of transactions to skip")

	return cmd
}

func (a *App) categoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "category",
		Short: "Manage categories",
		Long:  "Manage categories",
	}
	cmd.SetHelpTemplate(helpTemplate)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List categories",
		Long:  "List categories",
		RunE: func(cmd *cobra.Command, args []string) error {
			categories, err := a.categoryUseCase.List(cmd.Context())
			if err != nil {
				return cliExitError(err, exitCode(err))
			}
			for _, category := range categories {
				_, _ = fmt.Fprintf(a.stdout, "%s - %s\n", category.Name, category.Description)
			}
			return nil
		},
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a category",
		Long:  "Create a category",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := requiredString(cmd, "name")
			if err != nil {
				return cliExitError(err, 2)
			}
			description, _ := cmd.Flags().GetString("description")
			if _, err := a.categoryUseCase.Create(cmd.Context(), name, description); err != nil {
				return cliExitError(err, exitCode(err))
			}
			_, _ = fmt.Fprintln(a.stdout, "Category saved.")
			return nil
		},
	}
	createCmd.Flags().String("name", "", "Category name")
	createCmd.Flags().String("description", "", "Category description")

	cmd.AddCommand(listCmd, createCmd)

	return cmd
}

func (a *App) budgetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "budget",
		Short: "Manage budgets",
		Long:  "Manage budgets",
	}
	cmd.SetHelpTemplate(helpTemplate)

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set monthly budget",
		Long:  "Set monthly budget",
		RunE: func(cmd *cobra.Command, args []string) error {
			category, err := requiredString(cmd, "category")
			if err != nil {
				return cliExitError(err, 2)
			}
			month, err := requiredString(cmd, "month")
			if err != nil {
				return cliExitError(err, 2)
			}
			amountText, err := requiredString(cmd, "amount")
			if err != nil {
				return cliExitError(err, 2)
			}
			amount, err := validation.ParseAmount(amountText, false)
			if err != nil {
				return cliExitError(err, 2)
			}
			if err := a.budgetUseCase.SetBudget(cmd.Context(), category, month, amount); err != nil {
				return cliExitError(err, exitCode(err))
			}
			_, _ = fmt.Fprintln(a.stdout, "Budget stored.")
			return nil
		},
	}
	setCmd.Flags().String("category", "", "Category name")
	setCmd.Flags().String("amount", "", "Budget amount")
	setCmd.Flags().String("month", "", "Month (YYYY-MM)")

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show monthly budget status",
		Long:  "Show monthly budget status",
		RunE: func(cmd *cobra.Command, args []string) error {
			month, err := requiredString(cmd, "month")
			if err != nil {
				return cliExitError(err, 2)
			}
			statuses, err := a.budgetUseCase.Status(cmd.Context(), month)
			if err != nil {
				return cliExitError(err, exitCode(err))
			}
			for _, status := range statuses {
				_, _ = fmt.Fprintf(a.stdout, "%s: limit=%d, spent=%d, remaining=%d, over_limit=%t\n", status.CategoryName, status.MonthlyLimitMinor, status.SpentMonthToDateMinor, status.RemainingMinor, status.IsOverLimit)
			}
			return nil
		},
	}
	statusCmd.Flags().String("month", "", "Month (YYYY-MM)")

	cmd.AddCommand(setCmd, statusCmd)

	return cmd
}

func (a *App) reportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Render reports",
		Long:  "Render reports",
	}
	cmd.SetHelpTemplate(helpTemplate)

	dailyCmd := &cobra.Command{
		Use:   "daily",
		Short: "Show daily report",
		Long:  "Show daily report",
		RunE: func(cmd *cobra.Command, args []string) error {
			date, err := requiredString(cmd, "date")
			if err != nil {
				return cliExitError(err, 2)
			}
			summary, err := a.reportUseCase.Daily(cmd.Context(), date)
			if err != nil {
				return cliExitError(err, exitCode(err))
			}
			_, _ = fmt.Fprintf(a.stdout, "Daily report %s: income=%d expense=%d net=%d\n", summary.Period, summary.IncomeTotalMinor, summary.ExpenseTotalMinor, summary.NetBalanceMinor)
			return nil
		},
	}
	dailyCmd.Flags().String("date", "", "Date (YYYY-MM-DD)")

	monthlyCmd := &cobra.Command{
		Use:   "monthly",
		Short: "Show monthly report",
		Long:  "Show monthly report",
		RunE: func(cmd *cobra.Command, args []string) error {
			month, err := requiredString(cmd, "month")
			if err != nil {
				return cliExitError(err, 2)
			}
			summary, err := a.reportUseCase.Monthly(cmd.Context(), month)
			if err != nil {
				return cliExitError(err, exitCode(err))
			}
			_, _ = fmt.Fprintf(a.stdout, "Monthly report %s: income=%d expense=%d net=%d closing=%d\n", summary.Period, summary.IncomeTotalMinor, summary.ExpenseTotalMinor, summary.NetBalanceMinor, summary.ClosingBalanceMinor)
			for _, category := range summary.TopCategories {
				_, _ = fmt.Fprintf(a.stdout, " top: %s=%d\n", category.Category, category.AmountMinor)
			}
			return nil
		},
	}
	monthlyCmd.Flags().String("month", "", "Month (YYYY-MM)")

	cmd.AddCommand(dailyCmd, monthlyCmd)

	return cmd
}

func requiredString(cmd *cobra.Command, name string) (string, error) {
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", domainerror.ErrInvalidFlag.WithField(name).WithHint("flag is required")
	}
	return value, nil
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func exitCode(err error) int {
	var clientErr *transportclient.ClientError
	if errors.As(err, &clientErr) {
		return clientErr.ExitCode()
	}
	var domainErr domainerror.DomainError
	if errors.As(err, &domainErr) {
		return 2
	}
	return 3
}

func cliExitError(err error, code int) error {
	if err == nil {
		return nil
	}
	return domainerror.ExitError{
		Err:  formatCLIError(err),
		Code: code,
	}
}

func formatCLIError(err error) error {
	var clientErr *transportclient.ClientError
	if errors.As(err, &clientErr) {
		domainErrors := clientErr.ToDomainErrors()
		if len(domainErrors) > 1 {
			messages := make([]string, len(domainErrors))
			for i, domainErr := range domainErrors {
				messages[i] = domainErr.Error()
			}
			return errors.New(strings.Join(messages, "\n"))
		}
	}

	return err
}
