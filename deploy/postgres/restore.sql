SET idle_session_timeout = 0;

DO $$
BEGIN
    IF NOT pg_catalog.pg_try_advisory_lock(194917, 1) THEN
        RAISE EXCEPTION 'Another Alfa Course restore is running; retry after it finishes';
    END IF;
END
$$;

\ir database-empty.sql
\if :database_empty
    \! bash /restore/restore-archive.sh

    \if :SHELL_ERROR
        DO $$
        BEGIN
            RAISE EXCEPTION 'pg_restore failed; inspect the restore logs and database state before retrying';
        END
        $$;
    \endif

    \ir database-empty.sql
    \if :database_empty
        DO $$
        BEGIN
            RAISE EXCEPTION 'Archive created no user objects; database initialization is incomplete';
        END
        $$;
    \endif

    ANALYZE;
    \echo 'Restore and ANALYZE completed successfully.'
\else
    \echo 'Database already contains user objects; restore skipped.'
\endif
