package service

import (
	"math"
	"math/rand"
	"time"

	"jijin/internal/repository"
)

// PredictionService 预测分析服务
type PredictionService struct{}

var predictionService = &PredictionService{}

// GetPredictionService 获取预测服务实例
func GetPredictionService() *PredictionService {
	return predictionService
}

// RecoveryResult 回本预测结果
type RecoveryResult struct {
	FundCode       string
	FundName       string
	CurrentLoss    float64
	LossRate       float64
	PredictedDays  int
	ConfidenceLevel float64
	Suggestion     string
}

// ProbabilityResult 盈利概率结果
type ProbabilityResult struct {
	Period         int
	ProfitProb     float64
	ExpectedReturn float64
	WorstCase      float64
	BestCase       float64
}

// PredictRecoveryTime 预测回本时间
func (p *PredictionService) PredictRecoveryTime(holdingID uint) (*RecoveryResult, error) {
	holding, err := repository.GetHolding(holdingID)
	if err != nil {
		return nil, err
	}

	profit := holding.Profit()
	if profit >= 0 {
		return &RecoveryResult{
			FundCode:   holding.FundCode,
			FundName:   holding.FundName,
			Suggestion: "当前持仓已盈利，无需回本",
		}, nil
	}

	lossRate := holding.ProfitRate()
	mean, std, err := p.calculateHistoricalReturn(holding.FundCode, 250)
	if err != nil {
		return nil, err
	}

	days, confidence := p.monteCarloSimulation(-lossRate, mean, std, 1000)

	suggestion := "建议继续持有"
	if days > 365 {
		suggestion = "回本周期较长，建议评估是否继续持有"
	}

	return &RecoveryResult{
		FundCode:        holding.FundCode,
		FundName:        holding.FundName,
		CurrentLoss:     -profit,
		LossRate:        -lossRate,
		PredictedDays:   days,
		ConfidenceLevel: confidence,
		Suggestion:      suggestion,
	}, nil
}

// calculateHistoricalReturn 计算历史收益率统计
func (p *PredictionService) calculateHistoricalReturn(fundCode string, days int) (mean, std float64, err error) {
	histories, err := repository.GetNetValueHistory(fundCode, days)
	if err != nil || len(histories) < 30 {
		return 0, 0, err
	}

	returns := make([]float64, len(histories)-1)
	for i := 0; i < len(histories)-1; i++ {
		if histories[i+1].NetValue > 0 {
			returns[i] = (histories[i].NetValue - histories[i+1].NetValue) / histories[i+1].NetValue * 100
		}
	}

	sum := 0.0
	for _, r := range returns {
		sum += r
	}
	mean = sum / float64(len(returns))

	variance := 0.0
	for _, r := range returns {
		variance += (r - mean) * (r - mean)
	}
	std = math.Sqrt(variance / float64(len(returns)))
	return mean, std, nil
}

// monteCarloSimulation 蒙特卡洛模拟
func (p *PredictionService) monteCarloSimulation(lossRate, mean, std float64, simulations int) (days int, confidence float64) {
	rand.Seed(time.Now().UnixNano())
	recoveryDays := make([]int, 0, simulations)

	for i := 0; i < simulations; i++ {
		cumReturn := 0.0
		day := 0
		for cumReturn < lossRate && day < 1000 {
			dailyReturn := rand.NormFloat64()*std + mean
			cumReturn += dailyReturn
			day++
		}
		if day < 1000 {
			recoveryDays = append(recoveryDays, day)
		}
	}

	if len(recoveryDays) == 0 {
		return 999, 0
	}

	// 计算中位数
	sum := 0
	for _, d := range recoveryDays {
		sum += d
	}
	days = sum / len(recoveryDays)
	confidence = float64(len(recoveryDays)) / float64(simulations) * 100
	return days, confidence
}

// CalculateProfitProbability 计算盈利概率
func (p *PredictionService) CalculateProfitProbability(fundCode string, periods []int) ([]ProbabilityResult, error) {
	mean, std, err := p.calculateHistoricalReturn(fundCode, 250)
	if err != nil {
		return nil, err
	}

	results := make([]ProbabilityResult, len(periods))
	for i, period := range periods {
		profitCount := 0
		returns := make([]float64, 1000)

		for j := 0; j < 1000; j++ {
			cumReturn := 0.0
			for d := 0; d < period; d++ {
				cumReturn += rand.NormFloat64()*std + mean
			}
			returns[j] = cumReturn
			if cumReturn > 0 {
				profitCount++
			}
		}

		// 计算统计值
		sum := 0.0
		minR, maxR := returns[0], returns[0]
		for _, r := range returns {
			sum += r
			if r < minR {
				minR = r
			}
			if r > maxR {
				maxR = r
			}
		}

		results[i] = ProbabilityResult{
			Period:         period,
			ProfitProb:     float64(profitCount) / 10,
			ExpectedReturn: sum / 1000,
			WorstCase:      minR,
			BestCase:       maxR,
		}
	}
	return results, nil
}
