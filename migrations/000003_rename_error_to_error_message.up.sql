-- Переименование колонки error в error_message в таблице jobs
ALTER TABLE jobs RENAME COLUMN IF EXISTS error TO error_message;