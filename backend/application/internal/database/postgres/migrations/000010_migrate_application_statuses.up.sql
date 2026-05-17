UPDATE applications SET status = 'pending_verification' WHERE status = 'awaiting_approval';
UPDATE applications SET status = 'failed'               WHERE status = 'cancelled';
