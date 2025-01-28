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