package service

import (
	"math"
)

// IndicatorService 技术指标计算服务
type IndicatorService struct{}

var indicatorService = &IndicatorService{}

// GetIndicatorService 获取指标服务实例
func GetIndicatorService() *IndicatorService {
	return indicatorService
}

// MACDResult MACD计算结果
type MACDResult struct {
	MACD   []float64
	Signal []float64
	Hist   []float64
}

// CalculateEMA 计算指数移动平均线
func (i *IndicatorService) CalculateEMA(prices []float64, period int) []float64 {
	if len(prices) < period {
		return nil
	}
	ema := make([]float64, len(prices))
	multiplier := 2.0 / float64(period+1)

	// 第一个EMA值使用SMA
	sum := 0.0
	for j := 0; j < period; j++ {
		sum += prices[j]
	}
	ema[period-1] = sum / float64(period)

	// 计算后续EMA
	for j := period; j < len(prices); j++ {
		ema[j] = (prices[j]-ema[j-1])*multiplier + ema[j-1]
	}
	return ema
}

// CalculateMACD 计算MACD指标
func (i *IndicatorService) CalculateMACD(prices []float64) *MACDResult {
	if len(prices) < 26 {
		return nil
	}
	ema12 := i.CalculateEMA(prices, 12)
	ema26 := i.CalculateEMA(prices, 26)

	macd := make([]float64, len(prices))
	for j := 25; j < len(prices); j++ {
		macd[j] = ema12[j] - ema26[j]
	}

	signal := i.CalculateEMA(macd[25:], 9)
	hist := make([]float64, len(prices))
	for j := 0; j < len(signal); j++ {
		hist[j+25] = macd[j+25] - signal[j]
	}

	return &MACDResult{MACD: macd, Signal: signal, Hist: hist}
}

// CalculateRSI 计算RSI指标
func (i *IndicatorService) CalculateRSI(prices []float64, period int) []float64 {
	if len(prices) < period+1 {
		return nil
	}
	rsi := make([]float64, len(prices))
	gains := make([]float64, len(prices))
	losses := make([]float64, len(prices))

	for j := 1; j < len(prices); j++ {
		change := prices[j] - prices[j-1]
		if change > 0 {
			gains[j] = change
		} else {
			losses[j] = -change
		}
	}

	avgGain := 0.0
	avgLoss := 0.0
	for j := 1; j <= period; j++ {
		avgGain += gains[j]
		avgLoss += losses[j]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	if avgLoss == 0 {
		rsi[period] = 100
	} else {
		rsi[period] = 100 - 100/(1+avgGain/avgLoss)
	}

	for j := period + 1; j < len(prices); j++ {
		avgGain = (avgGain*float64(period-1) + gains[j]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[j]) / float64(period)
		if avgLoss == 0 {
			rsi[j] = 100
		} else {
			rsi[j] = 100 - 100/(1+avgGain/avgLoss)
		}
	}
	return rsi
}

// KDJResult KDJ计算结果
type KDJResult struct {
	K []float64
	D []float64
	J []float64
}

// CalculateKDJ 计算KDJ指标
func (i *IndicatorService) CalculateKDJ(prices []float64, period int) *KDJResult {
	if len(prices) < period {
		return nil
	}
	n := len(prices)
	k := make([]float64, n)
	d := make([]float64, n)
	j := make([]float64, n)

	for idx := period - 1; idx < n; idx++ {
		high, low := prices[idx], prices[idx]
		for p := 0; p < period; p++ {
			if prices[idx-p] > high {
				high = prices[idx-p]
			}
			if prices[idx-p] < low {
				low = prices[idx-p]
			}
		}
		rsv := 0.0
		if high != low {
			rsv = (prices[idx] - low) / (high - low) * 100
		}
		if idx == period-1 {
			k[idx] = rsv
			d[idx] = rsv
		} else {
			k[idx] = 2.0/3.0*k[idx-1] + 1.0/3.0*rsv
			d[idx] = 2.0/3.0*d[idx-1] + 1.0/3.0*k[idx]
		}
		j[idx] = 3*k[idx] - 2*d[idx]
	}
	return &KDJResult{K: k, D: d, J: j}
}

// BollingerResult 布林带结果
type BollingerResult struct {
	Upper  []float64
	Middle []float64
	Lower  []float64
}

// CalculateBollinger 计算布林带
func (i *IndicatorService) CalculateBollinger(prices []float64, period int) *BollingerResult {
	if len(prices) < period {
		return nil
	}
	n := len(prices)
	upper := make([]float64, n)
	middle := make([]float64, n)
	lower := make([]float64, n)

	for idx := period - 1; idx < n; idx++ {
		sum := 0.0
		for p := 0; p < period; p++ {
			sum += prices[idx-p]
		}
		ma := sum / float64(period)

		variance := 0.0
		for p := 0; p < period; p++ {
			variance += (prices[idx-p] - ma) * (prices[idx-p] - ma)
		}
		std := math.Sqrt(variance / float64(period))

		middle[idx] = ma
		upper[idx] = ma + 2*std
		lower[idx] = ma - 2*std
	}
	return &BollingerResult{Upper: upper, Middle: middle, Lower: lower}
}
