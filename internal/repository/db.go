package repository

import (
	"os"
	"path/filepath"
	"time"

	"jijin/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库
func InitDB() error {
	// 获取用户数据目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	dataDir := filepath.Join(homeDir, ".jijin")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	dbPath := filepath.Join(dataDir, "jijin.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	// 自动迁移
	err = db.AutoMigrate(
		// 现有模型
		&model.Fund{},
		&model.Holding{},
		&model.Transaction{},
		&model.Strategy{},
		&model.NetValueHistory{},
		// 新增模型
		&model.AlertRule{},
		&model.AlertHistory{},
		&model.FundRanking{},
		&model.FundRiskProfile{},
		&model.InstitutionHolding{},
		&model.RecoveryPrediction{},
		&model.ProfitProbability{},
		&model.TradingSignal{},
		&model.ImportRecord{},
		&model.QDIIFund{},
	)
	if err != nil {
		return err
	}

	DB = db
	return nil
}

// === Fund 操作 ===

// SaveFund 保存基金信息
func SaveFund(fund *model.Fund) error {
	fund.UpdatedAt = time.Now()
	return DB.Save(fund).Error
}

// GetFund 获取基金信息
func GetFund(code string) (*model.Fund, error) {
	var fund model.Fund
	err := DB.Where("code = ?", code).First(&fund).Error
	if err != nil {
		return nil, err
	}
	return &fund, nil
}

// GetAllFunds 获取所有基金
func GetAllFunds() ([]model.Fund, error) {
	var funds []model.Fund
	err := DB.Find(&funds).Error
	return funds, err
}

// === Holding 操作 ===

// SaveHolding 保存持仓
func SaveHolding(holding *model.Holding) error {
	return DB.Save(holding).Error
}

// GetHolding 获取持仓
func GetHolding(id uint) (*model.Holding, error) {
	var holding model.Holding
	err := DB.First(&holding, id).Error
	if err != nil {
		return nil, err
	}
	return &holding, nil
}

// GetHoldingByFundCode 根据基金代码获取持仓
func GetHoldingByFundCode(code string) (*model.Holding, error) {
	var holding model.Holding
	err := DB.Where("fund_code = ?", code).First(&holding).Error
	if err != nil {
		return nil, err
	}
	return &holding, nil
}

// GetAllHoldings 获取所有持仓
func GetAllHoldings() ([]model.Holding, error) {
	var holdings []model.Holding
	err := DB.Find(&holdings).Error
	return holdings, err
}

// DeleteHolding 删除持仓
func DeleteHolding(id uint) error {
	return DB.Delete(&model.Holding{}, id).Error
}

// === Transaction 操作 ===

// SaveTransaction 保存交易记录
func SaveTransaction(tx *model.Transaction) error {
	return DB.Create(tx).Error
}

// GetTransactionsByFundCode 获取基金的交易记录
func GetTransactionsByFundCode(code string) ([]model.Transaction, error) {
	var txs []model.Transaction
	err := DB.Where("fund_code = ?", code).Order("trade_date desc").Find(&txs).Error
	return txs, err
}

// GetAllTransactions 获取所有交易记录
func GetAllTransactions() ([]model.Transaction, error) {
	var txs []model.Transaction
	err := DB.Order("trade_date desc").Find(&txs).Error
	return txs, err
}

// DeleteTransaction 删除交易记录
func DeleteTransaction(id uint) error {
	return DB.Delete(&model.Transaction{}, id).Error
}

// === Strategy 操作 ===

// SaveStrategy 保存策略
func SaveStrategy(strategy *model.Strategy) error {
	return DB.Save(strategy).Error
}

// GetStrategy 获取策略
func GetStrategy(id uint) (*model.Strategy, error) {
	var strategy model.Strategy
	err := DB.First(&strategy, id).Error
	if err != nil {
		return nil, err
	}
	return &strategy, nil
}

// GetAllStrategies 获取所有策略
func GetAllStrategies() ([]model.Strategy, error) {
	var strategies []model.Strategy
	err := DB.Find(&strategies).Error
	return strategies, err
}

// GetActiveStrategies 获取激活的策略
func GetActiveStrategies() ([]model.Strategy, error) {
	var strategies []model.Strategy
	err := DB.Where("active = ?", true).Find(&strategies).Error
	return strategies, err
}

// DeleteStrategy 删除策略
func DeleteStrategy(id uint) error {
	return DB.Delete(&model.Strategy{}, id).Error
}

// === NetValueHistory 操作 ===

// SaveNetValueHistory 保存净值历史
func SaveNetValueHistory(history *model.NetValueHistory) error {
	return DB.Save(history).Error
}

// SaveNetValueHistories 批量保存净值历史
func SaveNetValueHistories(histories []model.NetValueHistory) error {
	if len(histories) == 0 {
		return nil
	}
	return DB.CreateInBatches(histories, 100).Error
}

// GetNetValueHistory 获取基金净值历史
func GetNetValueHistory(code string, days int) ([]model.NetValueHistory, error) {
	var histories []model.NetValueHistory
	err := DB.Where("fund_code = ?", code).
		Order("date desc").
		Limit(days).
		Find(&histories).Error
	return histories, err
}

// GetLatestNetValue 获取最新净值
func GetLatestNetValue(code string) (*model.NetValueHistory, error) {
	var history model.NetValueHistory
	err := DB.Where("fund_code = ?", code).
		Order("date desc").
		First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// === AlertRule 操作 ===

// SaveAlertRule 保存提醒规则
func SaveAlertRule(rule *model.AlertRule) error {
	return DB.Save(rule).Error
}

// GetAlertRule 获取提醒规则
func GetAlertRule(id uint) (*model.AlertRule, error) {
	var rule model.AlertRule
	err := DB.First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetAlertRulesByFundCode 获取基金的提醒规则
func GetAlertRulesByFundCode(code string) ([]model.AlertRule, error) {
	var rules []model.AlertRule
	err := DB.Where("fund_code = ?", code).Find(&rules).Error
	return rules, err
}

// GetAllAlertRules 获取所有提醒规则
func GetAllAlertRules() ([]model.AlertRule, error) {
	var rules []model.AlertRule
	err := DB.Find(&rules).Error
	return rules, err
}

// GetEnabledAlertRules 获取启用的提醒规则
func GetEnabledAlertRules() ([]model.AlertRule, error) {
	var rules []model.AlertRule
	err := DB.Where("enabled = ?", true).Find(&rules).Error
	return rules, err
}

// DeleteAlertRule 删除提醒规则
func DeleteAlertRule(id uint) error {
	return DB.Delete(&model.AlertRule{}, id).Error
}

// === AlertHistory 操作 ===

// SaveAlertHistory 保存提醒历史
func SaveAlertHistory(history *model.AlertHistory) error {
	return DB.Create(history).Error
}

// GetAlertHistoryByFundCode 获取基金的提醒历史
func GetAlertHistoryByFundCode(code string, limit int) ([]model.AlertHistory, error) {
	var histories []model.AlertHistory
	err := DB.Where("fund_code = ?", code).
		Order("triggered_at desc").
		Limit(limit).
		Find(&histories).Error
	return histories, err
}

// GetUnreadAlertHistory 获取未读提醒
func GetUnreadAlertHistory() ([]model.AlertHistory, error) {
	var histories []model.AlertHistory
	err := DB.Where("is_read = ?", false).
		Order("triggered_at desc").
		Find(&histories).Error
	return histories, err
}

// MarkAlertAsRead 标记提醒为已读
func MarkAlertAsRead(id uint) error {
	return DB.Model(&model.AlertHistory{}).Where("id = ?", id).Update("is_read", true).Error
}

// === FundRanking 操作 ===

// SaveFundRanking 保存基金排行
func SaveFundRanking(ranking *model.FundRanking) error {
	return DB.Save(ranking).Error
}

// GetFundRankingByType 获取指定类型的排行
func GetFundRankingByType(rankType, period string, limit int) ([]model.FundRanking, error) {
	var rankings []model.FundRanking
	err := DB.Where("rank_type = ? AND period = ?", rankType, period).
		Order("rank_position asc").
		Limit(limit).
		Find(&rankings).Error
	return rankings, err
}

// === FundRiskProfile 操作 ===

// SaveFundRiskProfile 保存风险档案
func SaveFundRiskProfile(profile *model.FundRiskProfile) error {
	return DB.Save(profile).Error
}

// GetFundRiskProfile 获取风险档案
func GetFundRiskProfile(code string) (*model.FundRiskProfile, error) {
	var profile model.FundRiskProfile
	err := DB.Where("fund_code = ?", code).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// === InstitutionHolding 操作 ===

// SaveInstitutionHolding 保存机构持仓
func SaveInstitutionHolding(holding *model.InstitutionHolding) error {
	return DB.Save(holding).Error
}

// GetInstitutionHolding 获取机构持仓
func GetInstitutionHolding(code string) (*model.InstitutionHolding, error) {
	var holding model.InstitutionHolding
	err := DB.Where("fund_code = ?", code).
		Order("report_date desc").
		First(&holding).Error
	if err != nil {
		return nil, err
	}
	return &holding, nil
}

// === TradingSignal 操作 ===

// SaveTradingSignal 保存交易信号
func SaveTradingSignal(signal *model.TradingSignal) error {
	return DB.Create(signal).Error
}

// GetActiveSignals 获取有效信号
func GetActiveSignals() ([]model.TradingSignal, error) {
	var signals []model.TradingSignal
	err := DB.Where("is_valid = ?", true).
		Order("generated_at desc").
		Find(&signals).Error
	return signals, err
}

// === ImportRecord 操作 ===

// SaveImportRecord 保存导入记录
func SaveImportRecord(record *model.ImportRecord) error {
	return DB.Create(record).Error
}

// === QDIIFund 操作 ===

// SaveQDIIFund 保存QDII基金
func SaveQDIIFund(fund *model.QDIIFund) error {
	return DB.Save(fund).Error
}

// GetQDIIFund 获取QDII基金
func GetQDIIFund(code string) (*model.QDIIFund, error) {
	var fund model.QDIIFund
	err := DB.Where("fund_code = ?", code).First(&fund).Error
	if err != nil {
		return nil, err
	}
	return &fund, nil
}
