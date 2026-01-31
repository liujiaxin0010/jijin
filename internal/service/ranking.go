package service

import (
	"jijin/internal/repository"
)

// RankingService 基金排行服务
type RankingService struct{}

var rankingService = &RankingService{}

// GetRankingService 获取排行服务实例
func GetRankingService() *RankingService {
	return rankingService
}

// RankType 排行类型常量
const (
	RankTypeGain = "gain"
	RankTypeLoss = "loss"
)

// RankingItem 排行项
type RankingItem struct {
	Rank     int
	FundCode string
	FundName string
	Value    float64
	Period   string
}

// GetRanking 获取排行榜
func (r *RankingService) GetRanking(rankType, period string, limit int) ([]RankingItem, error) {
	rankings, err := repository.GetFundRankingByType(rankType, period, limit)
	if err != nil {
		return nil, err
	}

	items := make([]RankingItem, len(rankings))
	for i, rank := range rankings {
		items[i] = RankingItem{
			Rank:     rank.RankPosition,
			FundCode: rank.FundCode,
			FundName: rank.FundName,
			Value:    rank.RankValue,
			Period:   rank.Period,
		}
	}
	return items, nil
}

// GetHoldingRanking 获取持仓基金排行
func (r *RankingService) GetHoldingRanking() ([]RankingItem, error) {
	holdings, err := repository.GetAllHoldings()
	if err != nil {
		return nil, err
	}

	items := make([]RankingItem, len(holdings))
	for i, h := range holdings {
		items[i] = RankingItem{
			Rank:     i + 1,
			FundCode: h.FundCode,
			FundName: h.FundName,
			Value:    h.ProfitRate(),
		}
	}
	return items, nil
}

// RefreshRankingFromAPI 从API刷新排行数据
func (r *RankingService) RefreshRankingFromAPI(rankType string, limit int) ([]RankingItem, error) {
	// 从API获取最新排行数据
	rankings, err := GetFundAPI().GetFundRanking(rankType, limit)
	if err != nil {
		return nil, err
	}

	// 保存到数据库
	for i := range rankings {
		repository.SaveFundRanking(&rankings[i])
	}

	// 转换为RankingItem返回
	items := make([]RankingItem, len(rankings))
	for i, rank := range rankings {
		items[i] = RankingItem{
			Rank:     rank.RankPosition,
			FundCode: rank.FundCode,
			FundName: rank.FundName,
			Value:    rank.RankValue,
			Period:   rank.Period,
		}
	}
	return items, nil
}
