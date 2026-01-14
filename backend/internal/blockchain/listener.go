package blockchain

import (
	"context"
	"time"

	"gorm.io/gorm"
	"go.uber.org/zap"
)

// EventListener 事件监听服务
type EventListener struct {
	client *Client
	db     *gorm.DB
	logger *zap.Logger

	ctx    context.Context
	cancel context.CancelFunc
}

// NewEventListener 创建事件监听器
func NewEventListener(client *Client, db *gorm.DB, logger *zap.Logger) *EventListener {
	ctx, cancel := context.WithCancel(context.Background())

	return &EventListener{
		client: client,
		db:     db,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动事件监听
func (l *EventListener) Start() {
	l.logger.Info("event listener started")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-l.ctx.Done():
			l.logger.Info("event listener stopped")
			return
		case <-ticker.C:
			// TODO: 实现事件监听逻辑
			l.logger.Debug("event listener tick")
		}
	}
}

// Stop 停止监听
func (l *EventListener) Stop() {
	l.cancel()
}
