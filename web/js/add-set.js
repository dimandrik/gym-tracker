const form = document.querySelector('.add-set-form');
const machineId = new URLSearchParams(window.location.search).get('machine_id');
const weightInput = document.querySelector('#weight');
const repsInput = document.querySelector('#reps');
const errorEl = document.querySelector('#error-message');
const backLink = document.querySelector('#back-link');
const toastEl = document.querySelector('#toast');
const toastMessageEl = document.querySelector('#toast-message');
const dateInput = document.querySelector('#date');
const dateDisplay = document.querySelector('#date-display');
const todayISO = new Date().toISOString().split('T')[0];

let toastTimeout;
let isSaving = false;

getToken();

backLink.href = 'machine?id=' + machineId;

// Кнопки +/- у полей веса и повторов — шаг и минимум берутся из
// data-атрибутов кнопки/инпута в разметке, а не хардкодятся здесь.
document.querySelectorAll('.stepper__btn').forEach(function(btn) {
    btn.addEventListener('click', function() {
        const target = document.querySelector('#' + btn.dataset.target);
        const step = parseFloat(btn.dataset.step);
        const min = parseFloat(target.min);
        const current = parseFloat(target.value) || 0;
        target.value = Math.max(min, current + step);
    });
});

function showError(message) {
    errorEl.textContent = message;
    errorEl.style.display = 'block';
}

function hideError() {
    errorEl.style.display = 'none';
}

function showToast(message) {
    clearTimeout(toastTimeout);
    toastMessageEl.textContent = message;
    toastEl.classList.add('toast--visible');
    toastTimeout = setTimeout(function() {
        toastEl.classList.remove('toast--visible');
    }, 1500);
}

// Общая логика сохранения для обеих кнопок ("сохранить и добавить ещё" и
// "готово") и для submit формы — сама показывает ошибку и возвращает
// true/false, чтобы вызывающий код решал, что делать дальше.
async function saveSet() {
    if (isSaving) {
        return false;
    }

    const weight = parseFloat(weightInput.value);
    const reps = parseInt(repsInput.value, 10);

    if (!weight || !reps) {
        showError('Укажите вес и количество повторов');
        return false;
    }

    if (dateInput.value > todayISO) {
        showError('Нельзя выбрать дату позже сегодняшней');
        return false;
    }

    isSaving = true;
    try {
        await addSet(machineId, weight, reps, dateInput.value);
        hideError();
        return true;
    } catch (error) {
        showError(error.message);
        return false;
    } finally {
        isSaving = false;
    }
}

document.querySelector('#save-and-add-btn').addEventListener('click', async function() {
    if (await saveSet()) {
        showToast('Подход записан');
    }
});

function finish() {
    showToast('Подход записан');
    setTimeout(function() {
        window.location.href = 'machine?id=' + machineId;
    }, 600);
}

document.querySelector('#done-btn').addEventListener('click', async function() {
    if (await saveSet()) {
        finish();
    }
});

form.addEventListener('submit', async function(event) {
    event.preventDefault();
    if (await saveSet()) {
        finish();
    }
});

async function loadMachine() {
    try {
        const machine = await getMachine(machineId);
        document.querySelector('#machine-name').textContent = machine.name;
    } catch (error) {
        console.error(error);
    }
}

function formatDateDisplay(dateString) {
    if (dateString === todayISO) {
        return 'Сегодня, ' + formatDate(dateString);
    }
    return formatDate(dateString);
}

dateInput.addEventListener('change', function() {
    if (dateInput.value) {
        dateDisplay.textContent = formatDateDisplay(dateInput.value);
    }
});

// Клик по обычной (текстовой) части нативного date-инпута просто ставит
// курсор в сегмент даты — попап-календарь открывает только клик строго по
// иконке-календарику. Наш инпут растянут на весь блок и скрыт, так что
// без showPicker() календарь почти нигде не открывался бы кликом.
dateInput.addEventListener('click', function() {
    if (typeof dateInput.showPicker === 'function') {
        dateInput.showPicker();
    }
});

dateInput.max = todayISO;
dateInput.value = todayISO;
dateDisplay.textContent = formatDateDisplay(todayISO);

loadMachine()