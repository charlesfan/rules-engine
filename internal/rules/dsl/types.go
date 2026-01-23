package dsl

import "time"

// RuleSet 完整的規則集
type RuleSet struct {
	EventID          string                    `json:"event_id"`
	Version          string                    `json:"version"`
	Name             string                    `json:"name"`
	DataSources      map[string]*DataSource    `json:"data_sources,omitempty"`
	Variables        map[string]interface{}    `json:"variables,omitempty"`
	RuleDefs         map[string]*RuleDef       `json:"rule_definitions,omitempty"`
	ComputedFields   map[string]*ComputedField `json:"computed_fields,omitempty"` // 計算欄位定義
	PricingRules     []*PricingRule            `json:"pricing_rules"`
	ValidationRules  []*ValidationRule         `json:"validation_rules"`
	DiscountStacking *DiscountStacking         `json:"discount_stacking,omitempty"`
	FormSchema       *FormSchema               `json:"form_schema,omitempty"` // 表單定義
}

// DataSource 外部資料源定義
type DataSource struct {
	Type        string   `json:"type"` // external_list, api, csv_upload
	Description string   `json:"description"`
	Source      string   `json:"source"` // api, csv_upload
	Endpoint    string   `json:"endpoint,omitempty"`
	Fields      []string `json:"fields"`
	CacheTTL    int      `json:"cache_ttl"` // 秒
}

// RuleDef 可重用的規則定義
type RuleDef struct {
	Type        string      `json:"type"` // condition
	Description string      `json:"description"`
	Expression  *Expression `json:"expression"`
}

// Expression 條件表達式
type Expression struct {
	Type       string        `json:"type"` // and, or, not, equals, compare, datetime_before, datetime_after, datetime_between, etc.
	Field      string        `json:"field,omitempty"`
	Operator   string        `json:"operator,omitempty"`
	Value      interface{}   `json:"value,omitempty"`
	Start      interface{}   `json:"start,omitempty"` // for datetime_between
	End        interface{}   `json:"end,omitempty"`   // for datetime_between
	Conditions []*Expression `json:"conditions,omitempty"`
	Condition  *Expression   `json:"condition,omitempty"` // for not, array operations
	Array      string        `json:"array,omitempty"`
	List       string        `json:"list,omitempty"`
	MatchField string        `json:"match_field,omitempty"`
	Rule       string        `json:"rule,omitempty"` // for rule_ref
}

// PricingRule 價格規則
type PricingRule struct {
	ID          string      `json:"id"`
	Priority    int         `json:"priority"`
	Description string      `json:"description"`
	Condition   *Expression `json:"condition"`
	Action      *Action     `json:"action"`
}

// ValidationRule 驗證規則
type ValidationRule struct {
	ID           string      `json:"id"`
	Description  string      `json:"description"`
	Condition    *Expression `json:"condition"`
	ErrorType    string      `json:"error_type"` // blocking, warning
	ErrorMessage string      `json:"error_message"`
}

// Action 規則動作
type Action struct {
	Type          string      `json:"type"`               // set_price, add_item, percentage_discount, fixed_discount, price_cap, replace_price
	Item          string      `json:"item,omitempty"`     // registration_fee, addon:tshirt, etc.
	ApplyTo       []string    `json:"apply_to,omitempty"` // 折扣作用目標（可多個）
	Value         interface{} `json:"value,omitempty"`
	UnitPrice     interface{} `json:"unit_price,omitempty"`  // 支援數字或變數引用 (如 "$variables.price")
	FixedPrice    interface{} `json:"fixed_price,omitempty"` // 固定價格（不乘以數量），支援數字或變數引用
	QuantityField string      `json:"quantity_field,omitempty"`
	Label         string      `json:"label,omitempty"`
}

// DiscountStacking 折扣堆疊方式
type DiscountStacking struct {
	Mode        string `json:"mode"` // multiplicative, additive, best_only
	Description string `json:"description"`
}

// Context 執行上下文
type Context struct {
	RegisterDate   time.Time              `json:"register_date"`
	User           map[string]interface{} `json:"user"`
	Team           map[string]interface{} `json:"team,omitempty"`
	TeamSize       int                    `json:"team_size,omitempty"`
	Addons         map[string]interface{} `json:"addons,omitempty"`
	Variables      map[string]interface{} `json:"variables,omitempty"`
	DataSources    map[string]interface{} `json:"data_sources,omitempty"`
	ComputedValues map[string]interface{} `json:"computed_values,omitempty"` // 計算欄位的值
}

// EvaluationResult 評估結果
type EvaluationResult struct {
	Valid        bool              `json:"valid"`
	Price        *PriceBreakdown   `json:"price"`
	Errors       []ValidationError `json:"errors,omitempty"`
	Warnings     []ValidationError `json:"warnings,omitempty"`
	AppliedRules []string          `json:"applied_rules"`
}

// PriceBreakdown 價格明細
type PriceBreakdown struct {
	Items         map[string]*PriceItem `json:"items"` // 所有項目 (registration_fee, addon:tshirt, etc.)
	Discounts     []DiscountItem        `json:"discounts"`
	SubTotal      float64               `json:"sub_total"`
	TotalDiscount float64               `json:"total_discount"`
	FinalPrice    float64               `json:"final_price"`
}

// PriceItem 價格項目
type PriceItem struct {
	ID              string  `json:"id"` // registration_fee, addon:tshirt, etc.
	Name            string  `json:"name"`
	Quantity        int     `json:"quantity"`
	UnitPrice       float64 `json:"unit_price"`
	OriginalPrice   float64 `json:"original_price"`   // 原價（未折扣前）
	DiscountedPrice float64 `json:"discounted_price"` // 折扣後價格
	FinalPrice      float64 `json:"final_price"`      // 最終價格
}

// DiscountItem 折扣項目
type DiscountItem struct {
	RuleID    string  `json:"rule_id"`
	Label     string  `json:"label"`
	Type      string  `json:"type"` // percentage, fixed
	Value     float64 `json:"value"`
	Amount    float64 `json:"amount"`
	AppliedTo string  `json:"applied_to"`
}

// ValidationError 驗證錯誤
type ValidationError struct {
	RuleID  string `json:"rule_id"`
	Type    string `json:"type"` // blocking, warning
	Message string `json:"message"`
}

// FormSchema 表單定義
type FormSchema struct {
	Fields []FormField `json:"fields"` // 表單欄位列表
}

// FormField 表單欄位
type FormField struct {
	ID           string                 `json:"id"`                      // 欄位 ID (e.g., "race_type", "age")
	Label        string                 `json:"label"`                   // 顯示標籤
	Type         string                 `json:"type"`                    // 欄位類型: text, number, select, radio, checkbox, etc.
	Field        string                 `json:"field"`                   // 對應到 Context 的欄位路徑 (e.g., "user.race_type")
	Required     bool                   `json:"required,omitempty"`      // 是否必填
	Options      []FormFieldOption      `json:"options,omitempty"`       // 選項（用於 select, radio）
	Placeholder  string                 `json:"placeholder,omitempty"`   // 提示文字
	DefaultValue interface{}            `json:"default_value,omitempty"` // 預設值
	Min          *int                   `json:"min,omitempty"`           // 最小值（用於 number）
	Max          *int                   `json:"max,omitempty"`           // 最大值（用於 number）
	Validation   map[string]interface{} `json:"validation,omitempty"`    // 額外驗證規則
}

// FormFieldOption 表單欄位選項
type FormFieldOption struct {
	Label string      `json:"label"` // 顯示文字
	Value interface{} `json:"value"` // 實際值
}

// ComputedField 計算欄位定義
type ComputedField struct {
	Type        string            `json:"type"`                  // sum_prices, count_items, item_price, count_array_field, sum_array_field, arithmetic, conditional, etc.
	Description string            `json:"description,omitempty"` // 欄位說明
	Items       []string          `json:"items,omitempty"`       // 用於 sum_prices, count_items（支援萬用字元）
	Item        string            `json:"item,omitempty"`        // 用於 item_price（單一項目）
	Pattern     string            `json:"pattern,omitempty"`     // 用於 count_items 的模式匹配
	Array       string            `json:"array,omitempty"`       // 用於 count_array_field, sum_array_field 指定陣列路徑 (e.g., "team.members")
	Field       string            `json:"field,omitempty"`       // 用於 sum_field, count_array_field, sum_array_field 指定欄位路徑（支援嵌套，如 "addons.water_bottle.quantity"）
	Value       interface{}       `json:"value,omitempty"`       // 用於 count_array_field 指定要匹配的值
	Operation   string            `json:"operation,omitempty"`   // 用於 arithmetic: add, subtract, multiply, divide
	Operands    []ComputedOperand `json:"operands,omitempty"`    // 用於 arithmetic 的運算元
	Condition   *Expression       `json:"condition,omitempty"`   // 用於 conditional
	TrueValue   interface{}       `json:"true_value,omitempty"`  // 用於 conditional
	FalseValue  interface{}       `json:"false_value,omitempty"` // 用於 conditional
	Parts       []interface{}     `json:"parts,omitempty"`       // 用於 concat 字串組合
	From        interface{}       `json:"from,omitempty"`        // 用於 date_diff
	To          interface{}       `json:"to,omitempty"`          // 用於 date_diff
	Unit        string            `json:"unit,omitempty"`        // 用於 date_diff: days, hours, minutes
}

// ComputedOperand 計算運算元（可以是常數或欄位參考）
type ComputedOperand struct {
	Field string      `json:"field,omitempty"` // 欄位參考 (e.g., "$computed.subtotal", "user.age")
	Value interface{} `json:"value,omitempty"` // 常數值
}
