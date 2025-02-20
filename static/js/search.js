document.addEventListener('DOMContentLoaded', function () {
    const searchInput = document.querySelector('.search-bar');
    const searchContainer = document.querySelector('.search-container');
    let suggestionsDiv = document.querySelector('.search-suggestions');

    if (!suggestionsDiv) {
        suggestionsDiv = document.createElement('div');
        suggestionsDiv.className = 'search-suggestions';
        searchContainer.appendChild(suggestionsDiv);
    }

    let debounceTimeout;
    const DEBOUNCE_DELAY = 300;

    searchInput.addEventListener('input', function (e) {
        clearTimeout(debounceTimeout);
        const query = e.target.value.trim();

        if (query.length < 2) {
            suggestionsDiv.style.display = 'none';
            return;
        }

        debounceTimeout = setTimeout(() => {
            fetch(`/search?q=${encodeURIComponent(query)}`)
                .then(response => response.json())
                .then(results => {
                    suggestionsDiv.innerHTML = '';

                    if (results.length === 0) {
                        suggestionsDiv.style.display = 'none';
                        return;
                    }

                    const groupedResults = {
                        user: results.filter(r => r.type === 'user'),
                        category: results.filter(r => r.type === 'category'),
                        post: results.filter(r => r.type === 'post')
                    };

                    Object.entries(groupedResults).forEach(([type, items]) => {
                        if (items.length > 0) {
                            const sectionTitle = document.createElement('div');
                            sectionTitle.className = 'suggestion-section-title';
                            sectionTitle.textContent = type.charAt(0).toUpperCase() + type.slice(1) + 's';
                            suggestionsDiv.appendChild(sectionTitle);

                            items.forEach(result => {
                                const div = createSuggestionItem(result);
                                suggestionsDiv.appendChild(div);
                            });
                        }
                    });

                    suggestionsDiv.style.display = 'block';
                })
                .catch(error => {
                    console.error('Search error:', error);
                });
        }, DEBOUNCE_DELAY);
    });

    function createSuggestionItem(result) {
        const div = document.createElement('div');
        div.className = 'search-suggestion-item';

        let icon, content;
        switch (result.type) {
            case 'user':
                icon = result.avatar ?
                    `<img src="${result.avatar}" alt="" class="suggestion-avatar">` :
                    `<div class="suggestion-icon"><i class="fas fa-user"></i></div>`;
                content = `<strong>@${result.username}</strong><br>${result.name}`;
                break;
            case 'category':
                icon = `<div class="suggestion-icon"><i class="fas fa-tag"></i></div>`;
                content = result.name;
                break;
            case 'post':
                icon = `<div class="suggestion-icon"><i class="fas fa-file-alt"></i></div>`;
                content = `<strong>${result.title}</strong><br>${result.content}`;
                break;
        }

        div.innerHTML = `
            ${icon}
            <div class="suggestion-content">
                ${content}
                <div class="suggestion-type">${result.type}</div>
            </div>
        `;

        div.addEventListener('click', () => handleSuggestionClick(result));
        return div;
    }

    function handleSuggestionClick(result) {
        switch (result.type) {
            case 'user':
                window.location.href = `/profile?id=${result.id}`;
                break;
            case 'category':
                window.location.href = `/home?tab=tags&filter=${encodeURIComponent(result.name)}`;
                break;
            case 'post':
                window.location.href = `/post?id=${result.id}`;
                break;
        }
    }

    document.addEventListener('click', function (e) {
        if (!searchContainer.contains(e.target)) {
            suggestionsDiv.style.display = 'none';
        }
    });

    searchInput.addEventListener('keydown', function (e) {
        if (e.key === 'Enter') {
            e.preventDefault();
            const query = searchInput.value.trim();
            if (query) {
                window.location.href = `/search?q=${encodeURIComponent(query)}`;
            }
        }
    });
});