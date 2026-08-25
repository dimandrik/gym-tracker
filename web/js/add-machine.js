const photoInput = document.querySelector('#photo-input');
const photoPreview = document.querySelector('#photo-preview');
const photoPlaceholder = document.querySelector('#photo-placeholder');

const form = document.querySelector('.add-machine-form');

getToken();

form.addEventListener('submit', async function(event) {
    event.preventDefault();

    const name = document.querySelector('#name').value;
    const photoFile = photoInput.files[0];

    if (!name || !photoFile) {
        return;
    }

    try {
        const machine = await addMachine(name, photoFile);
        window.location.href = 'index';
    } catch (error) {
        const errorEl = document.querySelector('#error-message');
        errorEl.textContent = error.message;
        errorEl.style.display = 'block';
    }
});

// Показываем превью выбранного фото до отправки формы, читая файл
// в base64 через FileReader — на сервер он уйдёт отдельно, при submit.
photoInput.addEventListener('change', function() {
    const file = photoInput.files[0];
    if (!file) {
        return;
    }

    const reader = new FileReader();
    reader.onload = function() {
        photoPreview.src = reader.result;
        photoPreview.style.display = 'block';
        photoPlaceholder.style.display = 'none';
    };
    reader.readAsDataURL(file);
});

async function addMachine(name, photoFile) {
    const formData = new FormData();
    formData.append('name', name);
    formData.append('photo', photoFile);

    return apiFetch('/api/machines', {
        method: 'POST',
        body: formData
    });
}