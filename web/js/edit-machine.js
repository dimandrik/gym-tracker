getToken();

const machineId = new URLSearchParams(window.location.search).get('id');

document.querySelector('#back-link').href = 'machine?id=' + machineId;

async function loadMachine() {
    try {
        const machine = await getMachine(machineId);

        document.querySelector('#name').value = machine.name;
        document.querySelector('#photo-preview').src = API_BASE_URL + machine.photo_url;

    } catch (error) {
        console.error(error);
    }
}

loadMachine();

const photoInput = document.querySelector('#photo-input');
const photoPreview = document.querySelector('#photo-preview');

photoInput.addEventListener('change', function() {
    const file = photoInput.files[0];
    if (!file) {
        return;
    }

    const reader = new FileReader();
    reader.onload = function() {
        photoPreview.src = reader.result;
    };
    reader.readAsDataURL(file);
});

document.querySelector('#edit-machine-form').addEventListener('submit', async function(event) {
    event.preventDefault();

    const name = document.querySelector('#name').value;
    const errorEl = document.querySelector('#error-message');

    if (!name) {
        errorEl.textContent = 'Введите название тренажёра';
        errorEl.style.display = 'block';
        return;
    }

    const formData = new FormData();
    formData.append('name', name);

    const photoFile = photoInput.files[0];
    if (photoFile) {
        formData.append('photo', photoFile);
    }

    try {
        await updateMachine(machineId, formData);
        window.location.href = 'machine?id=' + machineId;
    } catch (error) {
        errorEl.textContent = error.message;
        errorEl.style.display = 'block';
    }
});

document.querySelector('#delete-machine-btn').addEventListener('click', async function() {
    if (!confirm('Удалить этот тренажёр вместе со всей историей подходов? Действие необратимо.')) {
        return;
    }

    try {
        await deleteMachine(machineId);
        window.location.href = 'index';
    } catch (error) {
        alert('Не удалось удалить тренажёр: ' + error.message);
    }
});