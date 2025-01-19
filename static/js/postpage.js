document.addEventListener('DOMContentLoaded', function () {
    const textarea = document.querySelector('.add-comment textarea');
    const errorSpan = document.getElementById('char-limit-error');

    textarea.addEventListener('input', function () {
        if (textarea.value.length > 200) {
            errorSpan.style.display = 'block';
        } else {
            errorSpan.style.display = 'none';
        }
    });

    textarea.closest('form').addEventListener('submit', function (event) {
        const trimmedContent = textarea.value.trim();
        if (trimmedContent === "") {
            errorSpan.textContent = "Comment content cannot be empty or only spaces!";
            errorSpan.style.display = 'block';
            event.preventDefault();
        } else {
            textarea.value = trimmedContent;
        }
    });
});
