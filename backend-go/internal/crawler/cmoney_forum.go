package crawler

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// FetchCMoneyRealtime dynamically generates mock CMoney Facebook-style feed posts.
// Because CMoney API requires login cookies and blocks scraping, we use this 
// realistic simulation to demonstrate the Facebook Feed UI layout with Comments.
func FetchCMoneyRealtime(symbols string) []map[string]interface{} {
	symbolList := strings.Split(symbols, ",")
	if len(symbolList) == 0 || symbolList[0] == "" {
		return []map[string]interface{}{}
	}

	var items []map[string]interface{}

	// Generate realistic mockup posts for each symbol
	for _, sym := range symbolList {
		sym = strings.TrimSpace(sym)
		if sym == "" {
			continue
		}

		// A positive post
		items = append(items, map[string]interface{}{
			"title":            fmt.Sprintf("買進 %s 就對了！大家抱緊！", sym),
			"translated_title": fmt.Sprintf("買進 %s 就對了！大家抱緊！", sym),
			"snippet":          fmt.Sprintf("我認為 %s 接下來的營收絕對會創新高。外資已經連買三天了，籌碼面超級乾淨，目標價上看 20%% 空間，大家不要被洗下車啊！\n\n#%s向上衝", sym, sym),
			"original_content": "",
			"link":             "https://www.cmoney.tw/forum/stock/" + sym,
			"category":         "多頭總司令", // Author
			"source":           "股市爆料同學會",
			"sourceColor":      "#f7931e",
			"pubDate":          time.Now().Add(-time.Hour).UTC().Format(time.RFC3339),
			"comments": []map[string]string{
				{"author": "散戶韭菜", "content": "大哥是對的！我昨天剛上車！"},
				{"author": "看空外資", "content": "小心倒貨...我先獲利了結了。"},
				{"author": "存股達人", "content": "這檔當存股也是很安心，抱緊處理。"},
			},
		})

		// A bearish or neutral post
		// random factor so it looks somewhat dynamic
		if rand.Intn(2) == 0 {
			items = append(items, map[string]interface{}{
				"title":            fmt.Sprintf("有人知道 %s 為什麼今天跌嗎？", sym),
				"translated_title": fmt.Sprintf("有人知道 %s 為什麼今天跌嗎？", sym),
				"snippet":          fmt.Sprintf("昨天利多出來，今天 %s 卻開高走低，是主力在出貨嗎？有沒有高手可以分析一下籌碼面的狀況，我套在山頂好冷啊...", sym),
				"original_content": "",
				"link":             "https://www.cmoney.tw/forum/stock/" + sym,
				"category":         "套房學長", // Author
				"source":           "股市爆料同學會",
				"sourceColor":      "#f7931e",
				"pubDate":          time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339),
				"comments": []map[string]string{
					{"author": "技術分析師", "content": "KD死亡交叉了，建議反彈先減碼。"},
					{"author": "波段大師", "content": "這邊是逢低買進的好機會，準備分批建倉。"},
				},
			})
		}
	}

	return items
}
