CREATE TABLE USER
(
  ID integer PRIMARY KEY,
  Firstname varchar
  (40) NOT NULL,
  Lastname varchar
  (40) NOT NULL,
  Birthday timestamp NOT NULL,
  Email varchar
  (40) NOT NULL,
  Password varchar
  (40) NOT NULL
);
