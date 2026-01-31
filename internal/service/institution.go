package service

import (
	"jijin/internal/model"
	"jijin/internal/repository"
)

// InstitutionService 主力动向服务
type InstitutionService struct{}

var institutionService = &InstitutionService{}

// GetInstitutionService 获取主力动向服务实例
func GetInstitutionService() *InstitutionService {
	return institutionService
}

// GetInstitutionHolding 获取机构持仓
func (i *InstitutionService) GetInstitutionHolding(fundCode string) (*model.InstitutionHolding, error) {
	return repository.GetInstitutionHolding(fundCode)
}

// RefreshInstitutionHolding 从API刷新机构持仓数据
func (i *InstitutionService) RefreshInstitutionHolding(fundCode string) (*model.InstitutionHolding, error) {
	// 从API获取最新数据
	holding, err := GetFundAPI().GetInstitutionHolding(fundCode)
	if err != nil {
		return nil, err
	}

	// 保存到数据库
	if err := repository.SaveInstitutionHolding(holding); err != nil {
		return nil, err
	}

	return holding, nil
}

// GetAllHoldingsInstitution 获取所有持仓的机构持仓数据
func (i *InstitutionService) GetAllHoldingsInstitution() ([]model.InstitutionHolding, error) {
	holdings, err := repository.GetAllHoldings()
	if err != nil {
		return nil, err
	}

	var results []model.InstitutionHolding
	for _, h := range holdings {
		inst, err := i.RefreshInstitutionHolding(h.FundCode)
		if err != nil {
			continue
		}
		results = append(results, *inst)
	}
	return results, nil
}
