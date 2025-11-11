CREATE TABLE tamers (
    uuid SERIAL PRIMARY KEY,
    short_url VARCHAR(8) NOT NULL,
    original_url VARCHAR(255) NOT NULL,
    CONSTRAINT unique_short_url UNIQUE(short_url),
    CONSTRAINT unique_original_url UNIQUE(original_url)
);

CREATE INDEX idx_tamers_short_url ON tamers(short_url);
CREATE INDEX idx_tamers_original_url ON tamers(original_url);