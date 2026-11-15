SELECT NOT EXISTS (
    SELECT 1
    FROM pg_catalog.pg_namespace AS namespace
    LEFT JOIN pg_catalog.pg_depend AS dependency
      ON dependency.refclassid = 'pg_catalog.pg_namespace'::regclass
     AND dependency.refobjid = namespace.oid
    WHERE namespace.nspname !~ '^pg_'
      AND namespace.nspname <> 'information_schema'
      AND (namespace.nspname <> 'public' OR dependency.objid IS NOT NULL)
) AND NOT EXISTS (
    SELECT 1 FROM pg_catalog.pg_extension WHERE extname <> 'plpgsql'
) AND NOT EXISTS (
    SELECT 1 FROM pg_catalog.pg_largeobject_metadata
) AS database_empty
\gset
