document.addEventListener('DOMContentLoaded', () => {
     // Function to handle filtering posts by created posts
  function setupPostFilters() {
    // Get all instances of created-posts and liked-posts links (both in sidebar and filter menu)
    const createdPostLinks = document.querySelectorAll('#created-posts');
    const likedPostLinks = document.querySelectorAll('#liked-posts');

    // Add event listeners to all created-posts links
    createdPostLinks.forEach(link => {
      link.addEventListener('click', (event) => {
        event.preventDefault();
        filterPostsByUser();
        // Close filter menu if it's open
        const filterMenu = document.querySelector('.filter-menu');
        if (filterMenu.classList.contains('active')) {
          filterMenu.classList.remove('active');
          document.body.style.overflow = 'auto';
        }
      });
    });

    // Add event listeners to all liked-posts links
    likedPostLinks.forEach(link => {
      link.addEventListener('click', (event) => {
        event.preventDefault();
        filterPostsByLikes();
        // Close filter menu if it's open
        const filterMenu = document.querySelector('.filter-menu');
        if (filterMenu.classList.contains('active')) {
          filterMenu.classList.remove('active');
          document.body.style.overflow = 'auto';
        }
      });
    });
  }

  // Call the setup function
  setupPostFilters();
    saveAllPosts();
    console.log('DOM loaded');
    // Function to save all current posts to localStorage
    function saveAllPosts() {
        const posts = Array.from(document.querySelectorAll('.post')).map(post => ({
            id: post.getAttribute('post-id'),
            html: post.outerHTML
        }));
    
        // Avoid saving if posts haven't changed
        const currentPosts = JSON.parse(localStorage.getItem('allPosts') || '[]');
        if (JSON.stringify(posts) !== JSON.stringify(currentPosts)) {
            localStorage.setItem('allPosts', JSON.stringify(posts));
        }
    }
    
    // Function to filter posts by user
    function filterPostsByUser() {
        fetch('/userfilter')
            .then(response => {
                if (!response.ok) {
                    throw new Error('Failed to fetch user posts');
                }
                return response.json();
            })
            .then(userPostIds => {
                if (!userPostIds || userPostIds.length === 0) {
                    console.warn("No posts found for this user.");
                    return;
                }
    
                // Retrieve all posts from localStorage
                const allPosts = JSON.parse(localStorage.getItem('allPosts') || '[]');
    
                // Clear current feed while preserving the sort bar
                const feedElement = document.querySelector('.feed');
                const sortBar = feedElement.querySelector('.sort-bar');
                feedElement.innerHTML = '';
                feedElement.appendChild(sortBar);
    
                // Filter and add only user's posts
                userPostIds.forEach(postId => {
                    const matchingPost = allPosts.find(post => post.id == postId);
                    if (matchingPost) {
                        const postElement = document.createElement('div');
                        postElement.innerHTML = matchingPost.html;
    
                        // Add a class for visual differentiation (optional)
                        postElement.firstChild.classList.add('user-post');
                        
                        feedElement.appendChild(postElement.firstChild);
                    } else {
                        console.warn(`Post with ID ${postId} not found.`);
                    }
                });
            })
            .catch(error => {
                console.error('Error filtering posts:', error);
            });
    }
    
    // Add event listener to "Created Posts" link
    document.getElementById('created-posts').addEventListener('click', (event) => {
        event.preventDefault(); // Prevent default link behavior
         // Prevent default link behavior
        filterPostsByUser(); // Filter posts
    });

     // Function to filter posts by likes
     function filterPostsByLikes() {
        fetch('/likesfilter')
            .then(response => {
                if (!response.ok) {
                    throw new Error('Failed to fetch user posts');
                }
                return response.json();
            })
            .then(userPostIds => {
                if (!userPostIds || userPostIds.length === 0) {
                    console.warn("No posts found for this user.");
                    return;
                }
    
                // Retrieve all posts from localStorage
                const allPosts = JSON.parse(localStorage.getItem('allPosts') || '[]');
    
                // Clear current feed while preserving the sort bar
                const feedElement = document.querySelector('.feed');
                const sortBar = feedElement.querySelector('.sort-bar');
                feedElement.innerHTML = '';
                feedElement.appendChild(sortBar);
    
                // Filter and add only user's posts
                userPostIds.forEach(postId => {
                    const matchingPost = allPosts.find(post => post.id == postId);
                    if (matchingPost) {
                        const postElement = document.createElement('div');
                        postElement.innerHTML = matchingPost.html;
    
                        // Add a class for visual differentiation (optional)
                        postElement.firstChild.classList.add('user-post');
                        
                        feedElement.appendChild(postElement.firstChild);
                    } else {
                        console.warn(`Post with ID ${postId} not found.`);
                    }
                });
            })
            .catch(error => {
                console.error('Error filtering posts:', error);
            });
    }
    
    // Add event listener to "Created Posts" link
    document.getElementById('liked-posts').addEventListener('click', (event) => {
        event.preventDefault(); // Prevent default link behavior
        console.log('Filtering posts by user...');
        
        filterPostsByLikes(); // Filter posts
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
                console.log("Raw response:", text);
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
    
    document.addEventListener('DOMContentLoaded', () => {
        const filterButton = document.querySelector('.filter-button');
        const filterMenu = document.querySelector('.filter-menu');
        const closeFilter = document.querySelector('.close-filter');
  
        filterButton.addEventListener('click', () => {
          filterMenu.classList.add('active');
          document.body.style.overflow = 'hidden';
        });
  
        closeFilter.addEventListener('click', () => {
          filterMenu.classList.remove('active');
          document.body.style.overflow = 'auto';
        });
  
        // Close menu when clicking outside
        filterMenu.addEventListener('click', (e) => {
          if (e.target === filterMenu) {
            filterMenu.classList.remove('active');
            document.body.style.overflow = 'auto';
          }
        });
      });