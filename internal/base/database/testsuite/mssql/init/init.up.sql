IF NOT EXISTS (
    SELECT *
    FROM sys.schemas
    WHERE name = 'test'
)
BEGIN
    EXEC('CREATE SCHEMA test');
END;

IF OBJECT_ID('test.customer', 'U') IS NULL
BEGIN
    CREATE TABLE test.customer
    (
        id UNIQUEIDENTIFIER PRIMARY KEY
        , first_name NVARCHAR(255)
        , last_name NVARCHAR(255)
        , created_at BIGINT NOT NULL DEFAULT DATEDIFF_BIG(SECOND, '1970-01-01', SYSUTCDATETIME())
        , updated_at BIGINT NOT NULL DEFAULT DATEDIFF_BIG(SECOND, '1970-01-01', SYSUTCDATETIME())
        , deleted_at BIGINT NULL
    );
END;