package handler

import (
	"strconv"

	"bc-reconciliation-backend/internal/models"
	"bc-reconciliation-backend/internal/service"
	"bc-reconciliation-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// TransactionHandler 交易处理器
type TransactionHandler struct {
	txService *service.TransactionService
}

// NewTransactionHandler 创建交易处理器
func NewTransactionHandler(txService *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		txService: txService,
	}
}

// CreateTransaction 创建交易
// @Summary 创建交易记录
// @Description 创建单笔交易记录
// @Tags transactions
// @Accept json
// @Produce json
// @Param request body models.CreateTransactionRequest true "创建交易请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/transactions [post]
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	var req models.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 从上下文获取机构ID(需要认证中间件)
	institutionID := c.GetString("institution_id")
	if institutionID == "" {
		institutionID = "INST001" // 默认机构ID,实际应从JWT中获取
	}

	result, err := h.txService.CreateTransaction(&req, institutionID)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	if !result.Success {
		utils.Fail(c, utils.CodeDuplicate, result.Message)
		return
	}

	utils.Success(c, result)
}

// UploadExcel 上传Excel文件
// @Summary 上传Excel文件
// @Description 上传Excel文件并批量创建交易
// @Tags transactions
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Excel文件"
// @Success 200 {object} utils.Response
// @Router /api/v1/transactions/excel [post]
func (h *TransactionHandler) UploadExcel(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请上传Excel文件")
		return
	}

	// 保存临时文件
	filePath := "/tmp/" + file.Filename
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.ServerError(c, "保存文件失败: "+err.Error())
		return
	}

	// 获取机构ID
	institutionID := c.GetString("institution_id")
	if institutionID == "" {
		institutionID = "INST001"
	}

	// 解析Excel并创建交易
	result, err := h.txService.ParseExcelAndCreate(filePath, institutionID)
	if err != nil {
		utils.ServerError(c, "解析Excel失败: "+err.Error())
		return
	}

	utils.Success(c, result)
}

// UploadToChain 上链
// @Summary 交易上链
// @Description 将交易上传到区块链
// @Tags transactions
// @Accept json
// @Produce json
// @Param request body models.UploadChainRequest true "上链请求"
// @Success 200 {object} utils.Response
// @Router /api/v1/transactions/upload-chain [post]
func (h *TransactionHandler) UploadToChain(c *gin.Context) {
	var req models.UploadChainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if len(req.BizIDs) == 0 {
		utils.BadRequest(c, "请提供要上链的业务流水号")
		return
	}

	// 获取合约地址(从配置或数据库读取)
	contractAddress := c.GetString("contract_address")
	if contractAddress == "" {
		// 从配置读取
		contractAddress = "" // 需要从配置中获取
	}

	// 批量上链
	result := h.txService.BatchUploadToChain(c.Request.Context(), req.BizIDs, contractAddress)

	utils.Success(c, result)
}

// GetTransaction 查询交易详情
// @Summary 查询交易详情
// @Description 根据业务流水号查询交易详情
// @Tags transactions
// @Produce json
// @Param bizId path string true "业务流水号"
// @Success 200 {object} utils.Response
// @Router /api/v1/transactions/{bizId} [get]
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	bizId := c.Param("bizId")
	if bizId == "" {
		utils.BadRequest(c, "业务流水号不能为空")
		return
	}

	tx, err := h.txService.GetTransaction(bizId)
	if err != nil {
		utils.NotFound(c, "交易不存在")
		return
	}

	utils.Success(c, tx)
}

// ListTransactions 查询交易列表
// @Summary 查询交易列表
// @Description 分页查询交易列表
// @Tags transactions
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param status query int false "状态"
// @Param institution_id query string false "机构ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/transactions [get]
func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	// 解析状态参数
	status, _ := strconv.ParseInt(c.Query("status"), 10, 8)

	// 获取机构ID
	institutionID := c.Query("institution_id")
	if institutionID == "" {
		institutionID = c.GetString("institution_id")
	}

	// 查询列表
	result, err := h.txService.ListTransactions(institutionID, page, size, int8(status))
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.PageSuccess(c, result.Total, result.Page, result.Size, result.Data)
}

// GetStatistics 获取统计数据
// @Summary 获取统计数据
// @Description 获取交易统计数据
// @Tags dashboard
// @Produce json
// @Param institution_id query string false "机构ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/dashboard/statistics [get]
func (h *TransactionHandler) GetStatistics(c *gin.Context) {
	institutionID := c.Query("institution_id")
	if institutionID == "" {
		institutionID = c.GetString("institution_id")
	}

	stats, err := h.txService.GetStatistics(institutionID)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, stats)
}

// DownloadExcelTemplate 下载Excel模板
// @Summary 下载Excel模板
// @Description 下载交易Excel模板
// @Tags transactions
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Router /api/v1/transactions/template [get]
func (h *TransactionHandler) DownloadExcelTemplate(c *gin.Context) {
	templatePath := "/tmp/transaction_template.xlsx"

	// 生成模板
	if err := utils.CreateExcelTemplate(templatePath); err != nil {
		utils.ServerError(c, "生成模板失败: "+err.Error())
		return
	}

	// 返回文件
	c.FileAttachment(templatePath, "交易导入模板.xlsx")
}
