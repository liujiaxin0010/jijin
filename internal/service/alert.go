package service

import (
	"fmt"
	"sync"
	"time"

	"jijin/internal/model"
	"jijin/internal/repository"
)

// AlertType 提醒类型常量
const (
	AlertTypePriceChange = "price_change" // 盘中涨跌提醒
	AlertTypeConsecutive = "consecutive"  // 连涨连跌提醒
	AlertTypeNavUpdate   = "nav_update"   // 净值更新提醒
)

// AlertService 智能提醒服务
type AlertService struct {
	mu            sync.Mutex
	isRunning     bool
	stopChan      chan bool
	alertCallback func(alert *model.AlertHistory)
}

var alertService *AlertService
var alertOnce sync.Once

// GetAlertService 获取提醒服务实例
func GetAlertService() *AlertService {
	alertOnce.Do(func() {
		alertService = &AlertService{
			stopChan: make(chan bool),
		}
	})
	return alertService
}

// CreateAlertRule 创建提醒规则
func (a *AlertService) CreateAlertRule(rule *model.AlertRule) error {
	rule.Enabled = true
	return repository.SaveAlertRule(rule)
}

// UpdateAlertRule 更新提醒规则
func (a *AlertService) UpdateAlertRule(rule *model.AlertRule) error {
	return repository.SaveAlertRule(rule)
}

// DeleteAlertRule 删除提醒规则
func (a *AlertService) DeleteAlertRule(id uint) error {
	return repository.DeleteAlertRule(id)
}

// GetAlertRules 获取提醒规则
func (a *AlertService) GetAlertRules(fundCode string) ([]model.AlertRule, error) {
	if fundCode == "" {
		return repository.GetAllAlertRules()
	}
	return repository.GetAlertRulesByFundCode(fundCode)
}

// ToggleAlertRule 切换规则状态
func (a *AlertService) ToggleAlertRule(id uint) error {
	rule, err := repository.GetAlertRule(id)
	if err != nil {
		return err
	}
	rule.Enabled = !rule.Enabled
	return repository.SaveAlertRule(rule)
}

// StartMonitoring 启动监控
func (a *AlertService) StartMonitoring(callback func(alert *model.AlertHistory)) {
	a.mu.Lock()
	if a.isRunning {
		a.mu.Unlock()
		return
	}
	a.isRunning = true
	a.alertCallback = callback
	a.mu.Unlock()

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if isTradingTime() {
					a.checkAllAlerts()
				}
			case <-a.stopChan:
				return
			}
		}
	}()
}

// StopMonitoring 停止监控
func (a *AlertService) StopMonitoring() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.isRunning {
		a.stopChan <- true
		a.isRunning = false
	}
}

// checkAllAlerts 检查所有提醒
func (a *AlertService) checkAllAlerts() {
	rules, err := repository.GetEnabledAlertRules()
	if err != nil {
		return
	}

	for _, rule := range rules {
		var alert *model.AlertHistory
		switch rule.AlertType {
		case AlertTypePriceChange:
			alert = a.checkPriceChangeAlert(&rule)
		case AlertTypeConsecutive:
			alert = a.checkConsecutiveAlert(&rule)
		}
		if alert != nil && a.alertCallback != nil {
			repository.SaveAlertHistory(alert)
			a.alertCallback(alert)
		}
	}
}

// checkPriceChangeAlert 检查盘中涨跌提醒
func (a *AlertService) checkPriceChangeAlert(rule *model.AlertRule) *model.AlertHistory {
	fund, err := GetFundAPI().GetFundDetail(rule.FundCode)
	if err != nil {
		return nil
	}

	triggered := false
	var message string

	switch rule.Direction {
	case "up":
		if fund.EstGrowth >= rule.Threshold {
			triggered = true
			message = fmt.Sprintf("%s 涨幅达到 %.2f%%", fund.Name, fund.EstGrowth)
		}
	case "down":
		if fund.EstGrowth <= -rule.Threshold {
			triggered = true
			message = fmt.Sprintf("%s 跌幅达到 %.2f%%", fund.Name, fund.EstGrowth)
		}
	case "both":
		if fund.EstGrowth >= rule.Threshold || fund.EstGrowth <= -rule.Threshold {
			triggered = true
			message = fmt.Sprintf("%s 涨跌幅达到 %.2f%%", fund.Name, fund.EstGrowth)
		}
	}

	if !triggered {
		return nil
	}

	return &model.AlertHistory{
		RuleID:      rule.ID,
		FundCode:    rule.FundCode,
		FundName:    fund.Name,
		AlertType:   AlertTypePriceChange,
		Message:     message,
		Value:       fund.EstGrowth,
		TriggeredAt: time.Now(),
	}
}

// checkConsecutiveAlert 检查连涨连跌提醒
func (a *AlertService) checkConsecutiveAlert(rule *model.AlertRule) *model.AlertHistory {
	histories, err := repository.GetNetValueHistory(rule.FundCode, rule.ConsecutiveDays+1)
	if err != nil || len(histories) < rule.ConsecutiveDays {
		return nil
	}

	upCount, downCount := 0, 0
	for i := 0; i < len(histories)-1; i++ {
		if histories[i].DayGrowth > 0 {
			upCount++
		} else if histories[i].DayGrowth < 0 {
			downCount++
		}
	}

	var message string
	triggered := false

	if rule.Direction == "up" && upCount >= rule.ConsecutiveDays {
		triggered = true
		message = fmt.Sprintf("%s 连涨%d天", rule.FundName, upCount)
	} else if rule.Direction == "down" && downCount >= rule.ConsecutiveDays {
		triggered = true
		message = fmt.Sprintf("%s 连跌%d天", rule.FundName, downCount)
	}

	if !triggered {
		return nil
	}

	return &model.AlertHistory{
		RuleID:      rule.ID,
		FundCode:    rule.FundCode,
		FundName:    rule.FundName,
		AlertType:   AlertTypeConsecutive,
		Message:     message,
		Value:       float64(upCount),
		TriggeredAt: time.Now(),
	}
}

// GetAlertHistory 获取提醒历史
func (a *AlertService) GetAlertHistory(fundCode string, limit int) ([]model.AlertHistory, error) {
	return repository.GetAlertHistoryByFundCode(fundCode, limit)
}

// GetUnreadAlerts 获取未读提醒
func (a *AlertService) GetUnreadAlerts() ([]model.AlertHistory, error) {
	return repository.GetUnreadAlertHistory()
}

// MarkAsRead 标记提醒为已读
func (a *AlertService) MarkAsRead(id uint) error {
	return repository.MarkAlertAsRead(id)
}

// isTradingTime 判断是否为交易时间
func isTradingTime() bool {
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}
	hour, minute := now.Hour(), now.Minute()
	currentMinutes := hour*60 + minute
	// 9:30-15:00
	return currentMinutes >= 9*60+30 && currentMinutes <= 15*60
}
