CREATE TABLE COMPANY
(
    Id SERIAL PRIMARY KEY,
    Name VARCHAR(256) NOT NULL UNIQUE
);

CREATE TABLE COMPANY_ACCESS
(
    AccessID INT UNIQUE NOT NULL,
    AccessLevel VARCHAR(32) NOT NULL
);

INSERT INTO COMPANY_ACCESS
VALUES
    (1, 'admin'),
    (2, 'write'),
    (3, 'read');

CREATE TABLE USER_COMPANY
(
    CompanyId INT NOT NULL,
    UserId INT NOT NULL,
    AccessID INT NOT NULL,

    CONSTRAINT pk_user_company UNIQUE (CompanyId,UserId),
    CONSTRAINT fk_company FOREIGN KEY
    (CompanyId) REFERENCES COMPANY
    (ID),
    CONSTRAINT fk_user FOREIGN KEY
    (UserId) REFERENCES USERS
    (ID),
    CONSTRAINT fk_access FOREIGN KEY
    (AccessID) REFERENCES COMPANY_ACCESS
    (AccessID)
);

CREATE TABLE TEAM
(
    Id SERIAL PRIMARY KEY,
    CompanyID INT NOT NULL,
    Name VARCHAR
    (100),

    CONSTRAINT fk_team_company FOREIGN KEY (CompanyID) REFERENCES COMPANY(ID)
);


CREATE TABLE USER_TEAM
(
    TeamId INT NOT NULL,
    UserId INT NOT NULL,

    CONSTRAINT fk_team FOREIGN KEY (TeamId) REFERENCES TEAM
    (Id),
    CONSTRAINT fk_user FOREIGN KEY
    (UserId) REFERENCES USERS
    (ID)
);

ALTER TABLE FEEDBACK 
ADD CONSTRAINT fk_feedback_company FOREIGN KEY
    (CompanyID) REFERENCES Company
    (Id)