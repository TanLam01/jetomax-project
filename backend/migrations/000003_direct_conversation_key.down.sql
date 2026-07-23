DROP INDEX IF EXISTS uq_conversations_direct_key;
ALTER TABLE conversations DROP COLUMN IF EXISTS direct_key;
