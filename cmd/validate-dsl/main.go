package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charlesfan/rules-engine/internal/rules/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <dsl-file.json>")
		os.Exit(1)
	}

	filePath := os.Args[1]

	// 讀取 DSL 檔案
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("❌ 無法讀取檔案: %v\n", err)
		os.Exit(1)
	}

	// 檢查是否為有效的 JSON
	var jsonCheck interface{}
	if err := json.Unmarshal(data, &jsonCheck); err != nil {
		fmt.Printf("❌ JSON 格式錯誤: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ JSON 格式正確")

	// 使用 Parser 解析
	p := parser.NewParser()
	ruleSet, err := p.Parse(data)
	if err != nil {
		fmt.Printf("❌ DSL 解析失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ DSL 語法驗證通過")
	fmt.Printf("\n📋 規則集資訊:\n")
	fmt.Printf("   - Event ID: %s\n", ruleSet.EventID)
	fmt.Printf("   - 活動名稱: %s\n", ruleSet.Name)
	fmt.Printf("   - 版本: %s\n", ruleSet.Version)
	fmt.Printf("   - 變數數量: %d\n", len(ruleSet.Variables))
	fmt.Printf("   - 計算欄位: %d\n", len(ruleSet.ComputedFields))
	fmt.Printf("   - 定價規則: %d\n", len(ruleSet.PricingRules))
	fmt.Printf("   - 驗證規則: %d\n", len(ruleSet.ValidationRules))
	fmt.Printf("   - 表單欄位: %d\n", len(ruleSet.FormSchema.Fields))

	fmt.Printf("\n📝 定價規則列表:\n")
	for _, rule := range ruleSet.PricingRules {
		fmt.Printf("   [P%d] %s - %s\n", rule.Priority, rule.ID, rule.Description)
	}

	fmt.Printf("\n🔍 驗證規則列表:\n")
	for _, rule := range ruleSet.ValidationRules {
		fmt.Printf("   [%s] %s - %s\n", rule.ErrorType, rule.ID, rule.Description)
	}

	fmt.Println("\n✅ 所有驗證通過！")
}
