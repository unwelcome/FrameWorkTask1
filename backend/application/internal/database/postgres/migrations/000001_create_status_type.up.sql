CREATE TYPE application_status AS ENUM (
    'created',
    'assigned',
    'in_progress',
    'on_hold',
    'awaiting_approval',
    'completed',
    'cancelled',
    'failed',
    'archived'
);
