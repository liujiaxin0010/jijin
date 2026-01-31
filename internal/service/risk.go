package service

import (
	"math"

	"jijin/internal/repository"
)

// RiskService 风险分析服务
type RiskService struct{}

var riskService = &RiskService{}

// GetRiskService 获取风险服务实例
func GetRiskService() *RiskService {
	return riskService
}

// RiskResult 风险分析结果
type RiskResult struct {
	FundCode    string
	FundName    string
	RiskLevel   int
	RiskScore   float64
	MaxDrawdown float64
	Volatility  float64
	Suggestion  string
}

// AnalyzeFundRisk 分析基金风险
func (r *RiskService) AnalyzeFundRisk(fundCode string) (*RiskResult, error) {
	fund, err := repository.GetFund(fundCode)
	if err != nil {
		return nil, err
	}

	maxDrawdown, _ := r.CalculateMaxDrawdown(fundCode, 250)
	volatility, _ := r.CalculateVolatility(fundCode, 250)

	riskScore := maxDrawdown*0.4 + volatility*0.6
	riskLevel := int(riskScore/20) + 1
	if riskLevel > 5 {
		riskLevel = 5
	}

	suggestion := "风险适中"
	if riskLevel >= 4 {
		suggestion = "风险较高，建议谨慎"
	} else if riskLevel <= 2 {
		suggestion = "风险较低，适合稳健投资"
	}

	return &RiskResult{
		FundCode:    fundCode,
		FundName:    fund.Name,
		RiskLevel:   riskLevel,
		RiskScore:   riskScore,
		MaxDrawdown: maxDrawdown,
		Volatility:  volatility,
		Suggestion:  suggestion,
	}, nil
}

// CalculateMaxDrawdown 计算最大回撤
func (r *RiskService) CalculateMaxDrawdown(fundCode string, days int) (float64, error) {
	histories, err := repository.GetNetValueHistory(fundCode, days)
	if err != nil || len(histories) < 2 {
		return 0, err
	}

	maxDrawdown := 0.0
	peak := histories[len(histories)-1].NetValue

	for i := len(histories) - 2; i >= 0; i-- {
		if histories[i].NetValue > peak {
			peak = histories[i].NetValue
		}
		drawdown := (peak - histories[i].NetValue) / peak * 100
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}
	return maxDrawdown, nil
}

// CalculateVolatility 计算波动率
func (r *RiskService) CalculateVolatility(fundCode string, days int) (float64, error) {
	histories, err := repository.GetNetValueHistory(fundCode, days)
	if err != nil || len(histories) < 2 {
		return 0, err
	}

	returns := make([]float64, len(histories)-1)
	for i := 0; i < len(histories)-1; i++ {
		if histories[i+1].NetValue > 0 {
			returns[i] = (histories[i].NetValue - histories[i+1].NetValue) / histories[i+1].NetValue * 100
		}
	}

	sum := 0.0
	for _, ret := range returns {
		sum += ret
	}
	mean := sum / float64(len(returns))

	variance := 0.0
	for _, ret := range returns {
		variance += (ret - mean) * (ret - mean)
	}
	return math.Sqrt(variance/float64(len(returns))) * math.Sqrt(250), nil
}
