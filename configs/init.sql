-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    notion_token VARCHAR(255),
    notion_database_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Создание индекса для быстрого поиска по telegram_id
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);

-- Создание перечисления для статуса задачи
CREATE TYPE job_status AS ENUM (
    'pending',
    'processing',
    'transcribing',
    'summarizing',
    'integrating',
    'completed',
    'failed'
);

-- Создание перечисления для типа задачи
CREATE TYPE job_type AS ENUM (
    'transcription',
    'summarization',
    'notion_integration',
    'notification'
);

-- Создание таблицы задач
CREATE TABLE IF NOT EXISTS jobs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    audio_file_path VARCHAR(255) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    duration INTEGER,
    status job_status NOT NULL DEFAULT 'pending',
    transcription TEXT,
    summary TEXT,
    notion_page_id VARCHAR(255),
    notion_database_id VARCHAR(255),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Создание индекса для быстрого поиска по user_id
CREATE INDEX IF NOT EXISTS idx_jobs_user_id ON jobs(user_id);

-- Создание индекса для быстрого поиска по статусу
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);

-- Функция для обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер для обновления updated_at в таблице users
CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Триггер для обновления updated_at в таблице jobs
CREATE TRIGGER update_jobs_updated_at
BEFORE UPDATE ON jobs
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();