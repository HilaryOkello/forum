package post
import(
	"forum/db"
)
// delete a post by the post id
func DeletePost(postID int64) error {
	_, err := db.DB.Exec("DELETE FROM posts WHERE id = ?", postID)
	if err != nil {
		return err
	}
	return nil
}

//update post image by post id
func UpdatePostImage(postID int64, image string) {
	_, err := db.DB.Exec("UPDATE posts SET image = ? WHERE id = ?", image, postID)
	if err != nil {
		return
	}
}