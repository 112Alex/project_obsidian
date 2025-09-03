-- Удаление устаревшей колонки file_path из таблицы jobs
ALTER TABLE jobs
    DROP COLUMN IF EXISTS file_path;