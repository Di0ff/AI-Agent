CREATE TABLE IF NOT EXISTS tasks (
    id              SERIAL PRIMARY KEY,
    user_input      TEXT NOT NULL,
    status          VARCHAR(32) NOT NULL DEFAULT 'pending',
    result_summary  TEXT,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS agent_steps (
    id              SERIAL PRIMARY KEY,
    task_id         INT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    step_no         INT NOT NULL,
    action_type     VARCHAR(64) NOT NULL,
    target_selector TEXT,
    reasoning       TEXT,
    result          TEXT,
    screenshot_path TEXT,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS llm_logs (
    id              SERIAL PRIMARY KEY,
    task_id         INT REFERENCES tasks(id) ON DELETE CASCADE,
    step_id         INT REFERENCES agent_steps(id) ON DELETE SET NULL,
    role            VARCHAR(16) NOT NULL,
    prompt_text     TEXT NOT NULL,
    response_text   TEXT,
    model           VARCHAR(64),
    tokens_used     INT DEFAULT 0,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);


