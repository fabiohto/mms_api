-- Create MMS table
CREATE TABLE IF NOT EXISTS mms (
    id SERIAL PRIMARY KEY,
    pair VARCHAR(10) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    value DECIMAL(20, 8) NOT NULL,
    period INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(pair, timestamp, period)
);

-- Create index for common queries
CREATE INDEX IF NOT EXISTS idx_mms_pair_timestamp ON mms(pair, timestamp);
CREATE INDEX IF NOT EXISTS idx_mms_timestamp ON mms(timestamp);

-- Create function to update timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for automatic timestamp update
CREATE TRIGGER update_mms_updated_at
    BEFORE UPDATE ON mms
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
