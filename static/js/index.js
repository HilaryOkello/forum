document.querySelectorAll(".sort-btn").forEach((btn) => {
    btn.addEventListener("click", function () {
        document.querySelector(".sort-btn.active").classList.remove("active");
        this.classList.add("active");
    });
});
window.addEventListener("load", () => {
    const layout = document.querySelector(".layout");
    const updateLayout = () => {
        const windowWidth = window.innerWidth;
        const contentWidth = Math.min(1600, windowWidth * 0.95);
        layout.style.maxWidth = `${contentWidth}px`;
    };
    updateLayout();
    window.addEventListener("resize", updateLayout);
});
const userBtn = document.querySelector(".user-dropdown .nav-btn");
const userMenu = document.querySelector(".user-menu");
userBtn.addEventListener("click", (e) => {
    e.stopPropagation();
    userMenu.classList.toggle("show");
});
document.addEventListener("click", (e) => {
    if (!userMenu.contains(e.target)) {
        userMenu.classList.remove("show");
    }
});
userMenu.addEventListener("click", (e) => {
    e.stopPropagation();
});

// Get all 'post' divs with the class 'post'
const postDivs = document.querySelectorAll('.list-post');
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
})

document.addEventListener("DOMContentLoaded", () => {
    // Add event listeners for like-btn and dislike-btn dynamically for each post
    document.querySelectorAll(".post").forEach((post) => {
        const likeButton = post.querySelector(".like-btn");
        const dislikeButton = post.querySelector(".dislike-btn");
        const postID = post.getAttribute("post-id");

        likeButton.addEventListener("click", () =>
            handlePostReaction(postID, "LIKE", likeButton, dislikeButton)
        );
        dislikeButton.addEventListener("click", () =>
            handlePostReaction(postID, "DISLIKE", dislikeButton, likeButton)
        );
    });

    const handlePostReaction = async (postID, reactionType, clickedButton, otherButton) => {
        try {
            const response = await fetch("/post/react", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    post_id: Number(postID),
                    reaction_type: reactionType,
                }),
            });

            const text = await response.text(); // Read the response as text first
            if (!response.ok || text.startsWith("<")) {
                // Redirect to login if the response is HTML or not OK
                window.location.href = "/login";
                return;
            }

            const data = JSON.parse(text); // Parse the response as JSON
            const countSpan = clickedButton.nextElementSibling; // Get the sibling span
            const otherCountSpan = otherButton.nextElementSibling; // Get the sibling span of other button

            // Handle reaction based on the backend's response
            if (data.status === "added") {
                // Increment the clicked button's count
                countSpan.textContent = parseInt(countSpan.textContent || "0", 10) + 1;
                clickedButton.classList.add("selected");
            } else if (data.status === "updated") {
                // Update counts: increment the clicked button's count and decrement the other
                countSpan.textContent = parseInt(countSpan.textContent || "0", 10) + 1;
                otherCountSpan.textContent = Math.max(
                    parseInt(otherCountSpan.textContent || "0", 10) - 1,
                    0
                );
                clickedButton.classList.add("selected");
                otherButton.classList.remove("selected");
            } else if (data.status === "removed") {
                // Decrement the clicked button's count
                countSpan.textContent = Math.max(
                    parseInt(countSpan.textContent || "0", 10) - 1,
                    0
                );
                clickedButton.classList.remove("selected");
            }
        } catch (error) {
            console.error("Error reacting to the comment:", error);
            alert("An error occurred. Please try again.");
        }
    };
});


