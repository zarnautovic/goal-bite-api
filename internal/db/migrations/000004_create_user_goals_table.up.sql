CREATE TABLE IF NOT EXISTS user_goals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    target_kcal NUMERIC(12,4) NOT NULL CHECK (target_kcal > 0),
    target_protein_g NUMERIC(12,4) NOT NULL CHECK (target_protein_g > 0),
    target_carbs_g NUMERIC(12,4) NOT NULL CHECK (target_carbs_g > 0),
    target_fat_g NUMERIC(12,4) NOT NULL CHECK (target_fat_g > 0),
    weight_goal_kg NUMERIC(12,4) CHECK (weight_goal_kg > 0),
    activity_level TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_goals_user_id ON user_goals(user_id);
