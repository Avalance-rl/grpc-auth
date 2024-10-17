package models

const Schema = `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
	email TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL,
	registration_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- CREATE TABLE IF NOT EXISTS devices (
--     device_name TEXT NOT NULL,
--     email TEXT NOT NULL,
--     registration_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     expiry_time TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL '7 days'),  -- Время удаления через неделю
--     PRIMARY KEY (email, device_name),
--     CONSTRAINT fk_user_email
--         FOREIGN KEY (email)
--         REFERENCES users(email)
--         ON DELETE CASCADE
-- );
-- 
-- CREATE OR REPLACE FUNCTION check_email_limit() RETURNS TRIGGER AS $$
-- BEGIN
--     -- Проверка количества записей с одинаковым email
--     IF (SELECT COUNT(*) FROM devices WHERE email = NEW.email) >= 5 THEN
--         RAISE EXCEPTION 'Exceeded limit of 5 devices for the same email: %', NEW.email;
--     END IF;
--     RETURN NEW;
-- END;
-- $$ LANGUAGE plpgsql;
-- 
-- CREATE TRIGGER email_device_limit
-- BEFORE INSERT ON devices
-- FOR EACH ROW
-- EXECUTE FUNCTION check_email_limit();
-- 
-- DELETE FROM devices WHERE expiry_time < CURRENT_TIMESTAMP;

`

// TODO: delete сделать раз в 1 день
