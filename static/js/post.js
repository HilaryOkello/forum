fetch("/categories")
    .then(response => response.json())
    .then(categories => {
        const categoriesSelect = document.getElementById("categories");
        categories.forEach(category => {
            const option = document.createElement("option");
            option.value = category.ID;
            option.textContent = category.Name;
            categoriesSelect.appendChild(option);
        });
    })
    .catch(error => console.error("Error fetching categories:", error)
);

document.querySelector("form").addEventListener("submit", function(event) {
    event.preventDefault();

    const title = document.getElementById("title").value;
    const content = document.getElementById("content").value;
    const categories = Array.from(document.getElementById("categories").selectedOptions).map(option => option.value);

    const postData = new URLSearchParams();
    postData.append("title", title);
    postData.append("content", content);
    categories.forEach(category => postData.append("categories[]", category));

    fetch("/create-post", {
        method: "POST",
        body: postData,
    })
    .then(response => {
        if (response.redirected) {
            // Show a success pop-up and redirect after a delay
            showNotification("Post created successfully!");
            setTimeout(() => {
                window.location.href = response.url; // Redirect to the homepage
            }, 2000); // Delay of 2 seconds
        } else {
            return response.json();
        }
    })
    .then(data => {
        if (data && data.message) {
            showNotification(data.message, "error");
        }
    })
    .catch(error => {
        console.error("Error:", error);
        showNotification("Failed to create post", "error");
    });
});

/**
 * Function to display a notification
 * @param {string} message - The message to display
 * @param {string} type - Type of notification ('success' or 'error')
 */
function showNotification(message, type = "success") {
    const notification = document.createElement("div");
    notification.textContent = message;
    notification.className = `notification ${type}`;

    document.body.appendChild(notification);

    // Show the notification with smooth transitions
    setTimeout(() => {
        notification.classList.add("show");
    }, 10);

    // Hide and remove the notification after 2 seconds
    setTimeout(() => {
        notification.classList.remove("show");
        setTimeout(() => {
            notification.remove();
        }, 500);
    }, 2000);
}
