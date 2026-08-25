getToken();

const setId = new URLSearchParams(window.location.search).get('id');

let machineId = null;

async function loadSet() {
    try {
        const set = await getSet(setId);

        machineId = set.machine_id;
        document.querySelector('#machine-name').textContent = set.machine_name;
        document.querySelector('#weight').value = set.weight_kg;
        document.querySelector('#reps').value = set.reps;
        document.querySelector('#back-link').href = 'machine?id=' + machineId;

    } catch (error) {
        console.error(error);
    }
}

document.querySelectorAll('.stepper__btn').forEach(function(btn) {
    btn.addEventListener('click', function() {
        const target = document.querySelector('#' + btn.dataset.target);
        const step = parseFloat(btn.dataset.step);
        const min = parseFloat(target.min);
        const current = parseFloat(target.value) || 0;
        target.value = Math.max(min, current + step);
    });
});

document.querySelector('#edit-set-form').addEventListener('submit', async function(event) {
    event.preventDefault();

    const weight = parseFloat(document.querySelector('#weight').value);
    const reps = parseInt(document.querySelector('#reps').value, 10);
    const errorEl = document.querySelector('#error-message');

    if (!weight || !reps) {
        errorEl.textContent = 'Укажите вес и количество повторов';
        errorEl.style.display = 'block';
        return;
    }

    try {
        await updateSet(setId, weight, reps);
        window.location.href = 'machine?id=' + machineId;
    } catch (error) {
        errorEl.textContent = error.message;
        errorEl.style.display = 'block';
    }
});

document.querySelector('#delete-btn').addEventListener('click', async function() {
    if (!confirm('Удалить этот подход?')) {
        return;
    }

    try {
        await deleteSet(setId);
        window.location.href = 'machine?id=' + machineId;
    } catch (error) {
        alert('Не удалось удалить подход: ' + error.message);
    }
});

loadSet();