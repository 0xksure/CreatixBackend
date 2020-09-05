SELECT
	ID
	,UserID
	,Title
	,Description
    , c.tag_array
	FROM FEEDBACK as f
    LEFT JOIN (
        SELECT 
        c.feedbackid as ID
        ,array_agg(c.comment) as tag_array
        FROM comments c  
        GROUP BY c.feedbackid
    ) c using(ID)
	WHERE UserID=1 AND DeletedAt IS NULL

SELECT * FROM COMMENTS