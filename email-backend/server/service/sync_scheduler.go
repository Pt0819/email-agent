// Package service 业务逻辑层 - 同步调度器
package service

import (
	"context"
	"log"
	"sync"
	"time"

	"email-backend/server/global"
	"email-backend/server/pkg/agent"
	"email-backend/server/repository"
)

// SyncScheduler 同步调度器
// 负责定时自动同步所有账户的邮件
type SyncScheduler struct {
	accountRepo  *repository.AccountRepository
	emailRepo    *repository.EmailRepository
	agentClient  *agent.Client
	syncService  *SyncService

	// 调度控制
	interval    time.Duration
	ticker      *time.Ticker
	stopChan    chan struct{}
	running     bool
	mu          sync.Mutex

	// 统计信息
	lastSyncTime *time.Time
	syncCount    int
	errorCount   int
}

// SchedulerStatus 调度器状态
type SchedulerStatus struct {
	Running      bool       `json:"running"`
	Interval     int        `json:"interval"`      // 同步间隔(分钟)
	LastSyncTime *time.Time `json:"last_sync_time"`
	SyncCount    int        `json:"sync_count"`    // 总同步次数
	ErrorCount   int        `json:"error_count"`   // 错误次数
	NextSyncTime *time.Time `json:"next_sync_time"`
}

// NewSyncScheduler 创建同步调度器
func NewSyncScheduler(
	accountRepo *repository.AccountRepository,
	emailRepo *repository.EmailRepository,
	agentClient *agent.Client,
) *SyncScheduler {
	// 从配置获取同步间隔(分钟)
	intervalMinutes := 5
	if global.Config() != nil && global.Config().Email.SyncInterval > 0 {
		intervalMinutes = global.Config().Email.SyncInterval
	}

	return &SyncScheduler{
		accountRepo: accountRepo,
		emailRepo:   emailRepo,
		agentClient: agentClient,
		syncService: NewSyncService(accountRepo, emailRepo, agentClient),
		interval:    time.Duration(intervalMinutes) * time.Minute,
		stopChan:    make(chan struct{}),
	}
}

// Start 启动调度器
func (s *SyncScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil // 已在运行
	}

	s.ticker = time.NewTicker(s.interval)
	s.running = true

	go s.run()

	log.Printf("[Scheduler] 同步调度器已启动，间隔: %v", s.interval)
	return nil
}

// Stop 停止调度器
func (s *SyncScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	close(s.stopChan)
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.running = false

	log.Println("[Scheduler] 同步调度器已停止")
}

// run 运行调度循环
func (s *SyncScheduler) run() {
	// 启动后立即执行一次同步
	s.syncAll()

	for {
		select {
		case <-s.ticker.C:
			s.syncAll()
		case <-s.stopChan:
			return
		}
	}
}

// syncAll 同步所有账户
func (s *SyncScheduler) syncAll() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Println("[Scheduler] 开始定时同步...")

	// 获取所有用户（简化：假设单用户系统，用户ID=1）
	// TODO: 多用户支持
	userID := int64(1)

	results, err := s.syncService.SyncAll(ctx, userID)
	if err != nil {
		log.Printf("[Scheduler] 定时同步失败: %v", err)
		s.mu.Lock()
		s.errorCount++
		s.mu.Unlock()
		return
	}

	// 统计结果
	var totalSynced, totalError int
	for _, r := range results {
		totalSynced += r.SyncedCount
		totalError += r.ErrorCount
	}

	now := time.Now()
	s.mu.Lock()
	s.lastSyncTime = &now
	s.syncCount++
	s.errorCount += totalError
	s.mu.Unlock()

	log.Printf("[Scheduler] 定时同步完成: 同步%d封邮件, 错误%d", totalSynced, totalError)
}

// GetStatus 获取调度器状态
func (s *SyncScheduler) GetStatus() SchedulerStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	status := SchedulerStatus{
		Running:      s.running,
		Interval:     int(s.interval.Minutes()),
		LastSyncTime: s.lastSyncTime,
		SyncCount:    s.syncCount,
		ErrorCount:   s.errorCount,
	}

	// 计算下次同步时间
	if s.running && s.lastSyncTime != nil {
		next := s.lastSyncTime.Add(s.interval)
		status.NextSyncTime = &next
	}

	return status
}

// SetInterval 设置同步间隔
func (s *SyncScheduler) SetInterval(minutes int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.interval = time.Duration(minutes) * time.Minute

	// 如果正在运行，重新设置ticker
	if s.running && s.ticker != nil {
		s.ticker.Reset(s.interval)
		log.Printf("[Scheduler] 同步间隔已更新: %v", s.interval)
	}
}

// TriggerNow 立即触发一次同步
func (s *SyncScheduler) TriggerNow() {
	go s.syncAll()
}
