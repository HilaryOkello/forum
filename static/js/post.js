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
    .catch(error => console.error("Error fetching categories:", error));
