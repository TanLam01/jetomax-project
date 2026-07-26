CREATE TABLE message_attachments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id uuid NOT NULL UNIQUE REFERENCES messages(id) ON DELETE CASCADE,
    upload_id uuid NOT NULL UNIQUE REFERENCES media_uploads(id),
    object_key text NOT NULL,
    mime_type text NOT NULL,
    size bigint NOT NULL CHECK (size > 0)
);

CREATE INDEX idx_message_attachments_message ON message_attachments(message_id);
