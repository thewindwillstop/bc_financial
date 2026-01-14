package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bc-reconciliation-backend/internal/blockchain"
	"bc-reconciliation-backend/internal/config"
	"bc-reconciliation-backend/internal/database"
	"bc-reconciliation-backend/internal/handler"
	"bc-reconciliation-backend/internal/middleware"
	"bc-reconciliation-backend/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// 1. 加载配置
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化日志
	logger := initLogger(cfg)
	logger.Info("Starting BC Reconciliation Backend...",
		zap.String("version", "1.0.0"),
		zap.String("mode", cfg.Server.Mode))

	// 3. 连接数据库
	db, err := database.InitMySQL(&cfg.Database.MySQL)
	if err != nil {
		logger.Fatal("Failed to connect to MySQL", zap.Error(err))
	}
	logger.Info("MySQL connected successfully")

	// 自动迁移(开发环境)
	if cfg.Server.Mode == "debug" {
		if err := database.AutoMigrate(db); err != nil {
			logger.Warn("Auto migrate failed", zap.Error(err))
		} else {
			logger.Info("Auto migrate completed")
		}
	}

	// 4. 连接区块链
	bcClient, err := blockchain.NewClient(&cfg.Blockchain, logger)
	if err != nil {
		logger.Fatal("Failed to connect to FISCO BCOS", zap.Error(err))
	}
	logger.Info("FISCO BCOS connected successfully")

	// 5. 初始化服务层
	// TODO: 从配置读取加密密钥
	encryptionKey := "your-32-byte-encryption-key!!"

	txService := service.NewTransactionService(db, bcClient, logger, encryptionKey)

	// 6. 启动事件监听(Goroutine)
	eventListener := blockchain.NewEventListener(bcClient, db, logger)
	go eventListener.Start()
	logger.Info("Event listener started")

	// 7. 设置Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(gin.Recovery())

	// 8. 注册路由
	setupRoutes(router, txService)

	// 9. 启动HTTP服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// 在Goroutine中启动服务器
	go func() {
		logger.Info("HTTP server started", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 10. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 停止事件监听
	eventListener.Stop()

	// 关闭数据库连接
	database.Close(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// initLogger 初始化日志
func initLogger(cfg *config.Config) *zap.Logger {
	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 日志级别
	atomicLevel := zap.NewAtomicLevel()
	switch cfg.Log.Level {
	case "debug":
		atomicLevel.SetLevel(zap.DebugLevel)
	case "info":
		atomicLevel.SetLevel(zap.InfoLevel)
	case "warn":
		atomicLevel.SetLevel(zap.WarnLevel)
	case "error":
		atomicLevel.SetLevel(zap.ErrorLevel)
	default:
		atomicLevel.SetLevel(zap.InfoLevel)
	}

	// 创建Core - 输出到控制台
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		atomicLevel,
	)

	// 创建Logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	return logger
}

// setupRoutes 注册路由
func setupRoutes(router *gin.Engine, txService *service.TransactionService) {
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		// 交易相关
		txHandler := handler.NewTransactionHandler(txService)
		dashboardHandler := handler.NewDashboardHandler(txService)

		transactions := v1.Group("/transactions")
		{
			transactions.POST("", txHandler.CreateTransaction)
			transactions.POST("/excel", txHandler.UploadExcel)
			transactions.POST("/upload-chain", txHandler.UploadToChain)
			transactions.GET("/template", txHandler.DownloadExcelTemplate)
			transactions.GET("/:bizId", txHandler.GetTransaction)
			transactions.GET("", txHandler.ListTransactions)
		}

		// 仪表板相关
		dashboard := v1.Group("/dashboard")
		{
			dashboard.GET("/overview", dashboardHandler.GetOverview)
			dashboard.GET("/statistics", txHandler.GetStatistics)
			dashboard.GET("/chart-data", dashboardHandler.GetChartData)
		}
	}

	// 404处理
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"code":    404,
			"message": "API not found",
		})
	})
}
