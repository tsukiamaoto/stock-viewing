package main

import (
	"fmt"
	"stock-viewing-backend/internal/translate"
)

func main() {
	input := "摩根大通CEO戴蒙：修订后的美国巴塞尔协议III及全球系统重要性银行附加资本要求提案的某些方面仍“荒谬至极”。"
	result, err := translate.ToTraditionalChinese(input, "zh-CN")
	if err != nil {
		fmt.Printf("Error (zh-CN): %v\n", err)
	} else {
		fmt.Printf("zh-CN Output: %s\n", result)
	}

	resultAuto, err := translate.ToTraditionalChinese(input, "auto")
	if err != nil {
		fmt.Printf("Error (auto): %v\n", err)
	} else {
		fmt.Printf("auto Output: %s\n", resultAuto)
	}
}
