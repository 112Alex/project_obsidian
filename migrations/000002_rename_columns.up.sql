BEGIN;

-- Переименовываем колонку file_path в audio_file_path
ALTER TABLE jobs RENAME COLUMN file_path TO audio_file_path;

-- Переименовываем колонку error в error_message
ALTER TABLE jobs RENAME COLUMN error TO error_message;

COMMIT;