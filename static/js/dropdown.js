document.addEventListener('DOMContentLoaded', function () {
    const dropdownButtons = document.querySelectorAll('.dropbtn');
    
    dropdownButtons.forEach(button => {
        button.addEventListener('click', function (event) {
            event.stopPropagation();
            const dropdownContent = this.nextElementSibling;

            const isVisible = dropdownContent.classList.contains('show');
            document.querySelectorAll('.dropdown-content').forEach(content => {
                content.classList.remove('show');
                content.previousElementSibling.setAttribute('aria-expanded', 'false');
            });

            if (!isVisible) {
                dropdownContent.classList.add('show');
                this.setAttribute('aria-expanded', 'true');
            }
        });
    });

    window.addEventListener('click', function () {
        document.querySelectorAll('.dropdown-content').forEach(content => {
            content.classList.remove('show');
            content.previousElementSibling.setAttribute('aria-expanded', 'false');
        });
    });

    window.addEventListener('keydown', function (e) {
        if (e.key === 'Escape') {
            document.querySelectorAll('.dropdown-content').forEach(content => {
                content.classList.remove('show');
                content.previousElementSibling.setAttribute('aria-expanded', 'false');
            });
        }
    });
});