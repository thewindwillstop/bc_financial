package handler

import (
	"bc-reconciliation-backend/internal/service"
	"bc-reconciliation-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// DashboardHandler 仪表板处理器
type DashboardHandler struct {
	txService *service.TransactionService
}

// NewDashboardHandler 创建仪表板处理器
func NewDashboardHandler(txService *service.TransactionService) *DashboardHandler {
	return &DashboardHandler{
		txService: txService,
	}
}

// GetOverview 获取概览数据
// @Summary 获取概览数据
// @Description 获取系统概览统计数据
// @Tags dashboard
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/v1/dashboard/overview [get]
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	institutionID := c.Query("institution_id")

	stats, err := h.txService.GetStatistics(institutionID)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	// 构造概览数据
	overview := map[string]interface{}{
		"total_transactions": stats.TotalTransactions,
		"matched_count":      stats.MatchedCount,
		"mismatch_count":     stats.MismatchCount,
		"match_rate":         stats.MatchRate,
		"pending_count":      stats.PendingCount,
		"uploaded_count":     stats.UploadedCount,
	}

	utils.Success(c, overview)
}

// GetChartData 获取图表数据
// @Summary 获取图表数据
// @Description 获取用于前端图表展示的数据
// @Tags dashboard
// @Produce json
// @Param days query int false "天数" default(7)
// @Success 200 {object} utils.Response
// @Router /api/v1/dashboard/chart-data [get]
func (h *DashboardHandler) GetChartData(c *gin.Context) {
	// TODO: 实现图表数据查询
	// 这里需要查询每日的对账数据

	chartData := map[string]interface{}{
		"dates":      []string{"2026-01-07", "2026-01-08", "2026-01-09", "2026-01-10", "2026-01-11", "2026-01-12", "2026-01-13"},
		"total":      []int{10, 15, 20, 25, 30, 35, 40},
		"matched":    []int{8, 14, 18, 23, 28, 32, 38},
		"mismatch":   []int{2, 1, 2, 2, 2, 3, 2},
		"match_rate": []float64{80.0, 93.3, 90.0, 92.0, 93.3, 91.4, 95.0},
	}

	utils.Success(c, chartData)
}
