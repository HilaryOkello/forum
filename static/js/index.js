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
  // Handling comment reply functionality
  const commentsList = document.getElementById("comments-list");

  const createReplyForm = (commentId) => {
      const replyForm = document.createElement("div");
      replyForm.classList.add("reply-form");
      replyForm.innerHTML = `
          <textarea placeholder="Write your reply..."></textarea>
          <button class="submit-reply">Submit Reply</button>
          <button class="cancel-reply">Cancel</button>
      `;
      
      const cancelBtn = replyForm.querySelector(".cancel-reply");
      const submitBtn = replyForm.querySelector(".submit-reply");

      cancelBtn.addEventListener("click", () => {
          replyForm.remove();
      });

      submitBtn.addEventListener("click", () => {
          // Submit reply logic (e.g., send reply to server)
          alert("Reply submitted");
          replyForm.remove();
      });

      return replyForm;
  };

  // Add reply button for each comment
  const addReplyButton = (comment) => {
      const replyButton = document.createElement("button");
      replyButton.textContent = "Reply";
      replyButton.classList.add("reply-btn");

      replyButton.addEventListener("click", () => {
          const existingReplyForm = comment.querySelector(".reply-form");
          if (existingReplyForm) {
              existingReplyForm.remove(); // If a reply form already exists, remove it
          } else {
              const replyForm = createReplyForm(comment.id); // Create a new reply form
              comment.appendChild(replyForm); // Add the reply form below the comment
          }
      });

      return replyButton;
  };

  // Example comment rendering (you should loop through actual comments from your data)
  const renderComment = (commentData) => {
      const comment = document.createElement("div");
      comment.classList.add("comment");
      comment.id = `comment-${commentData.id}`;
      comment.innerHTML = `
          <div class="comment-header">
              <span class="comment-author">${commentData.author}</span>
              <span class="comment-time">${commentData.time}</span>
          </div>
          <p class="comment-content">${commentData.content}</p>
      `;

      // Append the reply button to each comment
      const replyButton = addReplyButton(comment);
      comment.appendChild(replyButton);

      commentsList.appendChild(comment);
  };

  // Example of adding comments
  const exampleComments = [
      { id: 1, author: "User1", time: "2 hours ago", content: "This is a comment." },
      { id: 2, author: "User2", time: "1 hour ago", content: "This is another comment." }
  ];

  exampleComments.forEach(renderComment);
});
