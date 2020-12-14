CREATE TABLE USERS
(
  ID SERIAL PRIMARY KEY,
  Firstname varchar(64) NOT NULL,
  Lastname varchar(64) NOT NULL,
  Username varchar(128) NOT NULL,
  Email varchar(64) NOT NULL UNIQUE,
  Password varchar(64) NOT NULL
);
