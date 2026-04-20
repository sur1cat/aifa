-- Extend OTP code column to store bcrypt hashes (60 characters)
ALTER TABLE otp_codes ALTER COLUMN code TYPE VARCHAR(72);
