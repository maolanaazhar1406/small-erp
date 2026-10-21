package handler

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/model"
	"github.com/user/small-erp/internal/repository"
)

type ReportHandler struct {
	repo *repository.ReportRepo
}

func NewReportHandler(db *pgxpool.Pool) *ReportHandler {
	return &ReportHandler{repo: repository.NewReportRepo(db)}
}

func getDateRange(r *http.Request) (string, string) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	if startDate == "" {
		startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}
	return startDate, endDate
}

func (h *ReportHandler) SalesReport(w http.ResponseWriter, r *http.Request) {
	startDate, endDate := getDateRange(r)
	report, summary, err := h.repo.SalesReport(r.Context(), startDate, endDate)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to generate sales report")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Sales report generated", map[string]interface{}{
		"summary": summary,
		"details": report,
	})
}

func (h *ReportHandler) PurchaseReport(w http.ResponseWriter, r *http.Request) {
	startDate, endDate := getDateRange(r)
	report, summary, err := h.repo.PurchaseReport(r.Context(), startDate, endDate)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to generate purchase report")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Purchase report generated", map[string]interface{}{
		"summary": summary,
		"details": report,
	})
}

func (h *ReportHandler) InventoryReport(w http.ResponseWriter, r *http.Request) {
	report, summary, err := h.repo.InventoryReport(r.Context())
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to generate inventory report")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Inventory report generated", map[string]interface{}{
		"summary": summary,
		"details": report,
	})
}

func (h *ReportHandler) StockMovementReport(w http.ResponseWriter, r *http.Request) {
	startDate, endDate := getDateRange(r)
	report, err := h.repo.StockMovementReport(r.Context(), startDate, endDate)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to generate stock movement report")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Stock movement report generated", report)
}

func (h *ReportHandler) APReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.repo.APReport(r.Context())
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to generate AP report")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Accounts payable report generated", report)
}

func (h *ReportHandler) ARReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.repo.ARReport(r.Context())
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to generate AR report")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Accounts receivable report generated", report)
}

func (h *ReportHandler) CashFlowReport(w http.ResponseWriter, r *http.Request) {
	startDate, endDate := getDateRange(r)
	report, err := h.repo.CashFlowReport(r.Context(), startDate, endDate)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to generate cash flow report")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Cash flow report generated", report)
}

func (h *ReportHandler) ProfitLossReport(w http.ResponseWriter, r *http.Request) {
	startDate, endDate := getDateRange(r)
	report, err := h.repo.ProfitLossReport(r.Context(), startDate, endDate)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to generate P&L report")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Profit & Loss report generated", report)
}

func (h *ReportHandler) BalanceSheetReport(w http.ResponseWriter, r *http.Request) {
	asOfDate := r.URL.Query().Get("as_of_date")
	if asOfDate == "" {
		asOfDate = time.Now().Format("2006-01-02")
	}
	report, err := h.repo.BalanceSheetReport(r.Context(), asOfDate)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to generate balance sheet")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Balance sheet generated", report)
}
