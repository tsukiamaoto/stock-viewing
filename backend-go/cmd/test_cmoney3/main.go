package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Add timeout
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var res string
	// Run task
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.cmoney.tw/forum/stock/1711"),
		// Wait for the main list to load (cmoney uses custom Vue/Nuxt elements, often article or div with a specific class)
		// We'll wait for a known string or just sleep for 3 sec, then dump body
		chromedp.Sleep(3*time.Second),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('article, .article-card, [data-v-2a943070], [class*="article"]')).map(e => e.innerText).filter(t => t.length > 20).join('\n---\n')
		`, &res),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Result length:", len(res))
	if len(res) > 200 {
		fmt.Println("Result snippet:", res[:200])
	} else {
		fmt.Println("Result:", res)
	}
}
