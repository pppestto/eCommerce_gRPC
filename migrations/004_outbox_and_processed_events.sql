-- Outbox table: атомарная запись бизнес-данных и событий в одной транзакции
CREATE TABLE IF NOT EXISTS outbox (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(64) NOT NULL,
    aggregate_id   VARCHAR(64) NOT NULL,
    event_type     VARCHAR(64) NOT NULL,
    payload        JSONB NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outbox_unpublished ON outbox (created_at) WHERE published_at IS NULL;

-- Processed events: идемпотентность консьюмера (exactly-once семантика)
CREATE TABLE IF NOT EXISTS processed_events (
    idempotency_key VARCHAR(256) PRIMARY KEY,
    processed_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Order events log: аудит событий заказов (обрабатывается идемпотентным консьюмером)
CREATE TABLE IF NOT EXISTS order_events_log (
    id            SERIAL PRIMARY KEY,
    order_id      VARCHAR(64) NOT NULL,
    event_type    VARCHAR(64) NOT NULL,
    payload       JSONB NOT NULL,
    processed_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_order_events_log_order_id ON order_events_log(order_id);
