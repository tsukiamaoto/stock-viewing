package service

import (
	"fmt"
	"strings"

	"stock-viewing-backend/internal/crawler"
	"stock-viewing-backend/internal/model"

	"github.com/PuerkitoBio/goquery"
)

// ────────────────────────────────────────────────────────────────────
// Stock Detail Service (Yahoo Finance fundamentals)
// ────────────────────────────────────────────────────────────────────

func GetStockDetail(code string) (*model.StockDetail, error) {
	tickerSym := crawler.ToYahooSymbol(code)
	info, err := crawler.FetchYahooQuoteSummary(tickerSym)
	if err != nil {
		return nil, err
	}
	if info.ShortName == "--" || info.ShortName == "" {
		return nil, fmt.Errorf("找不到股票代碼 %s", code)
	}

	detail := &model.StockDetail{
		Basic: model.StockBasic{
			Code:      code,
			ShortName: info.ShortName,
			LongName:  info.LongName,
			Sector:    info.Sector,
			Industry:  info.Industry,
			Website:   info.Website,
		},
		Price: model.StockPrice{
			CurrentPrice:         orDash(info.CurrentPrice),
			PreviousClose:        orDash(info.PreviousClose),
			Open:                 orDash(info.Open),
			DayHigh:              orDash(info.DayHigh),
			DayLow:               orDash(info.DayLow),
			Volume:               crawler.FmtNumber(info.Volume),
			AverageVolume:        crawler.FmtNumber(info.AverageVolume),
			FiftyTwoWeekHigh:     orDash(info.FiftyTwoWeekHigh),
			FiftyTwoWeekLow:      orDash(info.FiftyTwoWeekLow),
			FiftyDayAverage:      orDash(info.FiftyDayAverage),
			TwoHundredDayAverage: orDash(info.TwoHundredDayAverage),
			Beta:                 orDash(info.Beta),
		},
		Valuation: model.StockValuation{
			MarketCap:       crawler.FmtNumber(info.MarketCap),
			EnterpriseValue: crawler.FmtNumber(info.EnterpriseValue),
			TrailingPE:      crawler.FmtRound(info.TrailingPE),
			ForwardPE:       crawler.FmtRound(info.ForwardPE),
			PriceToBook:     crawler.FmtRound(info.PriceToBook),
			TrailingEps:     orDash(info.TrailingEps),
			ForwardEps:      crawler.FmtRound(info.ForwardEps),
		},
		Dividends: model.StockDividends{
			DividendRate:  orDash(info.DividendRate),
			DividendYield: crawler.FmtPct(info.DividendYield),
			PayoutRatio:   crawler.FmtPct(info.PayoutRatio),
		},
		Ownership: model.StockOwnership{
			SharesOutstanding:       crawler.FmtNumber(info.SharesOutstanding),
			FloatShares:             crawler.FmtNumber(info.FloatShares),
			HeldPercentInsiders:     crawler.FmtPct(info.HeldPercentInsiders),
			HeldPercentInstitutions: crawler.FmtPct(info.HeldPercentInstitutions),
		},
		Profitability: model.StockProfitability{
			GrossMargins:     crawler.FmtPct(info.GrossMargins),
			OperatingMargins: crawler.FmtPct(info.OperatingMargins),
			ProfitMargins:    crawler.FmtPct(info.ProfitMargins),
			ReturnOnEquity:   crawler.FmtPct(info.ReturnOnEquity),
			ReturnOnAssets:   crawler.FmtPct(info.ReturnOnAssets),
			RevenueGrowth:    crawler.FmtPct(info.RevenueGrowth),
			EarningsGrowth:   crawler.FmtPct(info.EarningsQuarterlyGrowth),
			TotalRevenue:     crawler.FmtNumber(info.TotalRevenue),
			NetIncome:        crawler.FmtNumber(info.NetIncomeToCommon),
		},
		MajorHolders:         []model.MajorHolder{},
		InstitutionalHolders: []model.InstitutionalHolder{},
	}

	return detail, nil
}

func orDash(v *float64) interface{} {
	if v == nil {
		return "--"
	}
	return *v
}

// ────────────────────────────────────────────────────────────────────
// Shareholders Distribution (集保結算所)
// ────────────────────────────────────────────────────────────────────

const shareholderURL = "https://norway.twsthr.info/StockHolders.aspx?stock=%s"

func GetShareholderDistribution(code string) (*model.ShareholderData, error) {
	url := fmt.Sprintf(shareholderURL, code)
	body, err := crawler.FetchURL(url, nil)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	summary := parseSummaryTable(doc)
	detail := parseDetailTable(doc)

	// Fetch EPS for PE calculation
	var eps *float64
	tickerSym := crawler.ToYahooSymbol(code)
	info, err := crawler.FetchYahooQuoteSummary(tickerSym)
	if err == nil && info.TrailingEps != nil {
		eps = info.TrailingEps
	}

	// Add PE to each summary row
	for i := range summary {
		summary[i].PE = calculatePE(summary[i].ClosePrice, eps)
	}

	// Limit to ~1 year of weekly data
	if len(summary) > 52 {
		summary = summary[:52]
	}

	var epsVal interface{} = nil
	if eps != nil {
		epsVal = *eps
	}

	return &model.ShareholderData{
		Code:    code,
		EPS:     epsVal,
		Summary: summary,
		Detail:  detail,
	}, nil
}

func calculatePE(closePriceStr string, eps *float64) interface{} {
	if eps == nil || *eps <= 0 {
		return "--"
	}
	var price float64
	if _, err := fmt.Sscanf(closePriceStr, "%f", &price); err != nil {
		return "--"
	}
	pe := price / *eps
	return fmt.Sprintf("%.2f", pe)
}

func parseSummaryTable(doc *goquery.Document) []model.ShareholderSummary {
	var results []model.ShareholderSummary

	var tbl *goquery.Selection
	doc.Find("table").Each(func(_ int, s *goquery.Selection) {
		id, _ := s.Attr("id")
		if id == "Details" && s.Find("tr").Length() > 100 {
			tbl = s
		}
	})
	if tbl == nil {
		return results
	}

	tbl.Find("tr").Each(func(_ int, row *goquery.Selection) {
		var cells []string
		row.Find("td, th").Each(func(_ int, cell *goquery.Selection) {
			cells = append(cells, strings.TrimSpace(cell.Text()))
		})
		// Filter empty cells
		var cleaned []string
		for _, c := range cells {
			if c != "" {
				cleaned = append(cleaned, c)
			}
		}
		if len(cleaned) < 10 {
			return
		}
		if cleaned[0] == "資料日期" {
			return
		}
		if len(cleaned[0]) == 8 && isDigits(cleaned[0]) {
			s := model.ShareholderSummary{
				Date:          fmt.Sprintf("%s/%s/%s", cleaned[0][:4], cleaned[0][4:6], cleaned[0][6:]),
				TotalShares:  safeGet(cleaned, 1),
				TotalHolders: safeGet(cleaned, 2),
				AvgShares:    safeGet(cleaned, 3),
				Gt400Shares:  safeGet(cleaned, 4),
				Gt400Pct:     safeGet(cleaned, 5),
				Gt400Count:   safeGet(cleaned, 6),
				Range400_600: safeGet(cleaned, 7),
				Range600_800: safeGet(cleaned, 8),
				Range800_1000: safeGet(cleaned, 9),
				Gt1000Count:  safeGet(cleaned, 10),
				Gt1000Pct:    safeGet(cleaned, 11),
				ClosePrice:   safeGet(cleaned, 12),
			}
			results = append(results, s)
		}
	})
	return results
}

func parseDetailTable(doc *goquery.Document) model.ShareholderDetail {
	result := model.ShareholderDetail{Dates: []string{}, Rows: []model.ShareholderDetailRow{}}

	var tbl *goquery.Selection
	doc.Find("table").Each(func(_ int, s *goquery.Selection) {
		id, _ := s.Attr("id")
		if id == "details" {
			tbl = s
		}
	})
	if tbl == nil {
		return result
	}

	rows := tbl.Find("tr")
	if rows.Length() < 3 {
		return result
	}

	// Row 1: dates
	rows.Eq(1).Find("td, th").Each(func(_ int, cell *goquery.Selection) {
		text := strings.TrimSpace(cell.Text())
		if len(text) == 8 && isDigits(text) {
			result.Dates = append(result.Dates, fmt.Sprintf("%s/%s/%s", text[:4], text[4:6], text[6:]))
		}
	})

	// Rows 3+: data
	rows.Each(func(i int, row *goquery.Selection) {
		if i < 3 {
			return
		}
		var cells []string
		row.Find("td, th").Each(func(_ int, cell *goquery.Selection) {
			cells = append(cells, strings.TrimSpace(cell.Text()))
		})
		var cleaned []string
		for _, c := range cells {
			if c != "" {
				cleaned = append(cleaned, c)
			}
		}
		if len(cleaned) == 0 {
			return
		}

		label := cleaned[0]
		var periods []model.ShareholderPeriod
		idx := 1
		for range result.Dates {
			p := model.ShareholderPeriod{
				Holders: safeGet(cleaned, idx),
				Shares:  safeGet(cleaned, idx+1),
				Pct:     safeGet(cleaned, idx+2),
			}
			periods = append(periods, p)
			idx += 3
		}

		result.Rows = append(result.Rows, model.ShareholderDetailRow{
			Range:   label,
			Periods: periods,
		})
	})

	return result
}

func isDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func safeGet(s []string, i int) string {
	if i < len(s) {
		return s[i]
	}
	return "--"
}
