CREATE TABLE media_uploads (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    object_key text NOT NULL UNIQUE,
    original_name text NOT NULL,
    mime_type text NOT NULL CHECK (mime_type IN ('image/jpeg', 'image/png', 'image/webp')),
    size bigint NOT NULL CHECK (size > 0 AND size <= 10485760),
    status text NOT NULL CHECK (status IN ('pending', 'uploaded', 'expired', 'failed')),
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_media_uploads_user_created ON media_uploads(user_id, created_at DESC);
