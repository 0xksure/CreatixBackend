CREATE TABLE USERS
(
  ID SERIAL PRIMARY KEY,
  Firstname varchar(64) NOT NULL,
  Lastname varchar(64) NOT NULL,
  Email varchar(64) NOT NULL,
  Password varchar(64) NOT NULL
);
