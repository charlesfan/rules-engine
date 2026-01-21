# Database Schema (Phase 1 - RAG)

## Overview

This document describes the database schema for the RAG (Retrieval-Augmented Generation) feature in Phase 1.

## Prerequisites

```bash
# Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;
```

## Tables

### 1. rule_examples

Stores verified rule examples for RAG retrieval.

```sql
CREATE TABLE rule_examples (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Original information
    description TEXT NOT NULL,           -- Human-readable rule description
    dsl JSONB NOT NULL,                  -- Corresponding DSL JSON

    -- Vector embedding
    description_embedding vector(768),   -- Ollama nomic-embed-text: 768 dimensions

    -- Metadata
    rule_type VARCHAR(50),               -- Rule category
    tags TEXT[],                         -- Tags for filtering
    event_type VARCHAR(50),              -- Event type (marathon, triathlon, etc.)

    -- Verification status
    verified BOOLEAN DEFAULT FALSE,      -- Human verified
    verified_by VARCHAR(100),            -- Who verified
    verified_at TIMESTAMP,               -- When verified

    -- Usage statistics
    usage_count INTEGER DEFAULT 0,       -- Times retrieved
    last_used_at TIMESTAMP,              -- Last retrieval time

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Index for vector similarity search (IVFFlat)
CREATE INDEX idx_rule_examples_embedding
ON rule_examples
USING ivfflat (description_embedding vector_cosine_ops)
WITH (lists = 100);

-- Index for filtering
CREATE INDEX idx_rule_examples_type ON rule_examples(rule_type);
CREATE INDEX idx_rule_examples_verified ON rule_examples(verified);
CREATE INDEX idx_rule_examples_event_type ON rule_examples(event_type);

-- GIN index for tags
CREATE INDEX idx_rule_examples_tags ON rule_examples USING GIN(tags);
```

### 2. conversation_logs

Stores conversation history for continuous learning.

```sql
CREATE TABLE conversation_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL,            -- Conversation session ID

    -- Conversation content
    user_input TEXT NOT NULL,            -- User's input
    ai_response TEXT NOT NULL,           -- AI's response
    generated_dsl JSONB,                 -- Generated DSL (if any)

    -- Intent detection
    detected_intent VARCHAR(20),         -- chat/rule/ambiguous
    intent_confidence DECIMAL(3,2),      -- 0.00 - 1.00

    -- User feedback
    user_confirmed BOOLEAN,              -- User confirmed DSL is correct
    user_feedback TEXT,                  -- User's modification/feedback
    final_dsl JSONB,                     -- Final DSL after modifications

    -- Processing metadata
    llm_provider VARCHAR(20),            -- ollama/claude
    llm_model VARCHAR(50),               -- Model used
    processing_time_ms INTEGER,          -- Response time

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_conversation_logs_session ON conversation_logs(session_id);
CREATE INDEX idx_conversation_logs_confirmed ON conversation_logs(user_confirmed);
CREATE INDEX idx_conversation_logs_created ON conversation_logs(created_at);
```

### 3. sessions

Tracks conversation sessions.

```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Session info
    started_at TIMESTAMP DEFAULT NOW(),
    ended_at TIMESTAMP,

    -- Statistics
    message_count INTEGER DEFAULT 0,
    rules_generated INTEGER DEFAULT 0,
    rules_confirmed INTEGER DEFAULT 0,

    -- Configuration snapshot
    llm_provider VARCHAR(20),
    llm_model VARCHAR(50),

    -- Metadata
    client_info JSONB                    -- Optional client metadata
);

CREATE INDEX idx_sessions_started ON sessions(started_at);
```

### 4. prompt_templates (Optional)

Stores prompt templates for easy modification without code changes.

```sql
CREATE TABLE prompt_templates (
    id VARCHAR(50) PRIMARY KEY,          -- Template identifier

    -- Template content
    template TEXT NOT NULL,              -- Prompt template with placeholders
    description TEXT,                    -- Description of this template

    -- Version control
    version INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## Vector Search Queries

### Find Similar Rules

```sql
-- Find top 3 similar rule examples
SELECT
    id,
    description,
    dsl,
    rule_type,
    1 - (description_embedding <=> $1) AS similarity
FROM rule_examples
WHERE verified = TRUE
ORDER BY description_embedding <=> $1  -- Cosine distance
LIMIT 3;
```

### Find Similar with Filters

```sql
-- Find similar rules with type filter
SELECT
    id,
    description,
    dsl,
    rule_type,
    1 - (description_embedding <=> $1) AS similarity
FROM rule_examples
WHERE
    verified = TRUE
    AND (rule_type = $2 OR $2 IS NULL)
    AND (event_type = $3 OR $3 IS NULL)
ORDER BY description_embedding <=> $1
LIMIT 5;
```

## Go Repository Implementation

```go
// internal/chatbot/repository/rule_examples.go

package repository

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/pgvector/pgvector-go"
)

type RuleExample struct {
    ID          string          `db:"id"`
    Description string          `db:"description"`
    DSL         json.RawMessage `db:"dsl"`
    RuleType    string          `db:"rule_type"`
    Tags        []string        `db:"tags"`
    Verified    bool            `db:"verified"`
    Similarity  float64         `db:"similarity"`
}

type RuleExampleRepository struct {
    pool *pgxpool.Pool
}

func NewRuleExampleRepository(pool *pgxpool.Pool) *RuleExampleRepository {
    return &RuleExampleRepository{pool: pool}
}

// FindSimilar finds similar rule examples using vector search
func (r *RuleExampleRepository) FindSimilar(
    ctx context.Context,
    embedding []float32,
    limit int,
) ([]RuleExample, error) {
    query := `
        SELECT
            id,
            description,
            dsl,
            rule_type,
            tags,
            verified,
            1 - (description_embedding <=> $1) AS similarity
        FROM rule_examples
        WHERE verified = TRUE
        ORDER BY description_embedding <=> $1
        LIMIT $2
    `

    vec := pgvector.NewVector(embedding)
    rows, err := r.pool.Query(ctx, query, vec, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var examples []RuleExample
    for rows.Next() {
        var ex RuleExample
        err := rows.Scan(
            &ex.ID,
            &ex.Description,
            &ex.DSL,
            &ex.RuleType,
            &ex.Tags,
            &ex.Verified,
            &ex.Similarity,
        )
        if err != nil {
            return nil, err
        }
        examples = append(examples, ex)
    }

    return examples, nil
}

// Insert adds a new rule example (unverified)
func (r *RuleExampleRepository) Insert(
    ctx context.Context,
    description string,
    dsl json.RawMessage,
    embedding []float32,
    ruleType string,
) (string, error) {
    query := `
        INSERT INTO rule_examples (description, dsl, description_embedding, rule_type)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `

    vec := pgvector.NewVector(embedding)
    var id string
    err := r.pool.QueryRow(ctx, query, description, dsl, vec, ruleType).Scan(&id)
    return id, err
}

// MarkVerified marks a rule example as verified
func (r *RuleExampleRepository) MarkVerified(
    ctx context.Context,
    id string,
    verifiedBy string,
) error {
    query := `
        UPDATE rule_examples
        SET verified = TRUE, verified_by = $2, verified_at = NOW()
        WHERE id = $1
    `
    _, err := r.pool.Exec(ctx, query, id, verifiedBy)
    return err
}

// IncrementUsage updates usage statistics
func (r *RuleExampleRepository) IncrementUsage(ctx context.Context, id string) error {
    query := `
        UPDATE rule_examples
        SET usage_count = usage_count + 1, last_used_at = NOW()
        WHERE id = $1
    `
    _, err := r.pool.Exec(ctx, query, id)
    return err
}
```

## Continuous Learning Flow

```
┌─────────────────────────────────────────────────────────────┐
│                 Continuous Learning Flow                    │
└─────────────────────────────────────────────────────────────┘

1. User confirms DSL is correct
         │
         ▼
┌─────────────────────────────────────┐
│  Save to conversation_logs          │
│  - user_confirmed = TRUE            │
│  - generated_dsl                    │
└──────────────────┬──────────────────┘
                   │
                   ▼
┌─────────────────────────────────────┐
│  Generate embedding                 │
│  - Embed user_input                 │
│  - Use Ollama nomic-embed-text      │
└──────────────────┬──────────────────┘
                   │
                   ▼
┌─────────────────────────────────────┐
│  Insert to rule_examples            │
│  - verified = FALSE (pending)       │
│  - Store embedding                  │
└──────────────────┬──────────────────┘
                   │
                   ▼
┌─────────────────────────────────────┐
│  Admin Review (manual)              │
│  - Verify DSL correctness           │
│  - Add tags/categorization          │
│  - Mark verified = TRUE             │
└─────────────────────────────────────┘
                   │
                   ▼
          Now part of RAG!
```

## Seed Data

Initial rule examples to bootstrap the system:

```sql
-- Insert seed data (run after creating tables)

INSERT INTO rule_examples (description, dsl, rule_type, verified, tags) VALUES
(
    '早鳥報名 3/1 前打 8 折',
    '{
        "id": "early_bird",
        "name": "早鳥優惠",
        "priority": 1,
        "condition": {
            "type": "datetime_before",
            "field": "$variables.register_date",
            "value": "2025-03-01T00:00:00Z"
        },
        "actions": [{
            "type": "percentage_discount",
            "value": 20,
            "target": "total"
        }]
    }',
    'discount',
    TRUE,
    ARRAY['early_bird', 'time_based', 'percentage']
),
(
    '3 人以上團體報名打 9 折',
    '{
        "id": "group_discount",
        "name": "團體優惠",
        "priority": 2,
        "condition": {
            "type": "compare",
            "field": "$variables.team_size",
            "operator": ">=",
            "value": 3
        },
        "actions": [{
            "type": "percentage_discount",
            "value": 10,
            "target": "total"
        }]
    }',
    'discount',
    TRUE,
    ARRAY['group', 'team_size', 'percentage']
),
(
    '未滿 8 歲需有成年人陪同',
    '{
        "id": "minor_guardian",
        "name": "未成年監護規則",
        "priority": 10,
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
                }
            ]
        },
        "validation": {
            "require": {
                "type": "array_any",
                "field": "$variables.members",
                "condition": {
                    "type": "compare",
                    "field": "age",
                    "operator": ">=",
                    "value": 18
                }
            },
            "error_message": "有未滿 8 歲的參賽者，必須有至少一位 18 歲以上的成年人陪同"
        }
    }',
    'validation',
    TRUE,
    ARRAY['age_restriction', 'guardian', 'minor']
);

-- Update embeddings (run separately with embedding API)
-- UPDATE rule_examples SET description_embedding = ... WHERE id = ...;
```

## Migration Commands

```bash
# Run migrations
make migrate-up-poc

# Rollback
make migrate-down-poc

# Seed data
make seed-poc
```
