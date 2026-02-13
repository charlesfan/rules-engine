package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/charlesfan/rules-engine/internal/api"
	"github.com/charlesfan/rules-engine/internal/config"
	"github.com/charlesfan/rules-engine/internal/rules/dsl"
	"github.com/charlesfan/rules-engine/internal/rules/evaluator"
	"github.com/charlesfan/rules-engine/internal/rules/parser"
	"github.com/charlesfan/rules-engine/internal/store"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	ctx := context.Background()
	eventStore, err := store.NewPostgresStore(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v", err)
		log.Println("Running without database support. Events API will not be available.")
		eventStore = nil
	} else {
		defer eventStore.Close()
		log.Println("Connected to PostgreSQL database")
	}

	r := gin.Default()

	// Static files
	r.Static("/static", "./cmd/rules-engine-demo/static")

	// Home page - visual rule builder
	r.GET("/", func(c *gin.Context) {
		c.File("./cmd/rules-engine-demo/static/builder.html")
	})

	// JSON editor
	r.GET("/editor", func(c *gin.Context) {
		c.File("./cmd/rules-engine-demo/static/index.html")
	})

	// Preview page - dynamic registration form
	r.GET("/preview", func(c *gin.Context) {
		c.File("./cmd/rules-engine-demo/static/preview-dynamic.html")
	})

	// Register Events API routes (if database is available)
	if eventStore != nil {
		api.RegisterRoutes(r, eventStore)
		log.Println("Events API registered at /api/events")
	}

	// API: Validate rules
	r.POST("/api/rules/validate", func(c *gin.Context) {
		var ruleSetData map[string]interface{}
		if err := c.ShouldBindJSON(&ruleSetData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Convert back to JSON bytes
		data, _ := json.Marshal(ruleSetData)

		p := parser.NewParser()
		ruleSet, err := p.Parse(data)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"valid": false,
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"valid":    true,
			"rule_set": ruleSet,
		})
	})

	// API: Evaluate rules (calculate price and validate)
	r.POST("/api/rules/evaluate", func(c *gin.Context) {
		var req struct {
			RuleSet map[string]interface{} `json:"rule_set"`
			Context *dsl.Context           `json:"context"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse rule set
		data, _ := json.Marshal(req.RuleSet)
		p := parser.NewParser()
		ruleSet, err := p.Parse(data)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Evaluate
		ev := evaluator.NewEvaluator(ruleSet)
		result, err := ev.Evaluate(req.Context)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	// API: Get example rules
	r.GET("/api/rules/examples", func(c *gin.Context) {
		examples := getExampleRules()
		c.JSON(http.StatusOK, examples)
	})

	log.Printf("Server starting on http://localhost:%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}

// getExampleRules returns example rules
func getExampleRules() []map[string]interface{} {
	examples := []map[string]interface{}{
		{
			"name":        "車廠路跑活動範例",
			"description": "包含車主優惠、早鳥優惠、團體優惠",
			"rule_set": map[string]interface{}{
				"event_id": "car-brand-marathon-2025",
				"version":  "1.0",
				"name":     "某某車廠路跑活動",
				"data_sources": map[string]interface{}{
					"car_owners": map[string]interface{}{
						"type":        "external_list",
						"description": "車廠提供的車主名單",
						"source":      "csv_upload",
						"fields":      []string{"email", "phone", "purchase_date", "car_model"},
						"cache_ttl":   86400,
					},
				},
				"variables": map[string]interface{}{
					"base_registration_fee": 1200,
					"tshirt_price":          300,
					"medal_upgrade_price":   500,
				},
				"rule_definitions": map[string]interface{}{
					"is_car_owner": map[string]interface{}{
						"type":        "condition",
						"description": "是否為車主",
						"expression": map[string]interface{}{
							"type":        "in_list",
							"field":       "user.email",
							"list":        "$data_sources.car_owners",
							"match_field": "email",
						},
					},
					"is_early_bird": map[string]interface{}{
						"type":        "condition",
						"description": "是否為早鳥",
						"expression": map[string]interface{}{
							"type":  "datetime_before",
							"field": "register_date",
							"value": "2025-11-01T00:00:00Z",
						},
					},
					"is_team": map[string]interface{}{
						"type":        "condition",
						"description": "是否為團體報名",
						"expression": map[string]interface{}{
							"type":     "compare",
							"field":    "team_size",
							"operator": ">=",
							"value":    3,
						},
					},
				},
				"pricing_rules": []interface{}{
					map[string]interface{}{
						"id":          "base_price",
						"priority":    0,
						"description": "基礎報名費",
						"condition": map[string]interface{}{
							"type": "always_true",
						},
						"action": map[string]interface{}{
							"type":   "set_price",
							"target": "registration_fee",
							"value":  "$variables.base_registration_fee",
						},
					},
					map[string]interface{}{
						"id":          "car_owner_registration_discount",
						"priority":    10,
						"description": "車主報名優惠 8 折",
						"condition": map[string]interface{}{
							"type": "rule_ref",
							"rule": "is_car_owner",
						},
						"action": map[string]interface{}{
							"type":   "percentage_discount",
							"target": "registration_fee",
							"value":  20,
							"label":  "車主優惠",
						},
					},
					map[string]interface{}{
						"id":          "early_bird_discount",
						"priority":    20,
						"description": "早鳥優惠 9 折",
						"condition": map[string]interface{}{
							"type": "rule_ref",
							"rule": "is_early_bird",
						},
						"action": map[string]interface{}{
							"type":   "percentage_discount",
							"target": "registration_fee",
							"value":  10,
							"label":  "早鳥優惠",
						},
					},
					map[string]interface{}{
						"id":          "team_discount",
						"priority":    30,
						"description": "團體優惠",
						"condition": map[string]interface{}{
							"type": "rule_ref",
							"rule": "is_team",
						},
						"action": map[string]interface{}{
							"type":   "percentage_discount",
							"target": "registration_fee",
							"value":  15,
							"label":  "團體優惠",
						},
					},
				},
				"validation_rules": []interface{}{
					map[string]interface{}{
						"id":          "min_team_size",
						"description": "團體報名至少 3 人",
						"condition": map[string]interface{}{
							"type": "and",
							"conditions": []interface{}{
								map[string]interface{}{
									"type":  "field_exists",
									"field": "team",
								},
								map[string]interface{}{
									"type":     "compare",
									"field":    "team_size",
									"operator": "<",
									"value":    3,
								},
							},
						},
						"error_type":    "blocking",
						"error_message": "團體報名至少需要 3 人",
					},
				},
				"discount_stacking": map[string]interface{}{
					"mode":        "multiplicative",
					"description": "多重折扣採用乘法計算",
				},
			},
		},
		{
			"name":        "簡單早鳥優惠",
			"description": "僅包含早鳥優惠規則",
			"rule_set": map[string]interface{}{
				"event_id": "simple-event-2025",
				"version":  "1.0",
				"name":     "簡單活動",
				"variables": map[string]interface{}{
					"base_price": 1000,
				},
				"pricing_rules": []interface{}{
					map[string]interface{}{
						"id":          "early_bird",
						"priority":    10,
						"description": "早鳥優惠",
						"condition": map[string]interface{}{
							"type":  "datetime_before",
							"field": "register_date",
							"value": time.Now().AddDate(0, 1, 0).Format(time.RFC3339),
						},
						"action": map[string]interface{}{
							"type":   "percentage_discount",
							"target": "registration_fee",
							"value":  20,
							"label":  "早鳥優惠 8 折",
						},
					},
				},
				"validation_rules": []interface{}{},
			},
		},
	}

	// Try to load dahoo marathon example
	if data, err := os.ReadFile("./examples/dahoo-marathon-2026.json"); err == nil {
		var ruleSet map[string]interface{}
		if err := json.Unmarshal(data, &ruleSet); err == nil {
			examples = append(examples, map[string]interface{}{
				"name":        "2026大湖草莓文化嘉年華馬拉松",
				"description": "完整的真實案例，包含動態表單定義",
				"rule_set":    ruleSet,
			})
		}
	}

	return examples
}
