package comments

const (
    // Query to check the current reaction of a user on a comment
    QueryCheckReaction = `
        SELECT reaction_type
        FROM comment_reactions
        WHERE comment_id = ? AND user_id = ?`

    // Query to delete a reaction
    QueryDeleteReaction = `
        DELETE FROM comment_reactions
        WHERE comment_id = ? AND user_id = ?`

    // Query to insert or update a reaction
    QueryUpsertReaction = `
        INSERT INTO comment_reactions (comment_id, user_id, reaction_type)
        VALUES (?, ?, ?)
        ON CONFLICT (comment_id, user_id) DO UPDATE SET reaction_type = ?`

    // Query to insert a new comment
    QueryCreateComment = `
        INSERT INTO comments (post_id, parent_id, content, user_id)
        VALUES (?, ?, ?, ?)
        RETURNING id, created_at`

    // Query to get comments for a post (example, adjust as needed)
    QueryGetComments = `
        SELECT id, post_id, parent_id, content, user_id, created_at
        FROM comments
        WHERE post_id = ?`
	queryGetPostComments = `
        SELECT 
            c.id, c.post_id, c.user_id, c.parent_id, c.content, c.created_at,
            u.username,
            (SELECT COUNT(*) FROM comment_reactions WHERE comment_id = c.id AND reaction_type = 'LIKE') as likes,
            (SELECT COUNT(*) FROM comment_reactions WHERE comment_id = c.id AND reaction_type = 'DISLIKE') as dislikes,
            (SELECT reaction_type FROM comment_reactions WHERE comment_id = c.id AND user_id = ?) as user_reaction
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = ?
        ORDER BY c.created_at DESC`
)