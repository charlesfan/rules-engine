package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charlesfan/rules-engine/internal/rules/dsl"
	"github.com/charlesfan/rules-engine/internal/rules/evaluator"
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

	// 解析 DSL
	p := parser.NewParser()
	ruleSet, err := p.Parse(data)
	if err != nil {
		fmt.Printf("❌ DSL 解析失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ DSL 載入成功")
	fmt.Println("\n" + strings.Repeat("=", 80))

	// 測試案例 1: 1人郵寄組
	fmt.Println("\n📦 測試案例 1: 1人選擇郵寄組")
	testCase1 := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"name":               "張小明",
			"birth_year":         2000,
			"pickup_method":      "mail",
			"tshirt_size":        "M",
			"agree_terms":        true,
			"has_parent_consent": false,
		},
		TeamSize: 1,
		Addons:   map[string]interface{}{},
	}
	runTest(ruleSet, testCase1)

	fmt.Println("\n" + strings.Repeat("=", 80))

	// 測試案例 2: 5人郵寄組
	fmt.Println("\n📦 測試案例 2: 5人選擇郵寄組")
	testCase2 := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"name":               "李大華",
			"birth_year":         1995,
			"pickup_method":      "mail",
			"tshirt_size":        "L",
			"agree_terms":        true,
			"has_parent_consent": false,
		},
		TeamSize: 5,
		Addons:   map[string]interface{}{},
	}
	runTest(ruleSet, testCase2)

	fmt.Println("\n" + strings.Repeat("=", 80))

	// 測試案例 3: 1人現場領物組（不應有運費）
	fmt.Println("\n📦 測試案例 3: 1人選擇現場領物（不應有運費）")
	testCase3 := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"name":               "王小華",
			"birth_year":         2005,
			"pickup_method":      "onsite_1121",
			"tshirt_size":        "S",
			"agree_terms":        true,
			"has_parent_consent": true,
		},
		TeamSize: 1,
		Addons:   map[string]interface{}{},
	}
	runTest(ruleSet, testCase3)

	fmt.Println("\n" + strings.Repeat("=", 80))

	// 測試案例 4: 除錯 - 手動檢查條件
	fmt.Println("\n🔍 Debug: 檢查運費規則條件")
	debugShippingRules(ruleSet, testCase1)
}

func runTest(ruleSet *dsl.RuleSet, ctx *dsl.Context) {
	eval := evaluator.NewEvaluator(ruleSet)
	result, err := eval.Evaluate(ctx)
	if err != nil {
		fmt.Printf("❌ 評估失敗: %v\n", err)
		return
	}

	// 顯示輸入資訊
	fmt.Printf("輸入資訊:\n")
	fmt.Printf("  - 領取方式: %v\n", ctx.User["pickup_method"])
	fmt.Printf("  - 訂單人數: %d\n", ctx.TeamSize)
	fmt.Printf("  - 出生年份: %v\n", ctx.User["birth_year"])

	// 顯示計算結果
	fmt.Printf("\n計算結果:\n")
	fmt.Printf("  - 驗證通過: %v\n", result.Valid)

	if len(result.Errors) > 0 {
		fmt.Printf("  - 錯誤訊息:\n")
		for _, err := range result.Errors {
			fmt.Printf("    ❌ [%s] %s\n", err.RuleID, err.Message)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("  - 警告訊息:\n")
		for _, warn := range result.Warnings {
			fmt.Printf("    ⚠️  [%s] %s\n", warn.RuleID, warn.Message)
		}
	}

	fmt.Printf("\n價格明細:\n")
	if result.Price != nil {
		for itemID, item := range result.Price.Items {
			fmt.Printf("  - [%s] %s\n", itemID, item.Name)
			fmt.Printf("    數量: %d | 單價: %.0f | 小計: %.0f\n",
				item.Quantity, item.UnitPrice, item.FinalPrice)
		}

		fmt.Printf("\n總計:\n")
		fmt.Printf("  - 小計: NT$ %.0f\n", result.Price.SubTotal)
		fmt.Printf("  - 折扣: NT$ -%.0f\n", result.Price.TotalDiscount)
		fmt.Printf("  - 應付金額: NT$ %.0f\n", result.Price.FinalPrice)
	}

	fmt.Printf("\n觸發規則: %v\n", result.AppliedRules)
}

func debugShippingRules(ruleSet *dsl.RuleSet, ctx *dsl.Context) {
	eval := evaluator.NewEvaluator(ruleSet)

	fmt.Printf("Context 內容:\n")
	fmt.Printf("  - user.pickup_method = %v (type: %T)\n", ctx.User["pickup_method"], ctx.User["pickup_method"])
	fmt.Printf("  - team_size = %v (type: %T)\n", ctx.TeamSize, ctx.TeamSize)

	fmt.Printf("\n檢查各運費規則條件:\n")
	for _, rule := range ruleSet.PricingRules {
		if rule.Action.Type == "add_item" && rule.Action.Item == "addon:shipping" {
			matches, err := eval.EvaluateExpression(rule.Condition, ctx)
			fmt.Printf("  [%s] 條件匹配: %v", rule.ID, matches)
			if err != nil {
				fmt.Printf(" (錯誤: %v)", err)
			}
			fmt.Println()

			// 詳細檢查條件
			if rule.Condition.Type == "and" {
				for i, cond := range rule.Condition.Conditions {
					condMatch, condErr := eval.EvaluateExpression(cond, ctx)
					fmt.Printf("    子條件 %d [%s]: %v", i+1, cond.Type, condMatch)
					if cond.Field != "" {
						value, _ := eval.GetFieldValue(cond.Field, ctx)
						fmt.Printf(" (field=%s, value=%v, expect=%v)", cond.Field, value, cond.Value)
					}
					if condErr != nil {
						fmt.Printf(" (錯誤: %v)", condErr)
					}
					fmt.Println()
				}
			}
		}
	}
}
