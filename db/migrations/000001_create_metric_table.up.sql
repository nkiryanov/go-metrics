CREATE TYPE metric_type AS ENUM ('counter', 'gauge');

CREATE TABLE "metric" (
    "type" metric_type,
    "name" varchar NOT NULL,
    "delta" bigint,
    "value" double precision,
    CONSTRAINT "metric_counter_has_delta_only" CHECK (("type" != 'counter') OR ("delta" IS NOT NULL AND "value" IS NULL)),
    CONSTRAINT "metric_gauge_has_value_only" CHECK (("type" != 'gauge') OR ("value" IS NOT NULL AND "delta" IS NULL)),
    PRIMARY KEY ("name", "type")
);
