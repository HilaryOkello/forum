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
    const commentCount = document.getElementById("comment-count");
    const viewPostContainer = document.getElementById("view-post");

    if (!viewPostContainer) return;

    const postID = viewPostContainer.getAttribute("post-id");
    const MAX_NESTING_LEVEL = 3;

    const fetchComments = async (postID) => {
        try {
            const response = await fetch(`/comments?post_id=${postID}`);
            return await response.json();
        } catch (error) {
            console.error("Error fetching comments:", error);
            return [];
        }
    };

    const createCommentElement = (commentData, level, postID) => {
        const comment = document.createElement("div");
        comment.classList.add("comment");
        comment.id = `comment-${commentData.id}`;
        comment.dataset.level = level;
        comment.innerHTML = `
            <div class="comment-header">
                <span class="comment-author">${commentData.username}</span>
                <span class="comment-time">${new Date(commentData.created_at).toLocaleString()}</span>
            </div>
            <p class="comment-content">${commentData.content}</p>
        `;
        if (level < MAX_NESTING_LEVEL) addReplyButton(comment, commentData, postID, level);
        return comment;
    };

    const addReplyButton = (comment, commentData, postID, level) => {
        const replyButton = document.createElement("button");
        replyButton.textContent = "Reply";
        replyButton.classList.add("reply-btn");
        replyButton.addEventListener("click", () => toggleReplyForm(comment, commentData, postID));
        comment.appendChild(replyButton);
    };

    const toggleReplyForm = (comment, commentData, postID) => {
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
            replyForm.querySelector(".cancel-reply").addEventListener("click", () => replyForm.remove());
            replyForm.querySelector(".submit-reply").addEventListener("click", () => submitReply(replyForm, commentData, postID, comment));
            comment.appendChild(replyForm);
        }
    };

    const submitReply = (replyForm, commentData, postID, comment) => {
        const replyContent = replyForm.querySelector("textarea").value;

        if (!replyContent.trim()) {
            alert("Reply cannot be empty.");
            return;
        }

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
                    const parentLevel = parseInt(comment.dataset.level);
                    const replyLevel = parentLevel + 1;
                    const tempReply = createCommentElement({
                        content: replyContent,
                        username: "You",
                        created_at: new Date().toISOString(),
                    }, replyLevel, postID);
                    let repliesContainer = comment.nextElementSibling;
                    if (!repliesContainer || !repliesContainer.classList.contains("replies-container")) {
                        repliesContainer = document.createElement("div");
                        repliesContainer.classList.add("replies-container");
                        comment.parentElement.insertBefore(repliesContainer, comment.nextSibling);
                    }
                    repliesContainer.style.display = "block";
                    repliesContainer.appendChild(tempReply);

                    // Update reply cound display
                    const replyCount = repliesContainer.children.length;
                    let viewRepliesBtn = comment.querySelector(".view-replies-btn");
                    if (!viewRepliesBtn) {
                        viewRepliesBtn = document.createElement("button");
                        viewRepliesBtn.classList.add("view-replies-btn");
                        comment.appendChild(viewRepliesBtn)
                    }
                    viewRepliesBtn.textContent = "Hide Replies"
                    replyForm.remove();
                } else {
                    alert("Failed to post reply.");
                }
            });
    };

    const renderReplies = (commentData, repliesContainer, level, postID) => {
        if (commentData.children && commentData.children.length > 0 && level < MAX_NESTING_LEVEL) {
            commentData.children.forEach((childComment) => {
                renderComment(childComment, repliesContainer, level + 1, postID);
            });
        }
    };

    const addViewRepliesButton = (comment, repliesContainer, commentData) => {
        if (commentData.children && commentData.children.length > 0) {
            const viewRepliesButton = document.createElement("button");
            viewRepliesButton.textContent = `+ ${commentData.children.length} Replies`;
            viewRepliesButton.classList.add("view-replies-btn");
            viewRepliesButton.addEventListener("click", () => {
                const isHidden = repliesContainer.style.display === "none";
                repliesContainer.style.display = isHidden ? "block" : "none";
                viewRepliesButton.textContent = isHidden
                    ? "Hide Replies"
                    : `+ ${commentData.children.length} Replies`;
            });
            comment.appendChild(viewRepliesButton);
        }
    };

    const renderComment = (commentData, parentElement, level, postID) => {
        const comment = createCommentElement(commentData, level, postID);
        const repliesContainer = document.createElement("div");
        repliesContainer.classList.add("replies-container");
        repliesContainer.style.display = "none";

        renderReplies(commentData, repliesContainer, level, postID);
        addViewRepliesButton(comment, repliesContainer, commentData);

        parentElement.appendChild(comment);
        parentElement.appendChild(repliesContainer);
    };

    const displayComments = (comments, commentsList, postID) => {
        commentsList.innerHTML = ""; // Clear existing comments
        commentCount.textContent = `${comments.length} comments`;
        comments.forEach((comment) => renderComment(comment, commentsList, 1, postID));
    };

    const init = async () => {
        const commentsList = document.getElementById("comments-list");
        const comments = await fetchComments(postID);
        displayComments(comments, commentsList, postID);
    };

    init();
});
