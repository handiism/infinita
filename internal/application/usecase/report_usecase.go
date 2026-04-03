package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/handiism/infinita/internal/application/port/output"
	"github.com/handiism/infinita/internal/application/validation"
	"github.com/handiism/infinita/internal/domain/entity"
	domainerror "github.com/handiism/infinita/internal/domain/error"
	"github.com/handiism/infinita/internal/domain/valueobject"
)

// ReportUseCase implements report generation operations.
type ReportUseCase struct {
	txRepo             output.TransactionRepository
	initialBalanceRepo output.InitialBalanceRepository
	settingsRepo       output.SettingsRepository
}

// NewReportUseCase creates a new ReportUseCase.
func NewReportUseCase(txRepo output.TransactionRepository, initialBalanceRepo output.InitialBalanceRepository, settingsRepo output.SettingsRepository) *ReportUseCase {
	return &ReportUseCase{txRepo: txRepo, initialBalanceRepo: initialBalanceRepo, settingsRepo: settingsRepo}
}

func (u *ReportUseCase) Daily(ctx context.Context, date string) (entity.DailySummary, error) {
	normalizedDate, err := validation.ParseISODate(date)
	if err != nil {
		return entity.DailySummary{}, err
	}
	if _, err := u.loadReportLocation(ctx); err != nil {
		return entity.DailySummary{}, err
	}
	income, expense, err := u.txRepo.SumTotalsForDay(ctx, normalizedDate)
	if err != nil {
		return entity.DailySummary{}, fmt.Errorf("daily totals: %w", err)
	}
	return entity.DailySummary{
		Period:            normalizedDate,
		CurrencyCode:      valueobject.DefaultCurrencyCode,
		IncomeTotalMinor:  income,
		ExpenseTotalMinor: expense,
		NetBalanceMinor:   income - expense,
	}, nil
}

func (u *ReportUseCase) Monthly(ctx context.Context, month string) (entity.MonthlySummary, error) {
	loc, err := u.loadReportLocation(ctx)
	if err != nil {
		return entity.MonthlySummary{}, err
	}
	normalizedMonth, err := validation.ParseISOMonth(month)
	if err != nil {
		return entity.MonthlySummary{}, err
	}
	parsed, err := time.ParseInLocation("2006-01", normalizedMonth, loc)
	if err != nil {
		return entity.MonthlySummary{}, domainerror.ErrInvalidMonth
	}
	monthKey := normalizedMonth
	income, expense, err := u.txRepo.SumTotalsForMonth(ctx, monthKey)
	if err != nil {
		return entity.MonthlySummary{}, fmt.Errorf("monthly totals: %w", err)
	}
	// Determine last day of the month
	endOfMonth := time.Date(parsed.Year(), parsed.Month()+1, 1, 0, 0, 0, 0, loc).Add(-time.Nanosecond)
	cumulativeIncome, cumulativeExpense, err := u.txRepo.SumCumulativeTotalsUpToDate(ctx, endOfMonth.Format("2006-01-02"))
	if err != nil {
		return entity.MonthlySummary{}, fmt.Errorf("cumulative totals: %w", err)
	}
	top, err := u.txRepo.TopCategoriesForMonth(ctx, monthKey)
	if err != nil {
		return entity.MonthlySummary{}, fmt.Errorf("top categories: %w", err)
	}
	initialBalance, err := u.initialBalanceRepo.Get(ctx)
	if err != nil {
		return entity.MonthlySummary{}, fmt.Errorf("initial balance: %w", err)
	}
	closingBalance := initialBalance.InitialBalanceMinor + cumulativeIncome - cumulativeExpense
	topCategories := make([]entity.TopSpendingCategory, 0, len(top))
	for _, t := range top {
		topCategories = append(topCategories, entity.TopSpendingCategory{Category: t.Category, AmountMinor: t.AmountMinor})
	}
	return entity.MonthlySummary{
		Period:              monthKey,
		CurrencyCode:        valueobject.DefaultCurrencyCode,
		IncomeTotalMinor:    income,
		ExpenseTotalMinor:   expense,
		NetBalanceMinor:     income - expense,
		ClosingBalanceMinor: closingBalance,
		TopCategories:       topCategories,
	}, nil
}

func (u *ReportUseCase) loadReportLocation(ctx context.Context) (*time.Location, error) {
	settings, err := u.settingsRepo.GetSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings: %w", err)
	}
	tz := strings.TrimSpace(settings.ReportTimezone)
	if tz == "" {
		return nil, domainerror.ErrInvalidTimezone.WithField("report_timezone").WithHint("report timezone must be configured")
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, domainerror.ErrInvalidTimezone.WithField("report_timezone").WithHint("provide a valid IANA timezone")
	}
	return loc, nil
}
