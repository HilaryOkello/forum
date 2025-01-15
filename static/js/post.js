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
    .then(response => response.json())
    .then(data => {
        alert(data.message); // Show the success message
        // Optionally, you can redirect or reset the form here
    })
    .catch(error => {
        console.error("Error:", error);
        alert("Failed to create post");
    });
});
    