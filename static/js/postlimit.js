
document.addEventListener('DOMContentLoaded', function () {
    const textarea = document.getElementById('content');
    const charCounter = document.getElementById('char-counter');
    const errorMessage = document.getElementById('error-message');
    const submitButton = document.getElementById('submit-button');
    const charLimit = 500;

    textarea.addEventListener('input', function () {
        const textLength = textarea.value.length;
        charCounter.textContent = `${textLength}/${charLimit}`;

        if (textLength > charLimit) {
            errorMessage.style.display = 'block';
            submitButton.disabled = true;
        } else {
            errorMessage.style.display = 'none';
            submitButton.disabled = false;
        }
    });

    submitButton.addEventListener('click', function (event) {
        const trimmedContent = textarea.value.trim();
        if (trimmedContent === "") {
            errorMessage.textContent = "Post content cannot be empty or only spaces!";
            errorMessage.style.display = 'block';
            event.preventDefault();
        } else {
            textarea.value = trimmedContent;
        }
    });
});