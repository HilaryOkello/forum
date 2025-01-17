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
  