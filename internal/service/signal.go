package service

import (
	"time"

	"jijin/internal/model"
	"jijin/internal/repository"
)

// SignalService 波段信号服务
type SignalService struct{}

var signalService = &SignalService{}

// GetSignalService 获取信号服务实例
func GetSignalService() *SignalService {
	return signalService
}

// SignalType 信号类型常量
const (
	SignalBuy  = "buy"
	SignalSell = "sell"
	SignalHold = "hold"
)

// GenerateSignal 生成交易信号
func (s *SignalService) GenerateSignal(fundCode string) (*model.TradingSignal, error) {
	histories, err := repository.GetNetValueHistory(fundCode, 60)
	if err != nil {
		return nil, err
	}
	if len(histories) < 30 {
		// 数据不足，返回默认持有信号
		return &model.TradingSignal{
			FundCode:       fundCode,
			SignalType:     SignalHold,
			SignalStrength: 50,
			Indicator:      "RSI",
			Reason:         "历史数据不足，暂无信号",
			GeneratedAt:    time.Now(),
			IsValid:        true,
		}, nil
	}

	prices := make([]float64, len(histories))
	for i, h := range histories {
		prices[len(histories)-1-i] = h.NetValue
	}

	indicator := GetIndicatorService()
	rsi := indicator.CalculateRSI(prices, 14)

	signalType := SignalHold
	reason := "观望"
	strength := 50.0

	if len(rsi) > 0 {
		lastRSI := rsi[len(rsi)-1]
		if lastRSI < 30 {
			signalType = SignalBuy
			reason = "RSI超卖，建议买入"
			strength = 80
		} else if lastRSI > 70 {
			signalType = SignalSell
			reason = "RSI超买，建议卖出"
			strength = 80
		}
	}

	signal := &model.TradingSignal{
		FundCode:       fundCode,
		SignalType:     signalType,
		SignalStrength: strength,
		Indicator:      "RSI",
		Price:          prices[len(prices)-1],
		Reason:         reason,
		GeneratedAt:    time.Now(),
		IsValid:        true,
	}
	repository.SaveTradingSignal(signal)
	return signal, nil
}
