package model

import (
	"time"

	"gorm.io/gorm"
)

// Fund 基金信息
type Fund struct {
	Code       string    `json:"code" gorm:"primaryKey;size:10"`
	Name       string    `json:"name" gorm:"size:100"`
	Type       string    `json:"type" gorm:"size:50"`
	NetValue   float64   `json:"netValue"`   // 单位净值
	TotalValue float64   `json:"totalValue"` // 累计净值
	DayGrowth  float64   `json:"dayGrowth"`  // 日涨跌幅(%)
	EstValue   float64   `json:"estValue"`   // 估算净值
	EstGrowth  float64   `json:"estGrowth"`  // 估算涨跌幅(%)
	UpdatedAt  time.Time `json:"updatedAt"`
}

// Holding 持仓
type Holding struct {
	gorm.Model
	FundCode   string  `json:"fundCode" gorm:"size:10;index"`
	FundName   string  `json:"fundName" gorm:"size:100"`
	Shares     float64 `json:"shares"`    // 持有份额
	Cost       float64 `json:"cost"`      // 持仓成本(总投入)
	CostPrice  float64 `json:"costPrice"` // 成本价(每份)
	CurrentNav float64 `json:"currentNav"` // 当前净值
}

// 计算当前市值
func (h *Holding) MarketValue() float64 {
	return h.Shares * h.CurrentNav
}

// 计算盈亏金额
func (h *Holding) Profit() float64 {
	return h.MarketValue() - h.Cost
}

// 计算盈亏比例
func (h *Holding) ProfitRate() float64 {
	if h.Cost == 0 {
		return 0
	}
	return (h.MarketValue() - h.Cost) / h.Cost * 100
}

// Transaction 交易记录
type Transaction struct {
	gorm.Model
	FundCode  string    `json:"fundCode" gorm:"size:10;index"`
	FundName  string    `json:"fundName" gorm:"size:100"`
	Type      string    `json:"type" gorm:"size:10"` // buy/sell
	Amount    float64   `json:"amount"`              // 交易金额
	NetValue  float64   `json:"netValue"`            // 成交净值
	Shares    float64   `json:"shares"`              // 成交份额
	Fee       float64   `json:"fee"`                 // 手续费
	TradeDate time.Time `json:"tradeDate"`           // 交易日期
}

// Strategy 定投策略
type Strategy struct {
	gorm.Model
	Name         string  `json:"name" gorm:"size:100"`
	FundCode     string  `json:"fundCode" gorm:"size:10;index"`
	FundName     string  `json:"fundName" gorm:"size:100"`
	BaseAmount   float64 `json:"baseAmount"`            // 基准定投金额
	Frequency    string  `json:"frequency" gorm:"size:20"` // daily/weekly/monthly
	StrategyType string  `json:"strategyType" gorm:"size:30"` // normal/ma_deviation/valuation/target_value
	Params       string  `json:"params" gorm:"type:text"`  // JSON参数
	Active       bool    `json:"active"`
}

// NetValueHistory 净值历史
type NetValueHistory struct {
	ID        uint      `gorm:"primaryKey"`
	FundCode  string    `json:"fundCode" gorm:"size:10;index"`
	NetValue  float64   `json:"netValue"`
	TotalValue float64  `json:"totalValue"`
	DayGrowth float64   `json:"dayGrowth"`
	Date      time.Time `json:"date" gorm:"index"`
}

// FundSearchResult 基金搜索结果(不存储)
type FundSearchResult struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Pinyin   string `json:"pinyin"`
	PinyinAbbr string `json:"pinyinAbbr"`
}

// ========== 智能提醒相关 ==========

// AlertRule 提醒规则
type AlertRule struct {
	gorm.Model
	FundCode        string    `json:"fundCode" gorm:"size:10;index"`
	FundName        string    `json:"fundName" gorm:"size:100"`
	AlertType       string    `json:"alertType" gorm:"size:30"`       // price_change/consecutive/nav_update
	Threshold       float64   `json:"threshold"`                      // 阈值（如涨跌幅百分比）
	Direction       string    `json:"direction" gorm:"size:10"`       // up/down/both
	ConsecutiveDays int       `json:"consecutiveDays"`                // 连续天数（连涨连跌用）
	Enabled         bool      `json:"enabled"`
	LastTriggered   time.Time `json:"lastTriggered"`
}

// AlertHistory 提醒历史
type AlertHistory struct {
	gorm.Model
	RuleID      uint      `json:"ruleId" gorm:"index"`
	FundCode    string    `json:"fundCode" gorm:"size:10;index"`
	FundName    string    `json:"fundName" gorm:"size:100"`
	AlertType   string    `json:"alertType" gorm:"size:30"`
	Message     string    `json:"message" gorm:"size:500"`
	Value       float64   `json:"value"`                              // 触发时的值
	TriggeredAt time.Time `json:"triggeredAt"`
	IsRead      bool      `json:"isRead"`
}

// ========== 基金排行相关 ==========

// FundRanking 基金排行数据
type FundRanking struct {
	gorm.Model
	FundCode     string    `json:"fundCode" gorm:"size:10;index"`
	FundName     string    `json:"fundName" gorm:"size:100"`
	RankType     string    `json:"rankType" gorm:"size:20;index"` // add_position/reduce_position/gain/loss
	RankValue    float64   `json:"rankValue"`                     // 排行值
	RankPosition int       `json:"rankPosition"`                  // 排名位置
	Period       string    `json:"period" gorm:"size:10"`         // day/week/month
	StatDate     time.Time `json:"statDate" gorm:"index"`
}

// ========== 基金风险分析相关 ==========

// FundRiskProfile 基金风险档案
type FundRiskProfile struct {
	gorm.Model
	FundCode           string    `json:"fundCode" gorm:"size:10;uniqueIndex"`
	FundName           string    `json:"fundName" gorm:"size:100"`
	RiskLevel          int       `json:"riskLevel"`          // 1-5 风险等级
	RiskScore          float64   `json:"riskScore"`          // 综合风险评分 0-100
	ManagerChangeCount int       `json:"managerChangeCount"` // 基金经理变更次数
	ScaleChangeRate    float64   `json:"scaleChangeRate"`    // 规模变化率
	MaxDrawdown        float64   `json:"maxDrawdown"`        // 最大回撤
	Volatility         float64   `json:"volatility"`         // 波动率
	SharpeRatio        float64   `json:"sharpeRatio"`        // 夏普比率
	RiskFactors        string    `json:"riskFactors" gorm:"type:text"` // JSON: 风险因素列表
	UpdatedAt          time.Time `json:"updatedAt"`
}

// ========== 主力动向相关 ==========

// InstitutionHolding 机构持仓
type InstitutionHolding struct {
	gorm.Model
	FundCode          string    `json:"fundCode" gorm:"size:10;index"`
	FundName          string    `json:"fundName" gorm:"size:100"`
	InstitutionRatio  float64   `json:"institutionRatio"`  // 机构持仓比例
	PersonalRatio     float64   `json:"personalRatio"`     // 个人持仓比例
	TotalShares       float64   `json:"totalShares"`       // 总份额（亿份）
	InstitutionChange float64   `json:"institutionChange"` // 机构持仓变化
	ReportDate        time.Time `json:"reportDate" gorm:"index"`
}

// ========== 预测分析相关 ==========

// RecoveryPrediction 回本预测
type RecoveryPrediction struct {
	gorm.Model
	HoldingID        uint      `json:"holdingId" gorm:"index"`
	FundCode         string    `json:"fundCode" gorm:"size:10;index"`
	CurrentLoss      float64   `json:"currentLoss"`      // 当前亏损金额
	LossRate         float64   `json:"lossRate"`         // 亏损比例
	PredictedDays    int       `json:"predictedDays"`    // 预测回本天数
	ConfidenceLevel  float64   `json:"confidenceLevel"`  // 置信度
	PredictionMethod string    `json:"predictionMethod" gorm:"size:30"`
	CalculatedAt     time.Time `json:"calculatedAt"`
}

// ProfitProbability 盈利概率
type ProfitProbability struct {
	gorm.Model
	FundCode       string    `json:"fundCode" gorm:"size:10;index"`
	HoldingPeriod  int       `json:"holdingPeriod"`  // 持有期限（天）
	ProfitProb     float64   `json:"profitProb"`     // 盈利概率
	ExpectedReturn float64   `json:"expectedReturn"` // 预期收益率
	WorstCase      float64   `json:"worstCase"`      // 最差情况收益率
	BestCase       float64   `json:"bestCase"`       // 最好情况收益率
	CalculatedAt   time.Time `json:"calculatedAt"`
}

// ========== 波段信号相关 ==========

// TradingSignal 交易信号
type TradingSignal struct {
	gorm.Model
	FundCode       string    `json:"fundCode" gorm:"size:10;index"`
	FundName       string    `json:"fundName" gorm:"size:100"`
	SignalType     string    `json:"signalType" gorm:"size:20"`    // buy/sell/hold
	SignalStrength float64   `json:"signalStrength"`               // 信号强度 0-100
	Indicator      string    `json:"indicator" gorm:"size:30"`     // MACD/RSI/KDJ/MA
	Price          float64   `json:"price"`                        // 信号价格
	TargetPrice    float64   `json:"targetPrice"`                  // 目标价格
	StopLoss       float64   `json:"stopLoss"`                     // 止损价格
	Reason         string    `json:"reason" gorm:"size:500"`       // 信号原因
	GeneratedAt    time.Time `json:"generatedAt" gorm:"index"`
	IsValid        bool      `json:"isValid"`                      // 是否仍有效
}

// ========== 截图导入相关 ==========

// ImportRecord 导入记录
type ImportRecord struct {
	gorm.Model
	ImportType    string    `json:"importType" gorm:"size:20"`    // screenshot/manual/file
	SourceApp     string    `json:"sourceApp" gorm:"size:50"`     // 来源APP
	ImportedCount int       `json:"importedCount"`                // 导入数量
	SuccessCount  int       `json:"successCount"`                 // 成功数量
	FailedCount   int       `json:"failedCount"`                  // 失败数量
	Details       string    `json:"details" gorm:"type:text"`     // JSON详情
	ImportedAt    time.Time `json:"importedAt"`
}

// ========== QDII基金相关 ==========

// QDIIFund QDII基金扩展信息
type QDIIFund struct {
	gorm.Model
	FundCode      string    `json:"fundCode" gorm:"size:10;uniqueIndex"`
	FundName      string    `json:"fundName" gorm:"size:100"`
	MarketType    string    `json:"marketType" gorm:"size:20"`    // us/hk/global
	TrackingIndex string    `json:"trackingIndex" gorm:"size:100"`
	Currency      string    `json:"currency" gorm:"size:10"`      // USD/HKD/CNY
	ExchangeRate  float64   `json:"exchangeRate"`
	T1EstValue    float64   `json:"t1EstValue"`                   // T+1估值
	T2EstValue    float64   `json:"t2EstValue"`                   // T+2估值
	LastTradeDate time.Time `json:"lastTradeDate"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
