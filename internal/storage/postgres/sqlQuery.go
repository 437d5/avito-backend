package postgres

const RollbackQuery = `
        BEGIN;
        -- Отключаем триггер
        ALTER TABLE tenders DISABLE TRIGGER tender_version_trigger;
        -- Версия из которой мы будем копировать все
        WITH target_version AS (
            SELECT version
            FROM tenders_history
            WHERE tender_id = $1 AND version = $2
        ),
        -- Тендер в который мы будем копировать все
        current_tender AS (
            SELECT *
            FROM tenders
            WHERE id = $1
        )
        UPDATE tenders
        SET organization_id = th.organization_id,
            creator_id = th.creator_id,
            name = th.name,
            description = th.description,
            status = th.status,
            service_type = th.service_type,
            created_at = th.created_at,
            version = (SELECT version FROM target_version)
        FROM (
            SELECT tender_id, organization_id, creator_id, name, description, status, service_type, created_at 
            FROM tenders_history
            WHERE tender_id = $1 AND version = $2
        ) AS th
        WHERE tenders.id = th.tender_id;
        -- Версия из которой мы будем копировать все
        WITH target_version AS (
            SELECT version
            FROM tenders_history
            WHERE tender_id = $1 AND version = $2
        )
        -- Удаляем все версии после той и ту, к которой мы делаем роллбэк чтобы не было дубликатов
        DELETE FROM tenders_history
        WHERE tender_id = $1 AND $2 >= (SELECT version FROM target_version);

        -- Включаем триггер обратно
        ALTER TABLE tenders ENABLE TRIGGER tender_version_trigger;

        COMMIT;

        SELECT id, name, description, status, service_type, version, created_at
        FROM tenders
        WHERE id=$1;
    `