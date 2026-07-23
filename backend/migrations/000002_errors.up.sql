CREATE TABLE errors (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id text NOT NULL,
    method text NOT NULL,
    path text NOT NULL,
    status integer NOT NULL,
    error_code text NOT NULL,
    message text NOT NULL,
    client_ip text,
    user_agent text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_errors_request_id ON errors(request_id);
CREATE INDEX idx_errors_status_created ON errors(status, created_at DESC);
