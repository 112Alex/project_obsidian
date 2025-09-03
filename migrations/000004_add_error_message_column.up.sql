-- Добавление колонки error_message в таблицу jobs, если её нет
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS error_message TEXT;

-- Если старая колонка error существует и новая пуста, копируем данные
UPDATE jobs SET error_message = error WHERE error_message IS NULL;