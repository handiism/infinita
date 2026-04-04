package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/handiism/infinita/internal/application/port/input"
	"github.com/handiism/infinita/internal/domain/entity"
	domainerror "github.com/handiism/infinita/internal/domain/error"
	transportclient "github.com/handiism/infinita/internal/transport/client"
)

type stubTransactionUseCase struct {
	addFn  func(context.Context, string, int64, string, string, string) (entity.Transaction, error)
	listFn func(context.Context, *string, int, int) (input.TransactionListResult, error)
}

func (s stubTransactionUseCase) AddTransaction(ctx context.Context, entryType string, amountMinor int64, category string, date string, description string) (entity.Transaction, error) {
	if s.addFn != nil {
		return s.addFn(ctx, entryType, amountMinor, category, date, description)
	}
	return entity.Transaction{}, nil
}

func (s stubTransactionUseCase) ListTransactions(ctx context.Context, category *string, limit, offset int) (input.TransactionListResult, error) {
	if s.listFn != nil {
		return s.listFn(ctx, category, limit, offset)
	}
	return input.TransactionListResult{}, nil
}

type stubCategoryUseCase struct {
	listFn   func(context.Context) ([]entity.Category, error)
	createFn func(context.Context, string, string) (entity.Category, error)
}

func (s stubCategoryUseCase) List(ctx context.Context) ([]entity.Category, error) {
	if s.listFn != nil {
		return s.listFn(ctx)
	}
	return nil, nil
}

func (s stubCategoryUseCase) Create(ctx context.Context, name, description string) (entity.Category, error) {
	if s.createFn != nil {
		return s.createFn(ctx, name, description)
	}
	return entity.Category{}, nil
}

type stubBudgetUseCase struct {
	setFn    func(context.Context, string, string, int64) error
	statusFn func(context.Context, string) ([]entity.BudgetStatus, error)
}

func (s stubBudgetUseCase) SetBudget(ctx context.Context, category, month string, limit int64) error {
	if s.setFn != nil {
		return s.setFn(ctx, category, month, limit)
	}
	return nil
}

func (s stubBudgetUseCase) Status(ctx context.Context, month string) ([]entity.BudgetStatus, error) {
	if s.statusFn != nil {
		return s.statusFn(ctx, month)
	}
	return nil, nil
}

type stubReportUseCase struct {
	dailyFn   func(context.Context, string) (entity.DailySummary, error)
	monthlyFn func(context.Context, string) (entity.MonthlySummary, error)
}

func (s stubReportUseCase) Daily(ctx context.Context, date string) (entity.DailySummary, error) {
	if s.dailyFn != nil {
		return s.dailyFn(ctx, date)
	}
	return entity.DailySummary{}, nil
}

func (s stubReportUseCase) Monthly(ctx context.Context, month string) (entity.MonthlySummary, error) {
	if s.monthlyFn != nil {
		return s.monthlyFn(ctx, month)
	}
	return entity.MonthlySummary{}, nil
}

type stubSettingsUseCase struct {
	setInitialBalanceFn   func(context.Context, int64) (entity.InitialBalance, error)
	resetInitialBalanceFn func(context.Context) error
}

func (s stubSettingsUseCase) SetInitialBalance(ctx context.Context, amount int64) (entity.InitialBalance, error) {
	if s.setInitialBalanceFn != nil {
		return s.setInitialBalanceFn(ctx, amount)
	}
	return entity.InitialBalance{}, nil
}

func (s stubSettingsUseCase) ResetInitialBalance(ctx context.Context) error {
	if s.resetInitialBalanceFn != nil {
		return s.resetInitialBalanceFn(ctx)
	}
	return nil
}

func TestHelpers(t *testing.T) {
	app := NewApp(nil, nil, nil, nil, "settings.yaml", &bytes.Buffer{}, &bytes.Buffer{})

	// Create a minimal subcommand for testing requiredString
	subCmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := requiredString(cmd, "missing")
			return err
		},
	}
	subCmd.Flags().String("missing", "", "")

	err := subCmd.RunE(subCmd, []string{})
	if err == nil {
		t.Fatal("requiredString() error = nil, want error")
	}

	if got := optionalString(""); got != nil {
		t.Fatalf("optionalString(empty) = %v, want nil", got)
	}
	value := optionalString("food")
	if value == nil || *value != "food" {
		t.Fatalf("optionalString() = %v, want food", value)
	}
	if exitCode(domainerror.ErrInvalidAmount) != 2 {
		t.Fatalf("exitCode(domain error) = %d, want 2", exitCode(domainerror.ErrInvalidAmount))
	}
	if exitCode(errors.New("boom")) != 3 {
		t.Fatalf("exitCode(runtime error) = %d, want 3", exitCode(errors.New("boom")))
	}
	if cliExitError(nil, 2) != nil {
		t.Fatal("cliExitError(nil) should return nil")
	}

	// Suppress unused variable warnings
	_ = app
}

func TestFormatCLIErrorExpandsMultipleClientErrors(t *testing.T) {
	err := &transportclient.ClientError{
		MultipleErrors: []domainerror.DomainError{
			domainerror.ErrInvalidTransactionType.WithField("type"),
			domainerror.ErrInvalidDate.WithField("date"),
		},
		StatusCode: 400,
	}

	formatted := formatCLIError(err).Error()

	if !strings.Contains(formatted, "INVALID_TYPE: type must be either 'income' or 'expense' (field=type)") {
		t.Fatalf("formatCLIError() missing first error: %q", formatted)
	}
	if !strings.Contains(formatted, "INVALID_DATE: date must be a valid YYYY-MM-DD value (field=date)") {
		t.Fatalf("formatCLIError() missing second error: %q", formatted)
	}
	if strings.Contains(formatted, "and 1 more errors") {
		t.Fatalf("formatCLIError() collapsed multiple errors: %q", formatted)
	}
}

func TestAddCommandSuccess(t *testing.T) {
	called := false
	app, stdout, stderr := newTestApp(
		stubTransactionUseCase{addFn: func(_ context.Context, entryType string, amountMinor int64, category string, date string, description string) (entity.Transaction, error) {
			called = true
			if entryType != "expense" || amountMinor != 12345 || category != "Food" || date != "2024-01-15" || description != "lunch" {
				t.Fatalf("unexpected add args: %q %d %q %q %q", entryType, amountMinor, category, date, description)
			}
			return entity.Transaction{}, nil
		}},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
	)

	err := runTestCommand(app.addCommand(), stdout, stderr, []string{"add", "--type", "expense", "--amount", "123.45", "--category", "Food", "--date", "2024-01-15", "--description", "lunch"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !called {
		t.Fatal("AddTransaction was not called")
	}
	if got := stdout.String(); got != "Transaction recorded.\n" {
		t.Fatalf("stdout = %q, want %q", got, "Transaction recorded.\n")
	}
}

func TestListCommandSuccess(t *testing.T) {
	app, stdout, stderr := newTestApp(
		stubTransactionUseCase{listFn: func(_ context.Context, category *string, limit, offset int) (input.TransactionListResult, error) {
			if category == nil || *category != "Food" || limit != 10 || offset != 5 {
				t.Fatalf("unexpected list args: %v %d %d", category, limit, offset)
			}
			return input.TransactionListResult{
				Transactions: []entity.Transaction{{ID: "1", Date: "2024-01-15", Type: "expense", CategoryNameSnapshot: "Food", AmountMinor: 12345, Description: "lunch"}},
			}, nil
		}},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
	)

	err := runTestCommand(app.listCommand(), stdout, stderr, []string{"list", "--category", "Food", "--limit", "10", "--offset", "5"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, "ID       Date") || !strings.Contains(got, "1        2024-01-15 expense  Food      12345     lunch") {
		t.Fatalf("unexpected stdout: %q", got)
	}
}

func TestCategoryBudgetReportAndSettingsCommands(t *testing.T) {
	app, stdout, stderr := newTestApp(
		stubTransactionUseCase{},
		stubCategoryUseCase{
			listFn: func(context.Context) ([]entity.Category, error) {
				return []entity.Category{{Name: "Food", Description: "Meals"}}, nil
			},
			createFn: func(_ context.Context, name, description string) (entity.Category, error) {
				if name != "Travel" || description != "Trips" {
					t.Fatalf("unexpected create args: %q %q", name, description)
				}
				return entity.Category{Name: name, Description: description}, nil
			},
		},
		stubBudgetUseCase{
			setFn: func(_ context.Context, category, month string, limit int64) error {
				if category != "Food" || month != "2024-01" || limit != 40000 {
					t.Fatalf("unexpected budget set args: %q %q %d", category, month, limit)
				}
				return nil
			},
			statusFn: func(_ context.Context, month string) ([]entity.BudgetStatus, error) {
				if month != "2024-01" {
					t.Fatalf("unexpected budget status month: %q", month)
				}
				return []entity.BudgetStatus{{CategoryName: "Food", MonthlyLimitMinor: 40000, SpentMonthToDateMinor: 12345, RemainingMinor: 27655, IsOverLimit: false}}, nil
			},
		},
		stubReportUseCase{
			dailyFn: func(_ context.Context, date string) (entity.DailySummary, error) {
				if date != "2024-01-15" {
					t.Fatalf("unexpected daily date: %q", date)
				}
				return entity.DailySummary{Period: date, IncomeTotalMinor: 50000, ExpenseTotalMinor: 12345, NetBalanceMinor: 37655}, nil
			},
			monthlyFn: func(_ context.Context, month string) (entity.MonthlySummary, error) {
				if month != "2024-01" {
					t.Fatalf("unexpected monthly month: %q", month)
				}
				return entity.MonthlySummary{Period: month, IncomeTotalMinor: 100000, ExpenseTotalMinor: 12345, NetBalanceMinor: 87655, ClosingBalanceMinor: 120000, TopCategories: []entity.TopSpendingCategory{{Category: "Food", AmountMinor: 12345}}}, nil
			},
		},
	)

	commands := []struct {
		command     *cobra.Command
		parentName  string
		args        []string
		wantContain string
	}{
		{command: app.categoryCommand().Commands()[0], parentName: "category", args: []string{"category", "list"}, wantContain: "Food - Meals"},
		{command: app.categoryCommand().Commands()[1], parentName: "category", args: []string{"category", "create", "--name", "Travel", "--description", "Trips"}, wantContain: "Category saved."},
		{command: app.budgetCommand().Commands()[0], parentName: "budget", args: []string{"budget", "set", "--category", "Food", "--amount", "400.00", "--month", "2024-01"}, wantContain: "Budget stored."},
		{command: app.budgetCommand().Commands()[1], parentName: "budget", args: []string{"budget", "status", "--month", "2024-01"}, wantContain: "Food: limit=40000, spent=12345, remaining=27655, over_limit=false"},
		{command: app.reportCommand().Commands()[0], parentName: "report", args: []string{"report", "daily", "--date", "2024-01-15"}, wantContain: "Daily report 2024-01-15: income=50000 expense=12345 net=37655"},
		{command: app.reportCommand().Commands()[1], parentName: "report", args: []string{"report", "monthly", "--month", "2024-01"}, wantContain: "Monthly report 2024-01: income=100000 expense=12345 net=87655 closing=120000"},
	}

	for _, tc := range commands {
		stdout.Reset()
		if err := runTestCommand(tc.command, stdout, stderr, tc.args); err != nil {
			t.Fatalf("Run(%v) error = %v", tc.args, err)
		}
		got := stdout.String()
		if !strings.Contains(got, tc.wantContain) {
			t.Fatalf("Run(%v): stdout missing %q in %q", tc.args, tc.wantContain, got)
		}
	}
}

func newTestApp(txn stubTransactionUseCase, category stubCategoryUseCase, budget stubBudgetUseCase, report stubReportUseCase) (*App, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	return NewApp(txn, category, budget, report, "settings.yaml", stdout, stderr), stdout, stderr
}

func runTestCommand(cmd *cobra.Command, stdout, stderr *bytes.Buffer, args []string) error {
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	// If the command has a parent (is a subcommand), we need to run it through the root
	// by setting up the proper command chain
	if cmd.Parent() != nil {
		// Build the command chain from root to this subcommand
		var chain []*cobra.Command
		current := cmd
		for current != nil {
			chain = append([]*cobra.Command{current}, chain...)
			current = current.Parent()
		}

		// The first command in the chain should be the root
		rootCmd := chain[0]
		rootCmd.SetOut(stdout)
		rootCmd.SetErr(stderr)

		// Extract just the subcommand args (skip the root command name)
		subcommandArgs := args
		if len(args) > 0 && args[0] == rootCmd.Name() {
			subcommandArgs = args[1:]
		}

		rootCmd.SetArgs(subcommandArgs)
		return rootCmd.Execute()
	}

	// For root-level commands, just execute with the given args
	cmd.SetArgs(args)
	return cmd.Execute()
}
