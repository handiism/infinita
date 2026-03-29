package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/handiism/infinita/internal/domain/entity"
	domainerror "github.com/handiism/infinita/internal/domain/error"
	ucli "github.com/urfave/cli/v3"
)

type stubTransactionUseCase struct {
	addFn  func(context.Context, string, int64, string, string, string) error
	listFn func(context.Context, *string, int, int) ([]entity.Transaction, error)
}

func (s stubTransactionUseCase) AddTransaction(ctx context.Context, entryType string, amountMinor int64, category string, date string, description string) error {
	if s.addFn != nil {
		return s.addFn(ctx, entryType, amountMinor, category, date, description)
	}
	return nil
}

func (s stubTransactionUseCase) ListTransactions(ctx context.Context, category *string, limit, offset int) ([]entity.Transaction, error) {
	if s.listFn != nil {
		return s.listFn(ctx, category, limit, offset)
	}
	return nil, nil
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
	showFn                func(context.Context) (entity.Settings, error)
	setInitialBalanceFn   func(context.Context, int64) error
	resetInitialBalanceFn func(context.Context) error
	setAnalyticsOptInFn   func(context.Context, bool) error
	setReportTimezoneFn   func(context.Context, string) error
}

func (s stubSettingsUseCase) Show(ctx context.Context) (entity.Settings, error) {
	if s.showFn != nil {
		return s.showFn(ctx)
	}
	return entity.Settings{}, nil
}

func (s stubSettingsUseCase) SetInitialBalance(ctx context.Context, amount int64) error {
	if s.setInitialBalanceFn != nil {
		return s.setInitialBalanceFn(ctx, amount)
	}
	return nil
}

func (s stubSettingsUseCase) ResetInitialBalance(ctx context.Context) error {
	if s.resetInitialBalanceFn != nil {
		return s.resetInitialBalanceFn(ctx)
	}
	return nil
}

func (s stubSettingsUseCase) SetAnalyticsOptIn(ctx context.Context, optIn bool) error {
	if s.setAnalyticsOptInFn != nil {
		return s.setAnalyticsOptInFn(ctx, optIn)
	}
	return nil
}

func (s stubSettingsUseCase) SetReportTimezone(ctx context.Context, timezone string) error {
	if s.setReportTimezoneFn != nil {
		return s.setReportTimezoneFn(ctx, timezone)
	}
	return nil
}

func TestHelpers(t *testing.T) {
	cmd := NewApp(nil, nil, nil, nil, nil, &bytes.Buffer{}, &bytes.Buffer{}).Command()
	cmd.Root().Flags = nil

	if _, err := requiredString(cmd, "missing"); err == nil {
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
}

func TestAddCommandSuccess(t *testing.T) {
	called := false
	app, stdout, stderr := newTestApp(
		stubTransactionUseCase{addFn: func(_ context.Context, entryType string, amountMinor int64, category string, date string, description string) error {
			called = true
			if entryType != "expense" || amountMinor != 12345 || category != "Food" || date != "2024-01-15" || description != "lunch" {
				t.Fatalf("unexpected add args: %q %d %q %q %q", entryType, amountMinor, category, date, description)
			}
			return nil
		}},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
		stubSettingsUseCase{},
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
		stubTransactionUseCase{listFn: func(_ context.Context, category *string, limit, offset int) ([]entity.Transaction, error) {
			if category == nil || *category != "Food" || limit != 10 || offset != 5 {
				t.Fatalf("unexpected list args: %v %d %d", category, limit, offset)
			}
			return []entity.Transaction{{ID: "1", Date: "2024-01-15", Type: "expense", CategoryNameSnapshot: "Food", AmountMinor: 12345, Description: "lunch"}}, nil
		}},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
		stubSettingsUseCase{},
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
		stubSettingsUseCase{
			showFn: func(context.Context) (entity.Settings, error) {
				return entity.Settings{StorageMode: "local", AnalyticsOptIn: true, ReportTimezone: "Asia/Jakarta"}, nil
			},
			setInitialBalanceFn: func(_ context.Context, amount int64) error {
				if amount != 50000 {
					t.Fatalf("unexpected initial balance amount: %d", amount)
				}
				return nil
			},
			resetInitialBalanceFn: func(context.Context) error { return nil },
			setAnalyticsOptInFn: func(_ context.Context, optIn bool) error {
				if !optIn {
					t.Fatal("expected opt-in true")
				}
				return nil
			},
			setReportTimezoneFn: func(_ context.Context, timezone string) error {
				if timezone != "UTC" {
					t.Fatalf("unexpected timezone: %q", timezone)
				}
				return nil
			},
		},
	)

	commands := []struct {
		command     *ucli.Command
		args        []string
		wantContain string
	}{
		{command: app.categoryCommand().Commands[0], args: []string{"list"}, wantContain: "Food - Meals"},
		{command: app.categoryCommand().Commands[1], args: []string{"create", "--name", "Travel", "--description", "Trips"}, wantContain: "Category saved."},
		{command: app.budgetCommand().Commands[0], args: []string{"set", "--category", "Food", "--amount", "400.00", "--month", "2024-01"}, wantContain: "Budget stored."},
		{command: app.budgetCommand().Commands[1], args: []string{"status", "--month", "2024-01"}, wantContain: "Food: limit=40000, spent=12345, remaining=27655, over_limit=false"},
		{command: app.reportCommand().Commands[0], args: []string{"daily", "--date", "2024-01-15"}, wantContain: "Daily report 2024-01-15: income=50000 expense=12345 net=37655"},
		{command: app.reportCommand().Commands[1], args: []string{"monthly", "--month", "2024-01"}, wantContain: "Monthly report 2024-01: income=100000 expense=12345 net=87655 closing=120000"},
		{command: app.settingsCommand().Commands[0], args: []string{"show"}, wantContain: "Storage mode: local"},
		{command: app.settingsCommand().Commands[1], args: []string{"set-initial-balance", "--amount", "500.00"}, wantContain: "Initial balance updated."},
		{command: app.settingsCommand().Commands[2], args: []string{"reset-initial-balance"}, wantContain: "Initial balance reset."},
		{command: app.settingsCommand().Commands[3], args: []string{"analytics", "--opt-in", "true"}, wantContain: "Analytics opt-in updated: true"},
		{command: app.settingsCommand().Commands[4], args: []string{"report-timezone", "--timezone", "UTC"}, wantContain: "Report timezone updated."},
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

func newTestApp(txn stubTransactionUseCase, category stubCategoryUseCase, budget stubBudgetUseCase, report stubReportUseCase, settings stubSettingsUseCase) (*App, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	return NewApp(txn, category, budget, report, settings, stdout, stderr), stdout, stderr
}

func runTestCommand(cmd *ucli.Command, stdout, stderr *bytes.Buffer, args []string) error {
	cmd.Writer = stdout
	cmd.ErrWriter = stderr
	return cmd.Run(context.Background(), args)
}
