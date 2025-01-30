fetch("/categories")
  .then((response) => response.json())
  .then((categories) => {
    const categoriesSelect = document.getElementById("categories");
    categories.forEach((category) => {
      const option = document.createElement("option");
      option.value = category.ID;
      option.textContent = category.Name;
      categoriesSelect.appendChild(option);
    });
  })
  .catch((error) => console.error("Error fetching categories:", error));

class NotificationManager {
  constructor() {
    // Create container if it doesn't exist
    this.container = document.querySelector(".notification-container");
    if (!this.container) {
      this.container = document.createElement("div");
      this.container.className = "notification-container";
      document.body.appendChild(this.container);
    }
  }

  show(message, type = "success", duration = 3000) {
    const notification = document.createElement("div");
    notification.className = `notification ${type}`;

    // Create notification content
    notification.innerHTML = `
            <i class="notification-icon fas ${this.getIcon(type)}"></i>
            <div class="notification-content">${message}</div>
            <button class="notification-close" aria-label="Close notification">×</button>
        `;

    // Add to container
    this.container.appendChild(notification);

    // Setup close button
    const closeBtn = notification.querySelector(".notification-close");
    closeBtn.onclick = () => this.hide(notification);

    // Show with animation
    requestAnimationFrame(() => notification.classList.add("show"));

    // Auto-hide
    if (duration) {
      setTimeout(() => this.hide(notification), duration);
    }
  }

  hide(notification) {
    notification.classList.remove("show");
    setTimeout(() => notification.remove(), 300);
  }

  getIcon(type) {
    switch (type) {
      case "success":
        return "fa-check-circle";
      case "error":
        return "fa-exclamation-circle";
      case "warning":
        return "fa-exclamation-triangle";
      default:
        return "fa-info-circle";
    }
  }
}

// Initialize notification manager
const notificationManager = new NotificationManager();

// Handle form submission
document.querySelector("form").addEventListener("submit", function (event) {
  event.preventDefault();

  const title = document.getElementById("title").value.trim();
  const content = document.getElementById("content").value.trim();
  const categories = Array.from(
    document.getElementById("categories").selectedOptions
  ).map((option) => option.value);

  // Validate inputs
  if (title.length > 50) {
    notificationManager.show("Title exceeds 50 characters.", "error");
    return;
  }
  if (title.length < 2) {
    notificationManager.show("Title is too short.", "error");
    return;
  }

  if (content.length < 5) {
    notificationManager.show("Content is too short.", "error");
    return;
  }

  if (content.length > 500) {
    notificationManager.show("Content exceeds Limit.", "error");
    return;
  }

  if (categories.length === 0) {
    notificationManager.show("Please select at least one category.", "error");
    return;
  }

  const postData = new URLSearchParams();
  postData.append("title", title);
  postData.append("content", content);
  categories.forEach((category) => postData.append("categories[]", category));

  const image = document.getElementById("image").files[0];
  const imagedata = new FormData();
  if (image) {
    imagedata.append("image", image);
  

    fetch("/upload-image", {
      method: "POST",
      body: imagedata,
    })
    .then((response) => {
      if (!response.ok) {
        return response.json().then((data) => {
          throw new Error(data.error || `Error: ${response.statusText}`);
        });
      }
      return response.json();
    })
    .then((data) => {
      notificationManager.show(data.message, "success");
    })
    .catch((error) => {
      notificationManager.show(error.message, "error");
    });
  
}

  fetch("/create-post", {
    method: "POST",
    body: postData,
  })
    .then((response) => {
      if (!response.ok) {
        return response.json().then((data) => {
          throw new Error(data.message || `Error: ${response.statusText}`);
        });
      }
      return response;
    })
    .then((response) => {
      if (response.redirected) {
        notificationManager.show("Post created successfully!", "success");
        setTimeout(() => {
          window.location.href = response.url;
        }, 2000);
      }
    })
    .catch((error) => {
      console.error("Error:", error);
      notificationManager.show(
        error.message || "Failed to create post",
        "error"
      );
    });
});

// Populate categories dropdown
fetch("/categories")
  .then((response) => response.json())
  .then((categories) => {
    const categoriesSelect = document.getElementById("categories");
    categories.forEach((category) => {
      const option = document.createElement("option");
      option.value = category.ID;
      option.textContent = category.Name;
      categoriesSelect.appendChild(option);
    });
  })
  .catch((error) => {
    console.error("Error fetching categories:", error);
    notificationManager.show("Failed to load categories", "error");
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
