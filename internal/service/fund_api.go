package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"jijin/internal/model"
	"jijin/internal/repository"

	"github.com/go-resty/resty/v2"
)

var client = resty.New().
	SetTimeout(10 * time.Second).
	SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

// FundAPI 基金数据服务
type FundAPI struct {
	allFunds []model.FundSearchResult // 缓存所有基金列表
}

var fundAPI = &FundAPI{}

// GetFundAPI 获取基金API服务实例
func GetFundAPI() *FundAPI {
	return fundAPI
}

// SearchFund 搜索基金
func (f *FundAPI) SearchFund(keyword string) ([]model.FundSearchResult, error) {
	// 如果没有缓存，先加载
	if len(f.allFunds) == 0 {
		if err := f.loadAllFunds(); err != nil {
			return nil, err
		}
	}

	keyword = strings.ToLower(keyword)
	var results []model.FundSearchResult

	for _, fund := range f.allFunds {
		if strings.Contains(fund.Code, keyword) ||
			strings.Contains(strings.ToLower(fund.Name), keyword) ||
			strings.Contains(strings.ToLower(fund.Pinyin), keyword) ||
			strings.Contains(strings.ToLower(fund.PinyinAbbr), keyword) {
			results = append(results, fund)
			if len(results) >= 50 { // 最多返回50条
				break
			}
		}
	}

	return results, nil
}

// loadAllFunds 加载所有基金列表
func (f *FundAPI) loadAllFunds() error {
	resp, err := client.R().
		Get("http://fund.eastmoney.com/js/fundcode_search.js")
	if err != nil {
		return err
	}

	body := string(resp.Body())

	// 解析 var r = [["000001","HXCZHH","华夏成长混合","混合型-灵活","HUAXIACHENGZHANGHUNHE"],...]
	re := regexp.MustCompile(`\["(\d+)","([^"]+)","([^"]+)","([^"]+)","([^"]+)"\]`)
	matches := re.FindAllStringSubmatch(body, -1)

	f.allFunds = make([]model.FundSearchResult, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 6 {
			f.allFunds = append(f.allFunds, model.FundSearchResult{
				Code:       match[1],
				PinyinAbbr: match[2],
				Name:       match[3],
				Type:       match[4],
				Pinyin:     match[5],
			})
		}
	}

	return nil
}

// GetFundDetail 获取基金详情(实时估值)
func (f *FundAPI) GetFundDetail(code string) (*model.Fund, error) {
	url := fmt.Sprintf("http://fundgz.1234567.com.cn/js/%s.js?rt=%d", code, time.Now().UnixMilli())

	resp, err := client.R().
		SetHeader("Referer", "http://fund.eastmoney.com/").
		Get(url)
	if err != nil {
		return nil, err
	}

	body := string(resp.Body())

	// 解析 jsonpgz({"fundcode":"000001","name":"华夏成长混合",...});
	re := regexp.MustCompile(`jsonpgz\((.*)\)`)
	match := re.FindStringSubmatch(body)
	if len(match) < 2 {
		return nil, fmt.Errorf("无法解析基金数据: %s", code)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(match[1]), &data); err != nil {
		return nil, err
	}

	fund := &model.Fund{
		Code:      code,
		Name:      getString(data, "name"),
		UpdatedAt: time.Now(),
	}

	// 解析净值
	if dwjz, ok := data["dwjz"].(string); ok {
		fund.NetValue, _ = strconv.ParseFloat(dwjz, 64)
	}

	// 解析估值
	if gsz, ok := data["gsz"].(string); ok {
		fund.EstValue, _ = strconv.ParseFloat(gsz, 64)
	}

	// 解析估算涨跌幅
	if gszzl, ok := data["gszzl"].(string); ok {
		fund.EstGrowth, _ = strconv.ParseFloat(gszzl, 64)
	}

	return fund, nil
}

// GetFundNetValue 获取基金净值(历史)
func (f *FundAPI) GetFundNetValue(code string) (*model.Fund, error) {
	url := fmt.Sprintf("https://fundf10.eastmoney.com/F10DataApi.aspx?type=lsjz&code=%s&page=1&per=1", code)

	resp, err := client.R().
		SetHeader("Referer", "https://fundf10.eastmoney.com/").
		Get(url)
	if err != nil {
		return nil, err
	}

	body := string(resp.Body())

	fund := &model.Fund{
		Code:      code,
		UpdatedAt: time.Now(),
	}

	// 解析表格数据
	// <td>2024-01-15</td><td class='tor bold'>1.2345</td><td class='tor bold'>1.5678</td><td class='tor bold red'>1.23%</td>
	reRow := regexp.MustCompile(`<td>(\d{4}-\d{2}-\d{2})</td><td[^>]*>([^<]+)</td><td[^>]*>([^<]+)</td><td[^>]*>([^<]+)</td>`)
	match := reRow.FindStringSubmatch(body)
	if len(match) >= 5 {
		fund.NetValue, _ = strconv.ParseFloat(match[2], 64)
		fund.TotalValue, _ = strconv.ParseFloat(match[3], 64)
		growth := strings.TrimSuffix(match[4], "%")
		fund.DayGrowth, _ = strconv.ParseFloat(growth, 64)
	}

	return fund, nil
}

// GetFundHistory 获取基金历史净值
func (f *FundAPI) GetFundHistory(code string, days int) ([]model.NetValueHistory, error) {
	url := fmt.Sprintf("https://fundf10.eastmoney.com/F10DataApi.aspx?type=lsjz&code=%s&page=1&per=%d", code, days)

	resp, err := client.R().
		SetHeader("Referer", "https://fundf10.eastmoney.com/").
		Get(url)
	if err != nil {
		return nil, err
	}

	body := string(resp.Body())

	var histories []model.NetValueHistory

	reRow := regexp.MustCompile(`<td>(\d{4}-\d{2}-\d{2})</td><td[^>]*>([^<]+)</td><td[^>]*>([^<]+)</td><td[^>]*>([^<]*)</td>`)
	matches := reRow.FindAllStringSubmatch(body, -1)

	for _, match := range matches {
		if len(match) >= 5 {
			date, _ := time.Parse("2006-01-02", match[1])
			netValue, _ := strconv.ParseFloat(match[2], 64)
			totalValue, _ := strconv.ParseFloat(match[3], 64)
			growth := strings.TrimSuffix(strings.TrimSpace(match[4]), "%")
			dayGrowth, _ := strconv.ParseFloat(growth, 64)

			histories = append(histories, model.NetValueHistory{
				FundCode:   code,
				NetValue:   netValue,
				TotalValue: totalValue,
				DayGrowth:  dayGrowth,
				Date:       date,
			})
		}
	}

	return histories, nil
}

// RefreshFund 刷新基金数据
func (f *FundAPI) RefreshFund(code string) (*model.Fund, error) {
	// 获取实时估值
	fund, err := f.GetFundDetail(code)
	if err != nil {
		// 如果获取估值失败，尝试获取历史净值
		fund, err = f.GetFundNetValue(code)
		if err != nil {
			return nil, err
		}
	}

	// 获取基金类型(从搜索列表)
	results, _ := f.SearchFund(code)
	for _, r := range results {
		if r.Code == code {
			fund.Type = r.Type
			break
		}
	}

	// 保存到数据库
	if err := repository.SaveFund(fund); err != nil {
		return nil, err
	}

	return fund, nil
}

// RefreshAllHoldings 刷新所有持仓基金数据
func (f *FundAPI) RefreshAllHoldings() error {
	holdings, err := repository.GetAllHoldings()
	if err != nil {
		return err
	}

	for _, h := range holdings {
		fund, err := f.RefreshFund(h.FundCode)
		if err != nil {
			continue // 跳过失败的
		}

		// 更新持仓的当前净值
		h.CurrentNav = fund.NetValue
		if fund.EstValue > 0 {
			h.CurrentNav = fund.EstValue
		}
		repository.SaveHolding(&h)
	}

	return nil
}

// GetFundRanking 获取基金排行榜
func (f *FundAPI) GetFundRanking(rankType string, limit int) ([]model.FundRanking, error) {
	// 排序字段: zzf-涨幅, 1nzf-1年涨幅, 6yzf-6月涨幅, 3yzf-3月涨幅
	sortField := "zzf"
	sortOrder := "desc"

	switch rankType {
	case "gain":
		sortField = "zzf"
		sortOrder = "desc"
	case "loss":
		sortField = "zzf"
		sortOrder = "asc"
	case "year":
		sortField = "1nzf"
		sortOrder = "desc"
	case "month":
		sortField = "3yzf"
		sortOrder = "desc"
	}

	url := fmt.Sprintf(
		"https://fund.eastmoney.com/data/rankhandler.aspx?op=ph&dt=kf&ft=all&rs=&gs=0&sc=%s&st=%s&pi=1&pn=%d&dx=1",
		sortField, sortOrder, limit,
	)

	resp, err := client.R().
		SetHeader("Referer", "https://fund.eastmoney.com/data/fundranking.html").
		Get(url)
	if err != nil {
		return nil, err
	}

	body := string(resp.Body())

	// 解析数据: var rankData = {datas:["000001,华夏成长混合,HXCZHH,2024-01-15,1.2345,1.5678,1.23,...",...]}
	re := regexp.MustCompile(`"([^"]+)"`)
	matches := re.FindAllStringSubmatch(body, -1)

	var rankings []model.FundRanking
	for i, match := range matches {
		if len(match) < 2 {
			continue
		}
		parts := strings.Split(match[1], ",")
		if len(parts) < 10 {
			continue
		}

		dayGrowth, _ := strconv.ParseFloat(parts[6], 64)

		ranking := model.FundRanking{
			FundCode:     parts[0],
			FundName:     parts[1],
			RankType:     rankType,
			RankValue:    dayGrowth,
			RankPosition: i + 1,
			Period:       "day",
			StatDate:     time.Now(),
		}
		rankings = append(rankings, ranking)

		if len(rankings) >= limit {
			break
		}
	}

	return rankings, nil
}

// GetInstitutionHolding 获取机构持仓数据
func (f *FundAPI) GetInstitutionHolding(code string) (*model.InstitutionHolding, error) {
	url := fmt.Sprintf("https://fundf10.eastmoney.com/FundArchivesDatas.aspx?type=jgcc&code=%s", code)

	resp, err := client.R().
		SetHeader("Referer", "https://fundf10.eastmoney.com/").
		Get(url)
	if err != nil {
		return nil, err
	}

	body := string(resp.Body())

	holding := &model.InstitutionHolding{
		FundCode:   code,
		ReportDate: time.Now(),
	}

	// 解析机构持仓比例
	reInst := regexp.MustCompile(`机构持有比例[^>]*>([0-9.]+)%`)
	if match := reInst.FindStringSubmatch(body); len(match) >= 2 {
		holding.InstitutionRatio, _ = strconv.ParseFloat(match[1], 64)
	}

	// 解析个人持仓比例
	rePerson := regexp.MustCompile(`个人持有比例[^>]*>([0-9.]+)%`)
	if match := rePerson.FindStringSubmatch(body); len(match) >= 2 {
		holding.PersonalRatio, _ = strconv.ParseFloat(match[1], 64)
	}

	// 获取基金名称
	fund, _ := repository.GetFund(code)
	if fund != nil {
		holding.FundName = fund.Name
	}

	return holding, nil
}

// RefreshFundRanking 刷新基金排行数据并保存
func (f *FundAPI) RefreshFundRanking(rankType string, limit int) error {
	rankings, err := f.GetFundRanking(rankType, limit)
	if err != nil {
		return err
	}

	for _, ranking := range rankings {
		if err := repository.SaveFundRanking(&ranking); err != nil {
			continue
		}
	}

	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
