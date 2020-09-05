CREATE TABLE FEEDBACK
(
    ID SERIAL PRIMARY KEY,
    UserID int NOT NULL,
    Title VARCHAR(40),
    Description VARCHAR(40),
    CreatedAt TIMESTAMP
    WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UpdatedAt TIMESTAMP,
    DeletedAt TIMESTAMP,

    CONSTRAINT fk_feedback_user FOREIGN KEY
    (UserID) REFERENCES USERS
    (ID)
);

    CREATE TABLE CLAPS
    (
        ID SERIAL PRIMARY KEY,
        UserID int NOT NULL,
        FeedbackID INT NOT NULL,
        CreatedAt TIMESTAMP
        WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        UpdatedAt TIMESTAMP,
    DeletedAt TIMESTAMP,

    CONSTRAINT fk_clap_feedback FOREIGN KEY
        (FeedbackID) REFERENCES FEEDBACK
        (ID),
    CONSTRAINT fk_clap_user FOREIGN KEY
        (UserID) REFERENCES USERS
        (ID)
);

        CREATE TABLE COMMENTS
        (
            ID SERIAL PRIMARY KEY,
            UserID int NOT NULL,
            FeedbackID INT NOT NULL,
            Comment VARCHAR(1000),
            CreatedAt TIMESTAMP
            WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            UpdatedAt TIMESTAMP,
    DeletedAt TIMESTAMP,

    CONSTRAINT fk_comment_feedback FOREIGN KEY
            (FeedbackID) REFERENCES FEEDBACK
            (ID),
    CONSTRAINT fk_comment_user FOREIGN KEY
            (UserID) REFERENCES USERS
            (ID)
);

            CREATE TABLE COMPANY
            (
                ID SERIAL PRIMARY KEY,
                Name VARCHAR(100)
            );

            CREATE TABLE TEAM
            (
                ID SERIAL PRIMARY KEY,
                CompanyID INT NOT NULL,
                Name VARCHAR
                (100),

                CONSTRAINT fk_team_company FOREIGN KEY (CompanyID) REFERENCES COMPANY(ID)
            );

            CREATE TABLE USER_COMPANY
            (
                CompanyID INT NOT NULL,
                UserID INT NOT NULL
            );

            CREATE TABLE USER_TEAM
            (
                TeamID INT NOT NULL,
                UserID INT NOT NULL
            );