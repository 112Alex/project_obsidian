-- Добавление колонки audio_file_path в таблицу jobs, если её нет
ALTER TABLE jobs
    ADD COLUMN IF NOT EXISTS audio_file_path VARCHAR(255);

-- Заполняем новую колонку существующим значением из file_path, если оно присутствует
UPDATE jobs
SET audio_file_path = file_path
WHERE audio_file_path IS NULL AND file_path IS NOT NULL;