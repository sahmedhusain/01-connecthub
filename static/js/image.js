
const maxFileSize = 20 * 1024 * 1024;
function validateImage(input) {
    const file = input.files[0];
    const allowedTypes = [
        'image/jpeg',
        'image/png',
        'image/gif',
        'image/webp',
        'image/bmp',
        'image/tiff',
        'image/svg+xml'
    ];

    if (!allowedTypes.includes(file.type)) {
        document.getElementById('errorMessage').innerHTML = 'Please upload a valid image type.';
        input.value = '';
        return false;
    }

    if (file.size > maxFileSize) {
        document.getElementById('errorMessage').innerHTML = 'File size exceeds 20 MB.';
        input.value = '';
        return false;
    }

    document.getElementById('errorMessage').innerHTML = '';
    return true;
}
