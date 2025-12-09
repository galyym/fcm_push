DROP TRIGGER IF EXISTS update_push_queue_updated_at ON push_queue;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_push_queue_status_scheduled;
DROP INDEX IF EXISTS idx_push_queue_client_id;
DROP INDEX IF EXISTS idx_push_queue_created_at;
DROP INDEX IF EXISTS idx_push_queue_status;

DROP TABLE IF EXISTS push_queue;
