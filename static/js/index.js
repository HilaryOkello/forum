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