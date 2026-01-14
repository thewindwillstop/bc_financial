package service

import (
	"context"
	"fmt"

	"bc-reconciliation-backend/internal/blockchain"
	"bc-reconciliation-backend/internal/models"
	"bc-reconciliation-backend/internal/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TransactionService 交易服务
type TransactionService struct {
	db           *gorm.DB
	blockchain   *blockchain.Client
	logger       *zap.Logger
	encryptionKey string // AES加密密钥(32字节)
}

// NewTransactionService 创建交易服务
func NewTransactionService(db *gorm.DB, bc *blockchain.Client, logger *zap.Logger, encryptionKey string) *TransactionService {
	return &TransactionService{
		db:           db,
		blockchain:   bc,
		logger:       logger,
		encryptionKey: encryptionKey,
	}
}

// CreateTransactionResult 创建交易结果
type CreateTransactionResult struct {
	Success bool   `json:"success"`
	BizID   string `json:"biz_id"`
	Message string `json:"message"`
}

// CreateTransaction 创建交易记录
func (s *TransactionService) CreateTransaction(req *models.CreateTransactionRequest, institutionID string) (*CreateTransactionResult, error) {
	// 1. 检查业务流水号是否已存在
	var existingTx models.Transaction
	err := s.db.Where("biz_id = ?", req.BizID).First(&existingTx).Error
	if err == nil {
		return &CreateTransactionResult{
			Success: false,
			BizID:   req.BizID,
			Message: "业务流水号已存在",
		}, nil
	}

	// 2. 生成随机盐
	salt, err := utils.GenerateRandomSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// 3. 计算数据哈希(用于上链)
	dataHash := utils.CalculateDataHash(req.BizID, req.Amount, salt)

	// 4. AES加密金额
	amountCipher, err := utils.EncryptAmount(s.encryptionKey, req.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt amount: %w", err)
	}

	// 5. 创建交易记录
	tx := &models.Transaction{
		BizID:         req.BizID,
		InstitutionID: institutionID,
		AmountCipher:  amountCipher,
		AmountHash:    utils.HashPassword(req.Amount),
		DataHash:      dataHash,
		Salt:          salt,
		Receiver:      req.Receiver,
		Sender:        req.Sender,
		TxType:        req.TxType,
		Status:        models.TxStatusPending,
	}

	if err := s.db.Create(tx).Error; err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	s.logger.Info("transaction created",
		zap.String("biz_id", req.BizID),
		zap.String("institution", institutionID))

	return &CreateTransactionResult{
		Success: true,
		BizID:   req.BizID,
		Message: "创建成功",
	}, nil
}

// UploadToChain 上链
func (s *TransactionService) UploadToChain(ctx context.Context, bizId string, contractAddress string) error {
	// 1. 查询交易
	var tx models.Transaction
	if err := s.db.Where("biz_id = ?", bizId).First(&tx).Error; err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	// 2. 检查状态
	if tx.Status != models.TxStatusPending {
		return fmt.Errorf("invalid transaction status: %d", tx.Status)
	}

	// 3. 调用智能合约上传
	txHash, err := s.blockchain.SendTransaction(ctx, contractAddress, tx.BizID, tx.DataHash)
	if err != nil {
		return fmt.Errorf("failed to upload to chain: %w", err)
	}

	// 4. 获取当前区块高度
	blockNumber, err := s.blockchain.GetBlockNumber(ctx)
	if err != nil {
		s.logger.Warn("failed to get block number", zap.Error(err))
		blockNumber = 0
	}

	// 5. 保存链上回执
	receipt := &models.ChainReceipt{
		BizID:           tx.BizID,
		TxHash:          txHash,
		BlockHeight:     blockNumber,
		ContractAddress: contractAddress,
		Status:          models.ChainReceiptStatusSuccess,
	}
	if err := s.db.Create(receipt).Error; err != nil {
		s.logger.Error("failed to save receipt", zap.Error(err))
	}

	// 6. 更新交易状态
	if err := s.db.Model(&tx).Update("status", models.TxStatusUploaded).Error; err != nil {
		s.logger.Error("failed to update status", zap.Error(err))
	}

	s.logger.Info("upload to chain success",
		zap.String("biz_id", bizId),
		zap.String("tx_hash", txHash),
		zap.Int64("block_height", blockNumber))

	return nil
}

// BatchUploadToChain 批量上链
func (s *TransactionService) BatchUploadToChain(ctx context.Context, bizIds []string, contractAddress string) *models.BatchUploadResult {
	result := &models.BatchUploadResult{
		Total:   len(bizIds),
		Success: 0,
		Failed:  0,
	}

	for _, bizId := range bizIds {
		if err := s.UploadToChain(ctx, bizId, contractAddress); err != nil {
			result.Failed++
			result.FailedIDs = append(result.FailedIDs, bizId)
			s.logger.Error("batch upload failed",
				zap.String("biz_id", bizId),
				zap.Error(err))
		} else {
			result.Success++
			result.SuccessIDs = append(result.SuccessIDs, bizId)
		}
	}

	return result
}

// GetTransaction 查询交易详情
func (s *TransactionService) GetTransaction(bizId string) (*models.TransactionResponse, error) {
	var tx models.Transaction
	if err := s.db.Where("biz_id = ?", bizId).First(&tx).Error; err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	return tx.ToResponse(), nil
}

// ListTransactions 查询交易列表(分页)
func (s *TransactionService) ListTransactions(institutionID string, page, size int, status int8) (*models.PageResponse, error) {
	var txs []models.Transaction
	var total int64

	query := s.db.Model(&models.Transaction{})

	// 机构过滤
	if institutionID != "" {
		query = query.Where("institution_id = ?", institutionID)
	}

	// 状态过滤
	if status > 0 {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count transactions: %w", err)
	}

	// 分页查询
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&txs).Error; err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	// 转换为响应格式
	responses := make([]*models.TransactionResponse, len(txs))
	for i, tx := range txs {
		responses[i] = tx.ToResponse()
	}

	return &models.PageResponse{
		Total: total,
		Page:  page,
		Size:  size,
		Data:  responses,
	}, nil
}

// ParseExcelAndCreate 解析Excel文件并创建交易
func (s *TransactionService) ParseExcelAndCreate(filePath, institutionID string) (*models.BatchUploadResult, error) {
	// 1. 解析Excel
	rows, err := utils.ParseExcelFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse excel: %w", err)
	}

	result := &models.BatchUploadResult{
		Total: len(rows),
	}

	// 2. 遍历行数据
	for _, row := range rows {
		req := &models.CreateTransactionRequest{
			BizID:    row.BizID,
			Amount:   row.Amount,
			Sender:   row.Sender,
			Receiver: row.Receiver,
			TxType:   row.TxType,
		}

		// 3. 创建交易
		createResult, err := s.CreateTransaction(req, institutionID)
		if err != nil {
			result.Failed++
			result.FailedIDs = append(result.FailedIDs, row.BizID)
			s.logger.Error("failed to create transaction",
				zap.String("biz_id", row.BizID),
				zap.Error(err))
		} else if !createResult.Success {
			result.Failed++
			result.FailedIDs = append(result.FailedIDs, row.BizID)
		} else {
			result.Success++
			result.SuccessIDs = append(result.SuccessIDs, row.BizID)
		}
	}

	s.logger.Info("excel parse completed",
		zap.Int("total", result.Total),
		zap.Int("success", result.Success),
		zap.Int("failed", result.Failed))

	return result, nil
}

// GetStatistics 获取统计数据
func (s *TransactionService) GetStatistics(institutionID string) (*models.StatisticsResponse, error) {
	var stats models.StatisticsResponse

	query := s.db.Model(&models.Transaction{})
	if institutionID != "" {
		query = query.Where("institution_id = ?", institutionID)
	}

	// 总交易数
	query.Count(&stats.TotalTransactions)

	// 对账成功数
	s.db.Model(&models.Transaction{}).
		Where("status = ?", models.TxStatusMatched).
		Count(&stats.MatchedCount)

	// 对账失败数
	s.db.Model(&models.Transaction{}).
		Where("status = ?", models.TxStatusMismatch).
		Count(&stats.MismatchCount)

	// 待上链数
	s.db.Model(&models.Transaction{}).
		Where("status = ?", models.TxStatusPending).
		Count(&stats.PendingCount)

	// 已上链数
	s.db.Model(&models.Transaction{}).
		Where("status = ?", models.TxStatusUploaded).
		Count(&stats.UploadedCount)

	// 计算匹配率
	if stats.TotalTransactions > 0 {
		stats.MatchRate = float64(stats.MatchedCount) / float64(stats.TotalTransactions) * 100
	}

	return &stats, nil
}

// DecryptAmount 解密金额(用于审计)
func (s *TransactionService) DecryptAmount(bizId string) (string, error) {
	var tx models.Transaction
	if err := s.db.Where("biz_id = ?", bizId).First(&tx).Error; err != nil {
		return "", fmt.Errorf("transaction not found: %w", err)
	}

	amount, err := utils.DecryptAmount(s.encryptionKey, tx.AmountCipher)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return amount, nil
}
