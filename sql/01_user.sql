CREATE TABLE  account (
  id integer PRIMARY KEY,
  name varchar(40) NOT NULL,
  email varchar(40) UNIQUE NOT NULL
);
