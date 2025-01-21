document.addEventListener("DOMContentLoaded", () => {
    const voteButtons = document.querySelectorAll(".vote-btn");

    voteButtons.forEach((btn) => {
        btn.addEventListener("click", (e) => {
            const isUpvote = e.target.classList.contains("fa-thumbs-up");
            const voteCountSpan = e.target.parentElement.nextElementSibling;

            let voteCount = parseInt(voteCountSpan.textContent) || 0;
            voteCount = isUpvote ? voteCount + 1 : voteCount - 1;

            voteCountSpan.textContent = voteCount;
        });
    });
});


// Get all 'post' divs with the class 'post'
const postDivs = document.querySelectorAll('.post');

// Add a click event listener to each 'post' div
postDivs.forEach((postDiv) => {
    postDiv.addEventListener('click', () => {
        const postId = postDiv.getAttribute('post-id');

        // Redirect to post/view with the post ID
        if (postId) {
            window.location.href = `/view-post?id=${postId}`;
        } else {
            console.error('Post ID not found!');
        }
    });
});

const commentForm = document.getElementById("comment-form");

commentForm.addEventListener("submit", (event) => {
    event.preventDefault();
    const postID = document.getElementById("view-post").getAttribute("post-id");
    const content = document.getElementById("comment-content").value;

    fetch(`/comments/create`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            post_id: Number(postID),
            parent_id: null, // Set to null for top-level comments
            user_id: 1, // Replace with the logged-in user's ID
            content: content,
        }),
    })
        .then((response) => response.json())
        .then((data) => {
            if (data.status === "success") {
                document.getElementById("comment-content").value = "";
                // Re-fetch comments to show the new one
                document.dispatchEvent(new Event("DOMContentLoaded"));
            } else {
                alert("Failed to post comment.");
            }
        });

});

document.addEventListener("DOMContentLoaded", () => {
    // Check if we are on the viewPost page by looking for a unique element
    const viewPostContainer = document.getElementById("view-post");
    if (viewPostContainer) {
        // Fetch post ID from the viewPost container
        const postID = viewPostContainer.getAttribute("post-id");

        // Send a request to fetch all comments for this post
        fetch(`/comments?post_id=${postID}`)
            .then((response) => response.json())
            .then((data) => {
                const commentsList = document.getElementById("comments-list");

                // Clear existing comments (if any)
                commentsList.innerHTML = "";

                // Function to recursively render comments and their replies
                const renderComment = (commentData, parentElement) => {
                    const comment = document.createElement("div");
                    comment.classList.add("comment");
                    comment.id = `comment-${commentData.id}`;
                    comment.innerHTML = `
                        <div class="comment-header">
                            <span class="comment-author">${commentData.username}</span>
                            <span class="comment-time">${new Date(commentData.created_at).toLocaleString()}</span>
                        </div>
                        <p class="comment-content">${commentData.content}</p>
                    `;

                    // Add reply button
                    const replyButton = document.createElement("button");
                    replyButton.textContent = "Reply";
                    replyButton.classList.add("reply-btn");
                    replyButton.addEventListener("click", () => {
                        const existingReplyForm = comment.querySelector(".reply-form");
                        if (existingReplyForm) {
                            existingReplyForm.remove();
                        } else {
                            const replyForm = document.createElement("div");
                            replyForm.classList.add("reply-form");
                            replyForm.innerHTML = `
                                <textarea placeholder="Write your reply..."></textarea>
                                <button class="submit-reply">Submit Reply</button>
                                <button class="cancel-reply">Cancel</button>
                            `;
                            const cancelBtn = replyForm.querySelector(".cancel-reply");
                            cancelBtn.addEventListener("click", () => replyForm.remove());
                            const submitBtn = replyForm.querySelector(".submit-reply");
                            submitBtn.addEventListener("click", () => {
                                // Logic to submit a reply
                                const replyContent = replyForm.querySelector("textarea").value;
                                fetch(`/comments/create`, {
                                    method: "POST",
                                    headers: { "Content-Type": "application/json" },
                                    body: JSON.stringify({
                                        post_id: Number(postID),
                                        parent_id: commentData.id,
                                        user_id: 1, // Replace with logged-in user's ID
                                        content: replyContent,
                                    }),
                                })
                                    .then((response) => response.json())
                                    .then((data) => {
                                        if (data.status === "success") {
                                            replyForm.remove();
                                            // Optionally re-fetch comments or add the new reply dynamically
                                        } else {
                                            alert("Failed to post reply.");
                                        }
                                    });
                            });
                            comment.appendChild(replyForm);
                        }
                    });
                    comment.appendChild(replyButton);

                    // Append comment to the parent element
                    parentElement.appendChild(comment);

                    // Render children (replies) recursively
                    if (commentData.children && commentData.children.length > 0) {
                        const repliesContainer = document.createElement("div");
                        repliesContainer.classList.add("replies");
                        commentData.children.forEach((childComment) =>
                            renderComment(childComment, repliesContainer)
                        );
                        parentElement.appendChild(repliesContainer);
                    }
                };

                // Render all root comments
                data.forEach((comment) => renderComment(comment, commentsList));
            })
            .catch((error) => console.error("Error fetching comments:", error));
    }
});




