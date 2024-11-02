CREATE TABLE IF NOT EXISTS users
(
    name        TEXT    PRIMARY KEY,
    pass_hash   BLOB    NOT NULL
)