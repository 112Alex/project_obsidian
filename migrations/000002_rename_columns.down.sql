BEGIN;

-- Переименовываем колонку audio_file_path обратно в file_path
ALTER TABLE jobs RENAME COLUMN audio_file_path TO file_path;

-- Переименовываем колонку error_message обратно в error
ALTER TABLE jobs RENAME COLUMN error_message TO error;

COMMIT;