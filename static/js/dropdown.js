document.addEventListener('DOMContentLoaded', function() {
    document.querySelectorAll('.dropbtn').forEach(button => {
        button.addEventListener('click', function(event) {
            event.stopPropagation();
            const dropdownContent = this.nextElementSibling;
            dropdownContent.classList.toggle('show');
        });
    });

    window.addEventListener('click', function(e) {
        document.querySelectorAll('.dropdown-content').forEach(content => {
            if (!content.previousElementSibling.contains(e.target)) {
                content.classList.remove('show');
            }
        });
    });

    window.addEventListener('click', function(e) {
        document.querySelectorAll('.dropbtn').forEach(content => {
            if (!content.previousElementSibling.contains(e.target)) {
                content.classList.remove('show');
            }
        });
    });
});