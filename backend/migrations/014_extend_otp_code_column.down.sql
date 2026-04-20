-- Revert to original size (will truncate existing hashes!)
ALTER TABLE otp_codes ALTER COLUMN code TYPE VARCHAR(6);
