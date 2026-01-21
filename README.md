# Rules Engine

A high-performance, DSL-driven rules engine written in Go. Designed for dynamic event registration systems with complex pricing rules, eligibility validation, and team composition requirements.

## Features

- **DSL Parser**: JSON-based domain-specific language for defining business rules
- **Rule Evaluator**: High-performance rule execution engine with context management
- **Price Calculator**: Flexible pricing system with discounts, caps, and itemized breakdowns
- **Computed Fields**: Dynamic field computation (sum, count, conditionals)
- **Two-level Caching**: L1 (local LRU) + L2 (Redis) caching architecture

## Installation

```bash
go get github.com/charlesfan/rules-engine
```

## Quick Start

```go
package main

import (
    "github.com/charlesfan/rules-engine/internal/rules/parser"
    "github.com/charlesfan/rules-engine/internal/rules/evaluator"
    "github.com/charlesfan/rules-engine/internal/rules/calculator"
)

func main() {
    // 1. Parse DSL
    p := parser.NewParser()
    ruleSet, err := p.Parse(dslJSON)

    // 2. Evaluate rules
    eval := evaluator.NewEvaluator()
    ctx := evaluator.NewContext(variables)
    results := eval.EvaluateAll(ruleSet.Rules, ctx)

    // 3. Calculate price
    calc := calculator.NewCalculator()
    priceResult := calc.Calculate(ruleSet.PriceRules, ctx)
}
```

## DSL Examples

### Early Bird Discount

```json
{
  "condition": {
    "type": "datetime_before",
    "field": "$variables.register_date",
    "value": "2025-03-01T00:00:00Z"
  },
  "action": {
    "type": "percentage_discount",
    "value": 20,
    "target": "base"
  }
}
```

### Team Size Discount

```json
{
  "condition": {
    "type": "compare",
    "field": "$variables.team_size",
    "operator": ">=",
    "value": 3
  },
  "action": {
    "type": "percentage_discount",
    "value": 10,
    "target": "total"
  }
}
```

### Age Restriction with Guardian

```json
{
  "condition": {
    "type": "and",
    "conditions": [
      {
        "type": "array_any",
        "field": "$variables.members",
        "condition": {
          "type": "compare",
          "field": "age",
          "operator": "<",
          "value": 8
        }
      },
      {
        "type": "not",
        "condition": {
          "type": "array_any",
          "field": "$variables.members",
          "condition": {
            "type": "compare",
            "field": "age",
            "operator": ">=",
            "value": 18
          }
        }
      }
    ]
  },
  "error_message": "Teams with members under 8 must have at least one adult (18+)"
}
```

## Supported Expression Types

| Type | Description |
|------|-------------|
| `and`, `or`, `not` | Logical operators |
| `equals`, `compare` | Value comparison (>, <, >=, <=, ==, !=) |
| `datetime_before`, `datetime_after`, `datetime_between` | Time-based conditions |
| `array_any`, `array_all` | Array iteration with conditions |
| `field_exists`, `field_empty` | Field validation |
| `in_list` | Data source lookup |
| `rule_ref` | Reference other rules |
| `always_true` | Unconditional match |

## Supported Price Actions

| Type | Description |
|------|-------------|
| `set_price` | Set base price |
| `add_item` | Add line item (unit_price x quantity or fixed_price) |
| `percentage_discount` | Apply percentage discount |
| `fixed_discount` | Apply fixed amount discount |
| `price_cap` | Set maximum price limit |

## Performance

Benchmarks on Apple M-series:

| Operation | Latency | Throughput |
|-----------|---------|------------|
| Parse (simple) | ~2 μs | 492K ops/s |
| Parse (complex) | ~12 μs | 82K ops/s |
| Evaluate (simple) | ~0.03 μs | 33M ops/s |
| Evaluate (complex) | ~0.12 μs | 8.5M ops/s |
| Calculate (with discounts) | ~0.35 μs | 2.9M ops/s |

## Development

```bash
# Run tests
go test ./internal/rules/...

# Run benchmarks
go test -bench=. ./internal/rules/...

# Test coverage
go test -cover ./internal/rules/...
```

## Project Structure

```
internal/rules/
├── dsl/           # DSL type definitions
├── parser/        # DSL parser (JSON -> AST)
├── evaluator/     # Rule evaluation engine
├── calculator/    # Price calculation
└── computed/      # Computed field handlers
```

## License

MIT
