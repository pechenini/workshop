CREATE TABLE IF NOT EXISTS todos
(
    id          serial NOT NULL PRIMARY KEY,
    title       varchar(255),
    description text(255)
);