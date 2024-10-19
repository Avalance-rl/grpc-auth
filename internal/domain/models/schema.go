package models

const Schema = `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
	email TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL,
	registration_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS devices (
    device_name TEXT NOT NULL,
    email TEXT NOT NULL,
    registration_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiry_time TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL '7 days'),  -- removal time in a week
    PRIMARY KEY (email, device_name),
    CONSTRAINT fk_user_email
        FOREIGN KEY (email)
        REFERENCES users(email)
        ON DELETE CASCADE
);
-- The email_device_limit trigger is intended to limit the number of devices that can be registered for
-- one user with the same email in the devices table. The limit is a maximum of 5 devices per user.
DROP TRIGGER IF EXISTS email_device_limit ON devices;

-- creating an email limit check function
CREATE OR REPLACE FUNCTION check_email_limit() RETURNS TRIGGER AS $$
BEGIN
    --checking the number of records with the same email
    IF (SELECT COUNT(*) FROM devices WHERE email = NEW.email) >= 5 THEN
        RAISE EXCEPTION 'Exceeded limit of 5 devices for the same email: %', NEW.email;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

--create a trigger
CREATE TRIGGER email_device_limit
BEFORE INSERT ON devices
FOR EACH ROW
EXECUTE FUNCTION check_email_limit();

DELETE FROM devices WHERE expiry_time < CURRENT_TIMESTAMP;
-- deletes once a day
SELECT cron.schedule('0 0 * * *', 'DELETE FROM devices WHERE expiry_time < CURRENT_TIMESTAMP');
`
