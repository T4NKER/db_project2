DO $$ BEGIN IF NOT EXISTS (
    SELECT
    FROM pg_database
    WHERE datname = 'my_database3'
) THEN CREATE DATABASE my_database3;
END IF;
END $$;