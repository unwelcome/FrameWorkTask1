-- Роль для auth сервиса
CREATE USER auth_user WITH PASSWORD 'auth_password';
GRANT CONNECT ON DATABASE auth_db TO auth_user;

-- Переходим в auth_db
\c auth_db;
GRANT USAGE ON SCHEMA public TO auth_user;
GRANT CREATE ON SCHEMA public TO auth_user;

-- Права на будущие таблицы в auth_db
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO auth_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT USAGE ON SEQUENCES TO auth_user;

-- Роль для application сервиса
\c postgres;
CREATE USER application_user WITH PASSWORD 'application_password';
GRANT CONNECT ON DATABASE application_db TO application_user;

-- Переходим в auth_db
\c application_db;
GRANT USAGE ON SCHEMA public TO application_user;
GRANT CREATE ON SCHEMA public TO application_user;

-- Права на будущие таблицы в application_db
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO application_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT USAGE ON SEQUENCES TO application_user;