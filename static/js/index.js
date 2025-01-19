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

document.querySelectorAll(".vote-btn").forEach((btn) => {
  btn.addEventListener("click", function () {
    const voteCount = this.parentElement.querySelector("span");
    const currentVotes = parseInt(
      voteCount.textContent.replace("k", "000")
    );
    this.parentElement
      .querySelectorAll(".vote-btn")
      .forEach((b) => (b.style.color = "#818384"));
    if (this.textContent === "▲") {
      voteCount.textContent = currentVotes + 1 + "k";
      this.style.color = "#FF4500";
    } else {
      voteCount.textContent = currentVotes - 1 + "k";
      this.style.color = "#7193FF";
    }
  });
});
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

// Get the 'post' div by its ID
const postDiv = document.getElementById('post');

postDiv.addEventListener('click', () => {
  const postId = postDiv.getAttribute('data-id');

  // Redirect to post/view with the post ID
  if (postId) {
    window.location.href = `/post/view?id=${postId}`;
  } else {
    console.error('Post ID not found!');
  }
});
