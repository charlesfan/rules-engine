package evaluator

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/charlesfan/rules-engine/internal/rules/calculator"
	"github.com/charlesfan/rules-engine/internal/rules/dsl"
)

// Evaluator 規則評估器
type Evaluator struct {
	ruleSet *dsl.RuleSet
}

// NewEvaluator 建立新的 Evaluator
func NewEvaluator(ruleSet *dsl.RuleSet) *Evaluator {
	return &Evaluator{
		ruleSet: ruleSet,
	}
}

// Evaluate 評估規則集
func (e *Evaluator) Evaluate(ctx *dsl.Context) (*dsl.EvaluationResult, error) {
	result := &dsl.EvaluationResult{
		Valid:        true,
		AppliedRules: []string{},
		Errors:       []dsl.ValidationError{},
		Warnings:     []dsl.ValidationError{},
	}

	// 初始化 Context 的變數和資料源
	if ctx.Variables == nil {
		ctx.Variables = e.ruleSet.Variables
	}
	if ctx.DataSources == nil {
		ctx.DataSources = make(map[string]interface{})
		// 載入資料源（簡化版）
		for name := range e.ruleSet.DataSources {
			ctx.DataSources[name] = []interface{}{}
		}
	}

	// 執行驗證規則
	for _, rule := range e.ruleSet.ValidationRules {
		matches, err := e.evaluateExpression(rule.Condition, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate validation rule %s: %w", rule.ID, err)
		}

		if matches {
			validationError := dsl.ValidationError{
				RuleID:  rule.ID,
				Type:    rule.ErrorType,
				Message: rule.ErrorMessage,
			}

			if rule.ErrorType == "blocking" {
				result.Valid = false
				result.Errors = append(result.Errors, validationError)
			} else {
				result.Warnings = append(result.Warnings, validationError)
			}
		}
	}

	// 計算價格
	calc := calculator.NewCalculator(e.ruleSet)
	priceBreakdown, appliedRules, err := calc.Calculate(ctx, e)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate price: %w", err)
	}

	result.Price = priceBreakdown
	result.AppliedRules = append(result.AppliedRules, appliedRules...)

	return result, nil
}

// EvaluateExpression 實作 calculator.ExpressionEvaluator 介面
func (e *Evaluator) EvaluateExpression(expr *dsl.Expression, ctx *dsl.Context) (bool, error) {
	return e.evaluateExpression(expr, ctx)
}

// GetFieldValue 實作 calculator.ExpressionEvaluator 介面
func (e *Evaluator) GetFieldValue(field string, ctx *dsl.Context) (interface{}, error) {
	return e.getFieldValue(field, ctx)
}

// evaluateExpression 評估表達式
func (e *Evaluator) evaluateExpression(expr *dsl.Expression, ctx *dsl.Context) (bool, error) {
	if expr == nil {
		return false, fmt.Errorf("expression is nil")
	}

	switch expr.Type {
	case "always_true":
		return true, nil

	case "rule_ref":
		// 引用其他規則定義
		ruleDef, exists := e.ruleSet.RuleDefs[expr.Rule]
		if !exists {
			return false, fmt.Errorf("referenced rule not found: %s", expr.Rule)
		}
		return e.evaluateExpression(ruleDef.Expression, ctx)

	case "and":
		for _, cond := range expr.Conditions {
			match, err := e.evaluateExpression(cond, ctx)
			if err != nil {
				return false, err
			}
			if !match {
				return false, nil
			}
		}
		return true, nil

	case "or":
		for _, cond := range expr.Conditions {
			match, err := e.evaluateExpression(cond, ctx)
			if err != nil {
				return false, err
			}
			if match {
				return true, nil
			}
		}
		return false, nil

	case "not":
		match, err := e.evaluateExpression(expr.Condition, ctx)
		if err != nil {
			return false, err
		}
		return !match, nil

	case "equals":
		fieldValue, err := e.getFieldValue(expr.Field, ctx)
		if err != nil {
			return false, nil // 欄位不存在時返回 false
		}
		return e.compareValues(fieldValue, expr.Value, "==")

	case "compare":
		fieldValue, err := e.getFieldValue(expr.Field, ctx)
		if err != nil {
			return false, nil
		}
		return e.compareValues(fieldValue, expr.Value, expr.Operator)

	case "datetime_before":
		fieldValue, err := e.getFieldValue(expr.Field, ctx)
		if err != nil {
			return false, nil
		}
		fieldTime, ok := fieldValue.(time.Time)
		if !ok {
			return false, fmt.Errorf("field %s is not a time.Time", expr.Field)
		}
		compareTime, err := time.Parse(time.RFC3339, expr.Value.(string))
		if err != nil {
			return false, fmt.Errorf("invalid datetime value: %w", err)
		}
		return fieldTime.Before(compareTime), nil

	case "datetime_after":
		fieldValue, err := e.getFieldValue(expr.Field, ctx)
		if err != nil {
			return false, nil
		}
		fieldTime, ok := fieldValue.(time.Time)
		if !ok {
			return false, fmt.Errorf("field %s is not a time.Time", expr.Field)
		}
		compareTime, err := time.Parse(time.RFC3339, expr.Value.(string))
		if err != nil {
			return false, fmt.Errorf("invalid datetime value: %w", err)
		}
		return fieldTime.After(compareTime), nil

	case "datetime_between":
		fieldValue, err := e.getFieldValue(expr.Field, ctx)
		if err != nil {
			return false, nil
		}
		fieldTime, ok := fieldValue.(time.Time)
		if !ok {
			return false, fmt.Errorf("field %s is not a time.Time", expr.Field)
		}
		startTime, err := time.Parse(time.RFC3339, expr.Start.(string))
		if err != nil {
			return false, fmt.Errorf("invalid start datetime value: %w", err)
		}
		endTime, err := time.Parse(time.RFC3339, expr.End.(string))
		if err != nil {
			return false, fmt.Errorf("invalid end datetime value: %w", err)
		}
		return (fieldTime.After(startTime) || fieldTime.Equal(startTime)) &&
			(fieldTime.Before(endTime) || fieldTime.Equal(endTime)), nil

	case "field_exists":
		_, err := e.getFieldValue(expr.Field, ctx)
		return err == nil, nil

	case "field_empty":
		fieldValue, err := e.getFieldValue(expr.Field, ctx)
		if err != nil {
			return true, nil // 欄位不存在視為空
		}
		return e.isEmpty(fieldValue), nil

	case "in_list":
		fieldValue, err := e.getFieldValue(expr.Field, ctx)
		if err != nil {
			return false, nil
		}
		list, err := e.getDataSourceList(expr.List, ctx)
		if err != nil {
			return false, err
		}
		return e.inList(fieldValue, list, expr.MatchField), nil

	case "array_any":
		array, err := e.getFieldValue(expr.Array, ctx)
		if err != nil {
			return false, nil
		}
		arrSlice, ok := array.([]interface{})
		if !ok {
			return false, fmt.Errorf("field %s is not an array", expr.Array)
		}
		for _, item := range arrSlice {
			// 建立臨時上下文，將當前項目加入
			itemCtx := e.createItemContext(ctx, item)
			match, err := e.evaluateExpression(expr.Condition, itemCtx)
			if err != nil {
				return false, err
			}
			if match {
				return true, nil
			}
		}
		return false, nil

	case "array_all":
		array, err := e.getFieldValue(expr.Array, ctx)
		if err != nil {
			return false, nil
		}
		arrSlice, ok := array.([]interface{})
		if !ok {
			return false, fmt.Errorf("field %s is not an array", expr.Array)
		}
		for _, item := range arrSlice {
			itemCtx := e.createItemContext(ctx, item)
			match, err := e.evaluateExpression(expr.Condition, itemCtx)
			if err != nil {
				return false, err
			}
			if !match {
				return false, nil
			}
		}
		return true, nil

	default:
		return false, fmt.Errorf("unknown expression type: %s", expr.Type)
	}
}

// getFieldValue 取得欄位值（支援巢狀欄位，如 user.email, team.members.0.gender）
func (e *Evaluator) getFieldValue(field string, ctx *dsl.Context) (interface{}, error) {
	// 特殊處理
	if field == "register_date" {
		return ctx.RegisterDate, nil
	}
	if field == "team_size" {
		return ctx.TeamSize, nil
	}

	parts := strings.Split(field, ".")
	var current interface{}

	switch parts[0] {
	case "user":
		current = ctx.User
	case "team":
		current = ctx.Team
	case "addons":
		current = ctx.Addons
	case "$computed":
		// 計算欄位
		if len(parts) < 2 {
			return nil, fmt.Errorf("$computed requires field name")
		}
		if ctx.ComputedValues == nil {
			return nil, fmt.Errorf("computed values not initialized")
		}
		value, exists := ctx.ComputedValues[parts[1]]
		if !exists {
			return nil, fmt.Errorf("computed field not found: %s", parts[1])
		}
		return value, nil
	default:
		return nil, fmt.Errorf("unknown root field: %s", parts[0])
	}

	// 遍歷巢狀路徑
	for i := 1; i < len(parts); i++ {
		if current == nil {
			return nil, fmt.Errorf("field not found: %s", field)
		}

		switch v := current.(type) {
		case map[string]interface{}:
			var exists bool
			current, exists = v[parts[i]]
			if !exists {
				return nil, fmt.Errorf("field not found: %s", field)
			}
		default:
			// 使用反射處理結構體
			rv := reflect.ValueOf(current)
			if rv.Kind() == reflect.Ptr {
				rv = rv.Elem()
			}
			if rv.Kind() == reflect.Map {
				key := reflect.ValueOf(parts[i])
				value := rv.MapIndex(key)
				if !value.IsValid() {
					return nil, fmt.Errorf("field not found: %s", field)
				}
				current = value.Interface()
			} else {
				return nil, fmt.Errorf("cannot access field: %s", field)
			}
		}
	}

	return current, nil
}

// compareValues 比較兩個值
func (e *Evaluator) compareValues(a, b interface{}, operator string) (bool, error) {
	switch operator {
	case "==", "!=":
		// 嘗試數字比較（處理 int vs float64 的問題）
		aFloat, aOk := e.toFloat64(a)
		bFloat, bOk := e.toFloat64(b)
		if aOk && bOk {
			// 兩者都是數字，比較數值
			if operator == "==" {
				return aFloat == bFloat, nil
			}
			return aFloat != bFloat, nil
		}
		// 非數字，使用 DeepEqual
		if operator == "==" {
			return reflect.DeepEqual(a, b), nil
		}
		return !reflect.DeepEqual(a, b), nil
	case ">", "<", ">=", "<=":
		return e.compareNumbers(a, b, operator)
	default:
		return false, fmt.Errorf("unknown operator: %s", operator)
	}
}

// compareNumbers 比較數字
func (e *Evaluator) compareNumbers(a, b interface{}, operator string) (bool, error) {
	aFloat, aOk := e.toFloat64(a)
	bFloat, bOk := e.toFloat64(b)

	if !aOk || !bOk {
		return false, fmt.Errorf("cannot compare non-numeric values")
	}

	switch operator {
	case ">":
		return aFloat > bFloat, nil
	case "<":
		return aFloat < bFloat, nil
	case ">=":
		return aFloat >= bFloat, nil
	case "<=":
		return aFloat <= bFloat, nil
	default:
		return false, fmt.Errorf("unknown operator: %s", operator)
	}
}

// toFloat64 轉換為 float64
func (e *Evaluator) toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}

// isEmpty 檢查值是否為空
func (e *Evaluator) isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		return rv.Len() == 0
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() == 0
	case reflect.Ptr:
		return rv.IsNil()
	default:
		return false
	}
}

// getDataSourceList 取得資料源列表
func (e *Evaluator) getDataSourceList(listRef string, ctx *dsl.Context) ([]interface{}, error) {
	// 假設格式為 $data_sources.car_owners
	if !strings.HasPrefix(listRef, "$data_sources.") {
		return nil, fmt.Errorf("invalid data source reference: %s", listRef)
	}

	name := strings.TrimPrefix(listRef, "$data_sources.")
	list, exists := ctx.DataSources[name]
	if !exists {
		return nil, fmt.Errorf("data source not found: %s", name)
	}

	listSlice, ok := list.([]interface{})
	if !ok {
		return nil, fmt.Errorf("data source %s is not an array", name)
	}

	return listSlice, nil
}

// inList 檢查值是否在列表中
func (e *Evaluator) inList(value interface{}, list []interface{}, matchField string) bool {
	for _, item := range list {
		if matchField != "" {
			// 比對特定欄位
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if itemMap[matchField] == value {
				return true
			}
		} else {
			// 直接比對
			if reflect.DeepEqual(item, value) {
				return true
			}
		}
	}
	return false
}

// createItemContext 建立項目上下文（用於陣列操作）
func (e *Evaluator) createItemContext(parentCtx *dsl.Context, item interface{}) *dsl.Context {
	// 簡化版：直接將 item 作為 user
	itemMap, ok := item.(map[string]interface{})
	if !ok {
		itemMap = map[string]interface{}{}
	}

	return &dsl.Context{
		RegisterDate: parentCtx.RegisterDate,
		User:         itemMap,
		Team:         parentCtx.Team,
		TeamSize:     parentCtx.TeamSize,
		Addons:       parentCtx.Addons,
		Variables:    parentCtx.Variables,
		DataSources:  parentCtx.DataSources,
	}
}
