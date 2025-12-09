CREATE TABLE IF NOT EXISTS push_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    data JSONB,
    priority VARCHAR(20) DEFAULT 'normal',
    client_id VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    error_message TEXT,
    fcm_message_id VARCHAR(255),
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_push_queue_status_scheduled ON push_queue(status, scheduled_at) 
    WHERE status IN ('pending', 'processing');

CREATE INDEX idx_push_queue_client_id ON push_queue(client_id);

CREATE INDEX idx_push_queue_created_at ON push_queue(created_at DESC);

CREATE INDEX idx_push_queue_status ON push_queue(status);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_push_queue_updated_at 
    BEFORE UPDATE ON push_queue
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE push_queue IS 'Queue table for push notification tasks';
COMMENT ON COLUMN push_queue.status IS 'Task status: pending, processing, success, failed';
COMMENT ON COLUMN push_queue.attempts IS 'Number of send attempts made';
COMMENT ON COLUMN push_queue.scheduled_at IS 'When the task should be processed next';
