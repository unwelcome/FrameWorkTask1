UPDATE applications SET status = 'awaiting_approval' WHERE status = 'pending_verification';
UPDATE applications SET status = 'cancelled'         WHERE status = 'failed';
