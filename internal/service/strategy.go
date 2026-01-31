package service

import (
	"encoding/json"
	"math"

	"jijin/internal/model"
	"jijin/internal/repository"
)

// StrategyService 策略服务
type StrategyService struct{}

var strategyService = &StrategyService{}

// GetStrategyService 获取策略服务实例
func GetStrategyService() *StrategyService {
	return strategyService
}

// MADeviationParams 均线偏离策略参数
type MADeviationParams struct {
	MAPeriod      int     `json:"maPeriod"`      // 均线周期(天)
	MaxMultiplier float64 `json:"maxMultiplier"` // 最大倍数
	MinMultiplier float64 `json:"minMultiplier"` // 最小倍数
}

// ValuationParams 估值定投策略参数
type ValuationParams struct {
	LowPE         float64 `json:"lowPE"`         // 低估PE
	HighPE        float64 `json:"highPE"`        // 高估PE
	MaxMultiplier float64 `json:"maxMultiplier"` // 最大倍数
	MinMultiplier float64 `json:"minMultiplier"` // 最小倍数
}

// TargetValueParams 目标市值策略参数
type TargetValueParams struct {
	TargetValue float64 `json:"targetValue"` // 目标市值
	GrowthRate  float64 `json:"growthRate"`  // 每期增长率(%)
}

// StrategyResult 策略计算结果
type StrategyResult struct {
	BaseAmount    float64 `json:"baseAmount"`    // 基准金额
	SuggestAmount float64 `json:"suggestAmount"` // 建议金额
	Multiplier    float64 `json:"multiplier"`    // 倍数
	Reason        string  `json:"reason"`        // 原因说明
}

// CreateStrategy 创建策略
func (s *StrategyService) CreateStrategy(name, fundCode, fundName, frequency, strategyType string, baseAmount float64, params interface{}) (*model.Strategy, error) {
	paramsJSON, _ := json.Marshal(params)

	strategy := &model.Strategy{
		Name:         name,
		FundCode:     fundCode,
		FundName:     fundName,
		BaseAmount:   baseAmount,
		Frequency:    frequency,
		StrategyType: strategyType,
		Params:       string(paramsJSON),
		Active:       true,
	}

	if err := repository.SaveStrategy(strategy); err != nil {
		return nil, err
	}

	return strategy, nil
}

// CalculateMADeviation 计算均线偏离策略建议金额
func (s *StrategyService) CalculateMADeviation(fundCode string, baseAmount float64, params MADeviationParams) (*StrategyResult, error) {
	// 获取历史净值
	histories, err := GetFundAPI().GetFundHistory(fundCode, params.MAPeriod+10)
	if err != nil || len(histories) < params.MAPeriod {
		return &StrategyResult{
			BaseAmount:    baseAmount,
			SuggestAmount: baseAmount,
			Multiplier:    1.0,
			Reason:        "数据不足，使用基准金额",
		}, nil
	}

	// 计算均线
	var sum float64
	for i := 0; i < params.MAPeriod; i++ {
		sum += histories[i].NetValue
	}
	ma := sum / float64(params.MAPeriod)

	// 当前净值
	currentNav := histories[0].NetValue

	// 计算偏离度
	deviation := (currentNav - ma) / ma

	// 计算倍数: 偏离度越负(低于均线)，倍数越大
	multiplier := 1.0 - deviation*2 // 简单线性映射

	// 限制范围
	if multiplier > params.MaxMultiplier {
		multiplier = params.MaxMultiplier
	}
	if multiplier < params.MinMultiplier {
		multiplier = params.MinMultiplier
	}

	suggestAmount := baseAmount * multiplier

	var reason string
	if deviation < -0.1 {
		reason = "低于均线10%以上，建议加倍定投"
	} else if deviation < 0 {
		reason = "低于均线，建议适当增加"
	} else if deviation > 0.1 {
		reason = "高于均线10%以上，建议减少定投"
	} else {
		reason = "接近均线，正常定投"
	}

	return &StrategyResult{
		BaseAmount:    baseAmount,
		SuggestAmount: math.Round(suggestAmount*100) / 100,
		Multiplier:    math.Round(multiplier*100) / 100,
		Reason:        reason,
	}, nil
}

// CalculateValuation 计算估值定投策略建议金额
func (s *StrategyService) CalculateValuation(fundCode string, baseAmount float64, currentPE float64, params ValuationParams) (*StrategyResult, error) {
	var multiplier float64
	var reason string

	// 根据PE估值计算倍数
	if currentPE <= params.LowPE {
		multiplier = params.MaxMultiplier
		reason = "估值低估，建议大幅加仓"
	} else if currentPE >= params.HighPE {
		multiplier = params.MinMultiplier
		reason = "估值高估，建议减少定投"
	} else {
		// 线性插值
		ratio := (currentPE - params.LowPE) / (params.HighPE - params.LowPE)
		multiplier = params.MaxMultiplier - ratio*(params.MaxMultiplier-params.MinMultiplier)
		reason = "估值适中，正常定投"
	}

	suggestAmount := baseAmount * multiplier

	return &StrategyResult{
		BaseAmount:    baseAmount,
		SuggestAmount: math.Round(suggestAmount*100) / 100,
		Multiplier:    math.Round(multiplier*100) / 100,
		Reason:        reason,
	}, nil
}

// CalculateTargetValue 计算目标市值策略建议金额
func (s *StrategyService) CalculateTargetValue(holding *model.Holding, params TargetValueParams, periods int) (*StrategyResult, error) {
	// 计算目标市值(含增长)
	targetValue := params.TargetValue * math.Pow(1+params.GrowthRate/100, float64(periods))

	// 当前市值
	currentValue := holding.MarketValue()

	// 需要补充的金额
	suggestAmount := targetValue - currentValue

	if suggestAmount < 0 {
		suggestAmount = 0
	}

	var reason string
	if currentValue >= targetValue {
		reason = "已达目标市值，本期无需定投"
	} else {
		reason = "距目标市值还差" + formatMoney(targetValue-currentValue)
	}

	return &StrategyResult{
		BaseAmount:    params.TargetValue,
		SuggestAmount: math.Round(suggestAmount*100) / 100,
		Multiplier:    suggestAmount / params.TargetValue,
		Reason:        reason,
	}, nil
}

// GetAllStrategies 获取所有策略
func (s *StrategyService) GetAllStrategies() ([]model.Strategy, error) {
	return repository.GetAllStrategies()
}

// GetStrategy 获取策略详情
func (s *StrategyService) GetStrategy(id uint) (*model.Strategy, error) {
	return repository.GetStrategy(id)
}

// UpdateStrategy 更新策略
func (s *StrategyService) UpdateStrategy(strategy *model.Strategy) error {
	return repository.SaveStrategy(strategy)
}

// DeleteStrategy 删除策略
func (s *StrategyService) DeleteStrategy(id uint) error {
	return repository.DeleteStrategy(id)
}

// ToggleStrategy 切换策略状态
func (s *StrategyService) ToggleStrategy(id uint) error {
	strategy, err := repository.GetStrategy(id)
	if err != nil {
		return err
	}

	strategy.Active = !strategy.Active
	return repository.SaveStrategy(strategy)
}

func formatMoney(amount float64) string {
	if amount >= 10000 {
		return string(rune('¥')) + string(rune(int(amount/10000))) + "万"
	}
	return string(rune('¥')) + string(rune(int(amount)))
}
