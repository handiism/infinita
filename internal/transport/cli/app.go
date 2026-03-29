package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	ucli "github.com/urfave/cli/v3"

	"github.com/handiism/infinita/internal/application/port/input"
	"github.com/handiism/infinita/internal/application/validation"
	domainerror "github.com/handiism/infinita/internal/domain/error"
	transportclient "github.com/handiism/infinita/internal/transport/client"
)

type App struct {
	txnUseCase      input.TransactionUseCase
	categoryUseCase input.CategoryUseCase
	budgetUseCase   input.BudgetUseCase
	reportUseCase   input.ReportUseCase
	settingsUseCase input.SettingsUseCase
	stdout          io.Writer
	stderr          io.Writer
}

func NewApp(
	txnUseCase input.TransactionUseCase,
	categoryUseCase input.CategoryUseCase,
	budgetUseCase input.BudgetUseCase,
	reportUseCase input.ReportUseCase,
	settingsUseCase input.SettingsUseCase,
	stdout io.Writer,
	stderr io.Writer,
) *App {
	return &App{
		txnUseCase:      txnUseCase,
		categoryUseCase: categoryUseCase,
		budgetUseCase:   budgetUseCase,
		reportUseCase:   reportUseCase,
		settingsUseCase: settingsUseCase,
		stdout:          stdout,
		stderr:          stderr,
	}
}

func (a *App) Command() *ucli.Command {
	cmd := &ucli.Command{
		Name:      "infinita",
		Usage:     "Personal finance CLI",
		UsageText: "infinita <command> [command options]",
		Writer:    a.stdout,
		ErrWriter: a.stderr,
		Commands: []*ucli.Command{
			a.addCommand(),
			a.listCommand(),
			a.categoryCommand(),
			a.budgetCommand(),
			a.reportCommand(),
			a.settingsCommand(),
		},
		Action: func(ctx context.Context, cmd *ucli.Command) error {
			_ = ucli.ShowRootCommandHelp(cmd)
			return cliExitError(domainerror.ErrMissingCommand.WithHint("select one of the available commands"), 2)
		},
		OnUsageError: func(ctx context.Context, cmd *ucli.Command, err error, _ bool) error {
			return cliExitError(err, 2)
		},
		CommandNotFound: func(ctx context.Context, cmd *ucli.Command, command string) {
			_, _ = fmt.Fprintln(cmd.ErrWriter, domainerror.ErrUnknownCommand.WithField("command").WithHint(fmt.Sprintf("'%s' is not a known command", command)).Error())
			ucli.ShowRootCommandHelpAndExit(cmd, 2)
		},
	}

	return cmd
}

func (a *App) addCommand() *ucli.Command {
	return &ucli.Command{
		Name:      "add",
		Usage:     "Add a transaction",
		UsageText: "infinita add --type <income|expense> --amount <decimal> --category <name> --date <YYYY-MM-DD> [--description <text>]",
		Flags: []ucli.Flag{
			&ucli.StringFlag{Name: "type"},
			&ucli.StringFlag{Name: "amount"},
			&ucli.StringFlag{Name: "category"},
			&ucli.StringFlag{Name: "date"},
			&ucli.StringFlag{Name: "description"},
		},
		Action: func(ctx context.Context, cmd *ucli.Command) error {
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
			if _, err := a.txnUseCase.AddTransaction(ctx, entryType, amount, category, date, cmd.String("description")); err != nil {
				return cliExitError(err, exitCode(err))
			}
			_, _ = fmt.Fprintln(cmd.Writer, "Transaction recorded.")
			return nil
		},
	}
}

func (a *App) listCommand() *ucli.Command {
	return &ucli.Command{
		Name:      "list",
		Usage:     "List transactions",
		UsageText: "infinita list [--category <name>] [--limit <n>] [--offset <n>]",
		Flags: []ucli.Flag{
			&ucli.StringFlag{Name: "category"},
			&ucli.IntFlag{Name: "limit", Value: 50},
			&ucli.IntFlag{Name: "offset", Value: 0},
		},
		Action: func(ctx context.Context, cmd *ucli.Command) error {
			result, err := a.txnUseCase.ListTransactions(ctx, optionalString(cmd.String("category")), cmd.Int("limit"), cmd.Int("offset"))
			if err != nil {
				return cliExitError(err, exitCode(err))
			}
			_, _ = fmt.Fprintln(cmd.Writer, "ID       Date       Type     Category  Amount    Description")
			for _, txn := range result.Transactions {
				shortID := txn.ID
				if len(shortID) > 8 {
					shortID = shortID[:8]
				}
				_, _ = fmt.Fprintf(cmd.Writer, "%-8s %-10s %-8s %-9s %-9d %s\n", shortID, txn.Date, txn.Type, txn.CategoryNameSnapshot, txn.AmountMinor, txn.Description)
			}
			return nil
		},
	}
}

func (a *App) categoryCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "category",
		Usage: "Manage categories",
		Commands: []*ucli.Command{
			{
				Name:  "list",
				Usage: "List categories",
				Action: func(ctx context.Context, cmd *ucli.Command) error {
					categories, err := a.categoryUseCase.List(ctx)
					if err != nil {
						return cliExitError(err, exitCode(err))
					}
					for _, category := range categories {
						_, _ = fmt.Fprintf(cmd.Writer, "%s - %s\n", category.Name, category.Description)
					}
					return nil
				},
			},
			{
				Name:      "create",
				Usage:     "Create a category",
				UsageText: "infinita category create --name <name> [--description <text>]",
				Flags: []ucli.Flag{
					&ucli.StringFlag{Name: "name"},
					&ucli.StringFlag{Name: "description"},
				},
				Action: func(ctx context.Context, cmd *ucli.Command) error {
					name, err := requiredString(cmd, "name")
					if err != nil {
						return cliExitError(err, 2)
					}
					if _, err := a.categoryUseCase.Create(ctx, name, cmd.String("description")); err != nil {
						return cliExitError(err, exitCode(err))
					}
					_, _ = fmt.Fprintln(cmd.Writer, "Category saved.")
					return nil
				},
			},
		},
	}
}

func (a *App) budgetCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "budget",
		Usage: "Manage budgets",
		Commands: []*ucli.Command{
			{
				Name:      "set",
				Usage:     "Set monthly budget",
				UsageText: "infinita budget set --category <name> --amount <decimal> --month <YYYY-MM>",
				Flags: []ucli.Flag{
					&ucli.StringFlag{Name: "category"},
					&ucli.StringFlag{Name: "amount"},
					&ucli.StringFlag{Name: "month"},
				},
				Action: func(ctx context.Context, cmd *ucli.Command) error {
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
					if err := a.budgetUseCase.SetBudget(ctx, category, month, amount); err != nil {
						return cliExitError(err, exitCode(err))
					}
					_, _ = fmt.Fprintln(cmd.Writer, "Budget stored.")
					return nil
				},
			},
			{
				Name:      "status",
				Usage:     "Show monthly budget status",
				UsageText: "infinita budget status --month <YYYY-MM>",
				Flags: []ucli.Flag{
					&ucli.StringFlag{Name: "month"},
				},
				Action: func(ctx context.Context, cmd *ucli.Command) error {
					month, err := requiredString(cmd, "month")
					if err != nil {
						return cliExitError(err, 2)
					}
					statuses, err := a.budgetUseCase.Status(ctx, month)
					if err != nil {
						return cliExitError(err, exitCode(err))
					}
					for _, status := range statuses {
						_, _ = fmt.Fprintf(cmd.Writer, "%s: limit=%d, spent=%d, remaining=%d, over_limit=%t\n", status.CategoryName, status.MonthlyLimitMinor, status.SpentMonthToDateMinor, status.RemainingMinor, status.IsOverLimit)
					}
					return nil
				},
			},
		},
	}
}

func (a *App) reportCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "report",
		Usage: "Render reports",
		Commands: []*ucli.Command{
			{
				Name:      "daily",
				Usage:     "Show daily report",
				UsageText: "infinita report daily --date <YYYY-MM-DD>",
				Flags: []ucli.Flag{
					&ucli.StringFlag{Name: "date"},
				},
				Action: func(ctx context.Context, cmd *ucli.Command) error {
					date, err := requiredString(cmd, "date")
					if err != nil {
						return cliExitError(err, 2)
					}
					summary, err := a.reportUseCase.Daily(ctx, date)
					if err != nil {
						return cliExitError(err, exitCode(err))
					}
					_, _ = fmt.Fprintf(cmd.Writer, "Daily report %s: income=%d expense=%d net=%d\n", summary.Period, summary.IncomeTotalMinor, summary.ExpenseTotalMinor, summary.NetBalanceMinor)
					return nil
				},
			},
			{
				Name:      "monthly",
				Usage:     "Show monthly report",
				UsageText: "infinita report monthly --month <YYYY-MM>",
				Flags: []ucli.Flag{
					&ucli.StringFlag{Name: "month"},
				},
				Action: func(ctx context.Context, cmd *ucli.Command) error {
					month, err := requiredString(cmd, "month")
					if err != nil {
						return cliExitError(err, 2)
					}
					summary, err := a.reportUseCase.Monthly(ctx, month)
					if err != nil {
						return cliExitError(err, exitCode(err))
					}
					_, _ = fmt.Fprintf(cmd.Writer, "Monthly report %s: income=%d expense=%d net=%d closing=%d\n", summary.Period, summary.IncomeTotalMinor, summary.ExpenseTotalMinor, summary.NetBalanceMinor, summary.ClosingBalanceMinor)
					for _, category := range summary.TopCategories {
						_, _ = fmt.Fprintf(cmd.Writer, " top: %s=%d\n", category.Category, category.AmountMinor)
					}
					return nil
				},
			},
		},
	}
}

func (a *App) settingsCommand() *ucli.Command {
	return &ucli.Command{
		Name:  "settings",
		Usage: "Manage settings",
		Commands: []*ucli.Command{
			{
				Name:  "show",
				Usage: "Show current settings",
				Action: func(ctx context.Context, cmd *ucli.Command) error {
					settings, err := a.settingsUseCase.Show(ctx)
					if err != nil {
						return cliExitError(err, exitCode(err))
					}
					_, _ = fmt.Fprintf(cmd.Writer, "Storage mode: %s\n", settings.StorageMode)
					_, _ = fmt.Fprintf(cmd.Writer, "Analytics opt-in: %t\n", settings.AnalyticsOptIn)
					_, _ = fmt.Fprintf(cmd.Writer, "Report timezone: %s\n", settings.ReportTimezone)
					return nil
				},
			},
			{
				Name:      "set-initial-balance",
				Usage:     "Set initial balance",
				UsageText: "infinita settings set-initial-balance --amount <decimal>",
				Flags: []ucli.Flag{
					&ucli.StringFlag{Name: "amount"},
				},
				Action: func(ctx context.Context, cmd *ucli.Command) error {
					amountText, err := requiredString(cmd, "amount")
					if err != nil {
						return cliExitError(err, 2)
					}
					amount, err := validation.ParseAmount(amountText, true)
					if err != nil {
						return cliExitError(err, 2)
					}
					if _, err := a.settingsUseCase.SetInitialBalance(ctx, amount); err != nil {
						return cliExitError(err, exitCode(err))
					}
					_, _ = fmt.Fprintln(cmd.Writer, "Initial balance updated.")
					return nil
				},
			},
			{
				Name:      "reset-initial-balance",
				Usage:     "Reset initial balance to zero",
				UsageText: "infinita settings reset-initial-balance",
				Action: func(ctx context.Context, cmd *ucli.Command) error {
					if err := a.settingsUseCase.ResetInitialBalance(ctx); err != nil {
						return cliExitError(err, exitCode(err))
					}
					_, _ = fmt.Fprintln(cmd.Writer, "Initial balance reset.")
					return nil
				},
			},
			{
				Name:      "analytics",
				Usage:     "Update analytics opt-in",
				UsageText: "infinita settings analytics --opt-in <true|false>",
				Flags: []ucli.Flag{
					&ucli.StringFlag{Name: "opt-in"},
				},
				Action: func(ctx context.Context, cmd *ucli.Command) error {
					optInText, err := requiredString(cmd, "opt-in")
					if err != nil {
						return cliExitError(err, 2)
					}
					optIn, err := strconv.ParseBool(optInText)
					if err != nil {
						return cliExitError(domainerror.ErrInvalidFlag.WithField("opt-in").WithHint("must be true or false"), 2)
					}
					if err := a.settingsUseCase.SetAnalyticsOptIn(ctx, optIn); err != nil {
						return cliExitError(err, exitCode(err))
					}
					_, _ = fmt.Fprintf(cmd.Writer, "Analytics opt-in updated: %t\n", optIn)
					return nil
				},
			},
			{
				Name:      "report-timezone",
				Usage:     "Update report timezone",
				UsageText: "infinita settings report-timezone --timezone <IANA name>",
				Flags: []ucli.Flag{
					&ucli.StringFlag{Name: "timezone"},
				},
				Action: func(ctx context.Context, cmd *ucli.Command) error {
					timezone, err := requiredString(cmd, "timezone")
					if err != nil {
						return cliExitError(err, 2)
					}
					if err := a.settingsUseCase.SetReportTimezone(ctx, timezone); err != nil {
						return cliExitError(err, exitCode(err))
					}
					_, _ = fmt.Fprintln(cmd.Writer, "Report timezone updated.")
					return nil
				},
			},
		},
	}
}

func requiredString(cmd *ucli.Command, name string) (string, error) {
	value := cmd.String(name)
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
	return ucli.Exit(formatCLIError(err), code)
}

func formatCLIError(err error) string {
	var clientErr *transportclient.ClientError
	if errors.As(err, &clientErr) {
		domainErrors := clientErr.ToDomainErrors()
		if len(domainErrors) > 1 {
			messages := make([]string, len(domainErrors))
			for i, domainErr := range domainErrors {
				messages[i] = domainErr.Error()
			}
			return strings.Join(messages, "\n")
		}
	}

	return err.Error()
}
