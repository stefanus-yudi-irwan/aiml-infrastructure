IF OBJECT_ID('test.customer', 'U') IS NOT NULL
BEGIN
    DROP TABLE test.customer;
END;

IF EXISTS (
    SELECT *
    FROM sys.schemas
    WHERE name = 'test'
)
BEGIN
    DROP SCHEMA test;
END;