package post

const (
	// FetchAllPostsWithMetadata retrieves all posts along with relevant metadata
	FetchAllPostsWithMetadata = `
		SELECT 
			p.id, 
			p.title, 
			p.content, 
			p.image,
			u.username, 
			p.created_at,
			COALESCE(c.comment_count, 0) AS comment_count,
			COALESCE(r.likes, 0) AS likes,
			COALESCE(r.dislikes, 0) AS dislikes,
			COALESCE(pr.reaction_type, '') AS user_reaction
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS comment_count 
			FROM comments 
			GROUP BY post_id
		) c ON p.id = c.post_id
		LEFT JOIN (
			SELECT 
				post_id, 
				SUM(CASE WHEN reaction_type = 'LIKE' THEN 1 ELSE 0 END) AS likes,
				SUM(CASE WHEN reaction_type = 'DISLIKE' THEN 1 ELSE 0 END) AS dislikes
			FROM post_reactions
			GROUP BY post_id
		) r ON p.id = r.post_id
		LEFT JOIN (
			SELECT post_id, reaction_type
			FROM post_reactions
			WHERE user_id = ?
		) pr ON p.id = pr.post_id
		ORDER BY p.created_at DESC;
	`

	// SQL query to fetch the post with additional fields, including the user's reaction.
	FetchPostWithUserReaction = `
		SELECT 
			p.id, 
			p.title, 
			p.content, 
			p.image,
			u.username, 
			p.created_at,
			COALESCE(c.comment_count, 0) AS comment_count,
			COALESCE(r.likes, 0) AS likes,
			COALESCE(r.dislikes, 0) AS dislikes,
			COALESCE(pr.reaction_type, '') AS user_reaction -- Fetch user's reaction or default to empty string
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS comment_count 
			FROM comments 
			GROUP BY post_id
		) c ON p.id = c.post_id
		LEFT JOIN (
			SELECT 
				post_id, 
				SUM(CASE WHEN reaction_type = 'LIKE' THEN 1 ELSE 0 END) AS likes,
				SUM(CASE WHEN reaction_type = 'DISLIKE' THEN 1 ELSE 0 END) AS dislikes
			FROM post_reactions
			GROUP BY post_id
		) r ON p.id = r.post_id
		LEFT JOIN (
			SELECT post_id, reaction_type
			FROM post_reactions
			WHERE user_id = ?
		) pr ON p.id = pr.post_id
		WHERE p.id = ?;
	`

	// SQL query to fetch posts for a specific category with additional fields, including the user's reaction.
	GetFilteredPostsByCategory = `
	SELECT 
		p.id, 
		p.title, 
		p.content, 
		p.image, -- Add the image column here
		u.username, 
		p.created_at,
		COALESCE(c.comment_count, 0) AS comment_count,
		COALESCE(r.likes, 0) AS likes,
		COALESCE(r.dislikes, 0) AS dislikes,
		COALESCE(pr.reaction_type, '') AS user_reaction -- Fetch user's reaction or default to empty string
	FROM posts p
	JOIN users u ON p.user_id = u.id
	JOIN post_categories pc ON p.id = pc.post_id
	JOIN categories cat ON pc.category_id = cat.id
	LEFT JOIN (
		SELECT post_id, COUNT(*) AS comment_count 
		FROM comments 
		GROUP BY post_id
	) c ON p.id = c.post_id
	LEFT JOIN (
		SELECT 
			post_id, 
			SUM(CASE WHEN reaction_type = 'LIKE' THEN 1 ELSE 0 END) AS likes,
			SUM(CASE WHEN reaction_type = 'DISLIKE' THEN 1 ELSE 0 END) AS dislikes
		FROM post_reactions
		GROUP BY post_id
	) r ON p.id = r.post_id
	LEFT JOIN (
		SELECT post_id, reaction_type
		FROM post_reactions
		WHERE user_id = ?
	) pr ON p.id = pr.post_id
	WHERE LOWER(cat.name) = ?
	ORDER BY p.id DESC;
`

	// The query filters `post_reactions` based on the user's ID and reaction type ('LIKE').
	FetchLikedPostsByUser = `
		SELECT post_id
		FROM post_reactions
		WHERE user_id = ? AND reaction_type = 'LIKE';
	`
)