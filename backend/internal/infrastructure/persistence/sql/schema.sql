-- Schema definitions for Research Helper

CREATE TABLE IF NOT EXISTS explorations (
    prompt TEXT PRIMARY KEY NOT NULL,
    normalized_term TEXT NOT NULL,
    angles_json TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS papers (
    id TEXT PRIMARY KEY NOT NULL,
    title TEXT NOT NULL,
    abstract TEXT NOT NULL DEFAULT '',
    authors_json TEXT NOT NULL DEFAULT '[]',
    year INTEGER NOT NULL DEFAULT 0,
    citation_count INTEGER NOT NULL DEFAULT 0,
    influential_citation_count INTEGER NOT NULL DEFAULT 0,
    tldr TEXT NOT NULL DEFAULT '',
    fields_of_study_json TEXT NOT NULL DEFAULT '[]',
    open_access_pdf_url TEXT NOT NULL DEFAULT '',
    doi TEXT NOT NULL DEFAULT '',
    venue TEXT NOT NULL DEFAULT '',
    fetched_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_papers_year ON papers(year);
CREATE INDEX IF NOT EXISTS idx_papers_citation_count ON papers(citation_count);
CREATE INDEX IF NOT EXISTS idx_papers_title ON papers(title);

CREATE TABLE IF NOT EXISTS citation_graphs (
    anchor_paper_id TEXT PRIMARY KEY NOT NULL,
    nodes_json TEXT NOT NULL DEFAULT '[]',
    edges_json TEXT NOT NULL DEFAULT '[]',
    traversed_at TIMESTAMP NOT NULL,
    FOREIGN KEY (anchor_paper_id) REFERENCES papers(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS grounded_syntheses (
    anchor_paper_id TEXT NOT NULL,
    concept TEXT NOT NULL,
    definitions_json TEXT NOT NULL DEFAULT '[]',
    mechanisms_json TEXT NOT NULL DEFAULT '[]',
    case_studies_json TEXT NOT NULL DEFAULT '[]',
    generated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (anchor_paper_id, concept)
);

CREATE TABLE IF NOT EXISTS chat_histories (
    id TEXT PRIMARY KEY NOT NULL,
    title TEXT NOT NULL,
    initial_prompt TEXT NOT NULL,
    model TEXT NOT NULL DEFAULT '',
    current_stage TEXT NOT NULL DEFAULT 'EXPLORATION',
    status TEXT NOT NULL DEFAULT 'DRAFT',
    mode TEXT NOT NULL DEFAULT 'AUTONOMOUS',
    approach TEXT NOT NULL DEFAULT 'NORMAL',
    checkpoint TEXT NOT NULL DEFAULT 'NONE',
    selected_angle_id TEXT NOT NULL DEFAULT '',
    anchor_paper_id TEXT NOT NULL DEFAULT '',
    stage_data_json TEXT NOT NULL DEFAULT '{}',
    pinned BOOLEAN NOT NULL DEFAULT 0,
    is_archived BOOLEAN NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_chat_histories_created_at ON chat_histories(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chat_histories_status ON chat_histories(status);
CREATE INDEX IF NOT EXISTS idx_chat_histories_pinned ON chat_histories(pinned DESC, updated_at DESC);

CREATE TABLE IF NOT EXISTS chat_messages (
    id TEXT PRIMARY KEY NOT NULL,
    chat_id TEXT NOT NULL,
    role TEXT NOT NULL,                                  -- 'user', 'assistant', 'system'
    content TEXT NOT NULL DEFAULT '',                    -- Message text content / markdown synthesis
    status TEXT NOT NULL DEFAULT 'done',                 -- 'sending', 'streaming', 'done', 'error'
    stage TEXT NOT NULL DEFAULT '',                      -- Stage milestone: 'EXPLORATION', 'RETRIEVAL', etc.
    steps_json TEXT NOT NULL DEFAULT '[]',                -- Array of execution / reasoning steps
    thinking_steps_json TEXT NOT NULL DEFAULT '[]',       -- Fine-grained consensus-style thinking steps
    stats_summary_json TEXT NOT NULL DEFAULT '{}',        -- Quantitative metrics (retrieved, eligible, included)
    citations_json TEXT NOT NULL DEFAULT '[]',            -- Referenced academic citations
    quote_json TEXT NOT NULL DEFAULT '{}',               -- Verbatim empirical quote evidence
    checkpoint_json TEXT NOT NULL DEFAULT '{}',          -- Checkpoint data (angles or candidates)
    suggested_follow_ups_json TEXT NOT NULL DEFAULT '[]', -- Dynamic contextual follow-up questions
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (chat_id) REFERENCES chat_histories(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_chat_id ON chat_messages(chat_id, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created_at ON chat_messages(created_at ASC);

-- Compatibility views
CREATE VIEW IF NOT EXISTS chat_history AS SELECT * FROM chat_histories;
CREATE VIEW IF NOT EXISTS research_pipelines AS SELECT * FROM chat_histories;
