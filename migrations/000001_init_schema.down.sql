-- Удаление триггеров
DROP TRIGGER IF EXISTS update_jobs_updated_at ON jobs;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Удаление функции
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаление индексов
DROP INDEX IF EXISTS idx_jobs_status;
DROP INDEX IF EXISTS idx_jobs_user_id;
DROP INDEX IF EXISTS idx_users_telegram_id;

-- Удаление таблиц
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS users;

-- Удаление перечислений
DROP TYPE IF EXISTS job_type;
DROP TYPE IF EXISTS job_status;