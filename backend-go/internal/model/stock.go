package model

// ────────────────────────────────────────────────────────────────────
// Stock Detail (from Yahoo Finance)
// ────────────────────────────────────────────────────────────────────

type StockBasic struct {
	Code      string `json:"code"`
	ShortName string `json:"shortName"`
	LongName  string `json:"longName"`
	Sector    string `json:"sector"`
	Industry  string `json:"industry"`
	Website   string `json:"website"`
}

type StockPrice struct {
	CurrentPrice        interface{} `json:"currentPrice"`
	PreviousClose       interface{} `json:"previousClose"`
	Open                interface{} `json:"open"`
	DayHigh             interface{} `json:"dayHigh"`
	DayLow              interface{} `json:"dayLow"`
	Volume              string      `json:"volume"`
	AverageVolume       string      `json:"averageVolume"`
	FiftyTwoWeekHigh    interface{} `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow     interface{} `json:"fiftyTwoWeekLow"`
	FiftyDayAverage     interface{} `json:"fiftyDayAverage"`
	TwoHundredDayAverage interface{} `json:"twoHundredDayAverage"`
	Beta                interface{} `json:"beta"`
}

type StockValuation struct {
	MarketCap       string      `json:"marketCap"`
	EnterpriseValue string      `json:"enterpriseValue"`
	TrailingPE      interface{} `json:"trailingPE"`
	ForwardPE       interface{} `json:"forwardPE"`
	PriceToBook     interface{} `json:"priceToBook"`
	TrailingEps     interface{} `json:"trailingEps"`
	ForwardEps      interface{} `json:"forwardEps"`
}

type StockDividends struct {
	DividendRate  interface{} `json:"dividendRate"`
	DividendYield string      `json:"dividendYield"`
	PayoutRatio   string      `json:"payoutRatio"`
}

type StockOwnership struct {
	SharesOutstanding       string `json:"sharesOutstanding"`
	FloatShares             string `json:"floatShares"`
	HeldPercentInsiders     string `json:"heldPercentInsiders"`
	HeldPercentInstitutions string `json:"heldPercentInstitutions"`
}

type StockProfitability struct {
	GrossMargins     string `json:"grossMargins"`
	OperatingMargins string `json:"operatingMargins"`
	ProfitMargins    string `json:"profitMargins"`
	ReturnOnEquity   string `json:"returnOnEquity"`
	ReturnOnAssets   string `json:"returnOnAssets"`
	RevenueGrowth    string `json:"revenueGrowth"`
	EarningsGrowth   string `json:"earningsGrowth"`
	TotalRevenue     string `json:"totalRevenue"`
	NetIncome        string `json:"netIncome"`
}

type MajorHolder struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type InstitutionalHolder struct {
	Holder       string `json:"holder"`
	Shares       string `json:"shares"`
	DateReported string `json:"dateReported"`
	PctHeld      string `json:"pctHeld"`
	Value        string `json:"value"`
}

type StockDetail struct {
	Basic                StockBasic            `json:"basic"`
	Price                StockPrice            `json:"price"`
	Valuation            StockValuation        `json:"valuation"`
	Dividends            StockDividends        `json:"dividends"`
	Ownership            StockOwnership        `json:"ownership"`
	Profitability        StockProfitability    `json:"profitability"`
	MajorHolders         []MajorHolder         `json:"majorHolders"`
	InstitutionalHolders []InstitutionalHolder `json:"institutionalHolders"`
}

// ────────────────────────────────────────────────────────────────────
// Shareholders Distribution (集保結算所)
// ────────────────────────────────────────────────────────────────────

type ShareholderSummary struct {
	Date         string      `json:"date"`
	TotalShares  string      `json:"totalShares"`
	TotalHolders string      `json:"totalHolders"`
	AvgShares    string      `json:"avgShares"`
	Gt400Shares  string      `json:"gt400Shares"`
	Gt400Pct     string      `json:"gt400Pct"`
	Gt400Count   string      `json:"gt400Count"`
	Range400_600 string      `json:"range400_600"`
	Range600_800 string      `json:"range600_800"`
	Range800_1000 string     `json:"range800_1000"`
	Gt1000Count  string      `json:"gt1000Count"`
	Gt1000Pct    string      `json:"gt1000Pct"`
	ClosePrice   string      `json:"closePrice"`
	PE           interface{} `json:"pe"`
}

type ShareholderPeriod struct {
	Holders string `json:"holders"`
	Shares  string `json:"shares"`
	Pct     string `json:"pct"`
}

type ShareholderDetailRow struct {
	Range   string              `json:"range"`
	Periods []ShareholderPeriod `json:"periods"`
}

type ShareholderDetail struct {
	Dates []string               `json:"dates"`
	Rows  []ShareholderDetailRow `json:"rows"`
}

type ShareholderData struct {
	Code    string               `json:"code"`
	EPS     interface{}          `json:"eps"`
	Summary []ShareholderSummary `json:"summary"`
	Detail  ShareholderDetail    `json:"detail"`
}
