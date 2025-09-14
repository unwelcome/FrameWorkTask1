\connect app_db;

DROP TYPE IF EXISTS "roles";
DROP TYPE IF EXISTS "priorities";
DROP TYPE IF EXISTS "statuses";

CREATE TYPE "roles" AS ENUM (
  'инженер',
  'менеджер',
  'руководитель'
);

CREATE TYPE "priorities" AS ENUM (
  'обычно',
  'важно',
  'срочно',
  'критично',
  'первостепенно'
);

CREATE TYPE "statuses" AS ENUM (
  'создано',
  'в ожидании',
  'в работе',
  'выполнено',
  'приостановлено',
  'не выполнено',
  'на пересмотрении',
  'отозвано'
);

CREATE SEQUENCE serial_number_seq START 1 INCREMENT 1;

CREATE TABLE "users" (
    "id" SERIAL PRIMARY KEY,
    "login" varchar(255) NOT NULL,
    "password_hash" varchar(255) NOT NULL,
    "password_salt" varchar(255) NOT NULL,
    "name" varchar(40) NOT NULL,
    "second_name" varchar(40) NOT NULL,
    "third_name" varchar(40),
    "email" varchar(255) NOT NULL UNIQUE,
    "role" roles NOT NULL,
    "created_at" timestamp DEFAULT NOW()
);

CREATE TABLE "applications" (
    "id" SERIAL PRIMARY KEY,
    "serial_number" integer NOT NULL DEFAULT nextval('serial_number_seq'),
    "title" text NOT NULL,
    "priority" priorities NOT NULL,
    "description" text NOT NULL,
    "created_at" timestamp DEFAULT NOW(),
    "created_by" integer NOT NULL,
    "execution_time" timestamp,
    "closed_at" timestamp DEFAULT NULL,
    "status" statuses NOT NULL,
    "responsible_engineer" integer DEFAULT NULL,
    "responsible_manager" integer DEFAULT NULL
);

CREATE TABLE "application_changes" (
    "id" SERIAL PRIMARY KEY,
    "application_serial_number" integer NOT NULL,
    "application_id" integer NOT NULL,
    "is_current" bool DEFAULT TRUE,
    "author" integer NOT NULL,
    "changed_at" timestamp DEFAULT NOW()
);

CREATE TABLE "departments" (
    "id" SERIAL PRIMARY KEY,
    "title" varchar(255) NOT NULL,
    "created_at" timestamp DEFAULT NOW(),
    "created_by" integer NOT NULL
);

CREATE TABLE "department_employees" (
    "id" SERIAL PRIMARY KEY,
    "user_id" integer NOT NULL,
    "department_id" integer NOT NULL
);


ALTER TABLE "applications" DROP CONSTRAINT IF EXISTS "applicaitons_created_by_to_users";
ALTER TABLE "applications" DROP CONSTRAINT IF EXISTS "applications_responsible_engineer_to_users";
ALTER TABLE "applications" DROP CONSTRAINT IF EXISTS "applications_responsible_manager_to_users";
ALTER TABLE "application_changes" DROP CONSTRAINT IF EXISTS "application_changes_to_users";
ALTER TABLE "application_changes" DROP CONSTRAINT IF EXISTS "application_changes_to_application_id";
ALTER TABLE "application_changes" DROP CONSTRAINT IF EXISTS "application_changes_to_application_serial_number";
ALTER TABLE "departments" DROP CONSTRAINT IF EXISTS "departments_to_users";
ALTER TABLE "department_employees" DROP CONSTRAINT IF EXISTS "department_employees_to_users";
ALTER TABLE "department_employees" DROP CONSTRAINT IF EXISTS "department_employees_to_departments";

ALTER TABLE "applications" ADD CONSTRAINT "applicaitons_created_by_to_users" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE NULL;

ALTER TABLE "applications" ADD CONSTRAINT "applications_responsible_engineer_to_users" FOREIGN KEY ("responsible_engineer") REFERENCES "users" ("id") ON DELETE NULL;

ALTER TABLE "applications" ADD CONSTRAINT "applications_responsible_manager_to_users" FOREIGN KEY ("responsible_manager") REFERENCES "users" ("id") ON DELETE NULL;

ALTER TABLE "application_changes" ADD CONSTRAINT "application_changes_to_users" FOREIGN KEY ("author") REFERENCES "users" ("id") ON DELETE NULL;

ALTER TABLE "application_changes" ADD CONSTRAINT "application_changes_to_application_id" FOREIGN KEY ("application_id") REFERENCES "applications" ("id") ON DELETE CASCADE;

ALTER TABLE "application_changes" ADD CONSTRAINT "application_changes_to_application_serial_number" FOREIGN KEY ("application_serial_number") REFERENCES "applications" ("serial_number") ON DELETE CASCADE;

ALTER TABLE "departments" ADD CONSTRAINT "departments_to_users" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE NULL;

ALTER TABLE "department_employees" ADD CONSTRAINT "department_employees_to_users" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;

ALTER TABLE "department_employees" ADD CONSTRAINT "department_employees_to_departments" FOREIGN KEY ("department_id") REFERENCES "departments" ("id") ON DELETE CASCADE;
