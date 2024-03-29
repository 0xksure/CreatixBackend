CREATE TABLE FEEDBACK
(
    Id SERIAL PRIMARY KEY,
    UserID int NOT NULL,
    CompanyID int NOT NULL,
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
        Id SERIAL PRIMARY KEY,
        UserID int NOT NULL,
        FeedbackID INT NOT NULL,
        CreatedAt TIMESTAMP
        WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, 
UpdatedAt TIMESTAMP,
DeletedAt TIMESTAMP,

CONSTRAINT fk_clap_feedback FOREIGN KEY
        (FeedbackID) REFERENCES FEEDBACK
        (Id),
CONSTRAINT fk_clap_user FOREIGN KEY
        (UserId) REFERENCES USERS
        (Id)
);

        CREATE TABLE COMMENTS
        (
            Id SERIAL PRIMARY KEY,
            UserId int NOT NULL,
            FeedbackId INT NOT NULL,
            Comment VARCHAR(1000),
            CreatedAt TIMESTAMP
            WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UpdatedAt TIMESTAMP,
    DeletedAt TIMESTAMP,

    CONSTRAINT fk_comment_feedback FOREIGN KEY
            (FeedbackId) REFERENCES FEEDBACK
            (Id),
    CONSTRAINT fk_comment_user FOREIGN KEY
            (UserId) REFERENCES USERS
            (Id)
);

        