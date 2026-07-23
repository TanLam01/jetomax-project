ALTER TABLE conversations ADD COLUMN direct_key text;
CREATE UNIQUE INDEX uq_conversations_direct_key
    ON conversations(direct_key);
