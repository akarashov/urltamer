CREATE TABLE users (
    id INT PRIMARY KEY
);
INSERT INTO users (id) VALUES (1);

ALTER TABLE tamers ADD COLUMN user_id INT NOT NULL DEFAULT 1;
-- ALTER TABLE tamers ADD CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id);
ALTER TABLE tamers ADD CONSTRAINT unique_user_original_url UNIQUE(user_id, original_url);

