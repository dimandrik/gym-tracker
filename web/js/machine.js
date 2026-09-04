getToken();


let isEditMode = false;
let lastGroupedSets = {};
let pendingDeleteSetId = null;

const editToggleBtn = document.querySelector('#edit-toggle-btn');
const addSetLink = document.querySelector('#add-set-link');
const editMachineLink = document.querySelector('#edit-machine-link');
const chartsLink = document.querySelector('#charts-link');
const historyList = document.querySelector('#history-list');
const confirmModal = document.querySelector('#delete-confirm-modal');
const confirmModalDesc = document.querySelector('#confirm-modal-desc');
const confirmDeleteBtn = document.querySelector('#confirm-delete-btn');
const confirmCancelBtn = document.querySelector('#confirm-cancel-btn');
const confirmModalBackdrop = document.querySelector('#confirm-modal-backdrop');
const machineId = new URLSearchParams(window.location.search).get('id')

async function loadMachine() {
    try {
        const machine = await getMachine(machineId);

        document.querySelector('#machine-photo').src = API_BASE_URL + machine.photo_url;
        document.querySelector('#machine-photo').alt = machine.name;
        document.querySelector('#machine-name').textContent = machine.name;
        addSetLink.href = 'add-set.html?machine_id=' + machine.id;
        editMachineLink.href = 'edit-machine.html?id=' + machine.id;
        chartsLink.href = 'charts.html?machine_id=' + machine.id;

        const sets = await getSets(machineId);

        const record = findRecord(sets);
        if (record) {
            document.querySelector('#machine-record').textContent = 'рекорд · ' + record.weight_kg + ' кг × ' + record.reps;
        } else {
            document.querySelector('#machine-record').textContent = 'Нет данных';
        }

        const grouped = groupSetsByDate(sets);
        lastGroupedSets = grouped;
        renderHistory(grouped);

    } catch (error) {
        console.error(error);
    }
}

loadMachine();


// Бэкенд отдаёт плоский список сетов — группируем по дате тренировки
// на клиенте, чтобы отрисовать историю днями, как в дизайне.
function groupSetsByDate(sets) {
    const groups = {};

    sets.forEach(function(set) {
        const date = set.workout_date.split('T')[0];

        if (!groups[date]) {
            groups[date] = [];
        }
        groups[date].push(set);
    });

    return groups;
}

function renderHistory(groupedSets) {
    historyList.innerHTML = '';

    const dates = Object.keys(groupedSets);
    dates.sort();
    dates.reverse();

    dates.forEach(function(date) {
        const dayBlock = createDayBlock(date, groupedSets[date]);
        historyList.appendChild(dayBlock);
    });
}

function createDayBlock(date, sets) {
    const block = document.createElement('div');
    block.className = 'history-day';

    const header = document.createElement('div');
    header.className = 'history-day__header';

    const dateEl = document.createElement('span');
    dateEl.className = 'history-day__date';
    dateEl.textContent = formatDate(date);

    const countEl = document.createElement('span');
    countEl.className = 'history-day__count';
    countEl.textContent = sets.length + ' подхода';

    header.appendChild(dateEl);
    header.appendChild(countEl);
    block.appendChild(header);

    const setsContainer = document.createElement('div');
    setsContainer.className = 'history-day__sets';

    sets.forEach(function(set) {
        const row = document.createElement('div');
        row.className = 'history-set history-set--editable';

        // Кнопка удаления всегда в DOM — раскрывается через CSS-transition
        // при входе в режим изменения, а не появляется рывком по innerHTML.
        const deleteBtn = document.createElement('button');
        deleteBtn.className = 'history-set__delete-btn';
        deleteBtn.textContent = '−';
        deleteBtn.setAttribute('aria-label', 'Удалить подход');
        deleteBtn.addEventListener('click', function(event) {
            event.stopPropagation();
            if (!isEditMode) {
                return;
            }
            openDeleteConfirm(set);
        });
        row.appendChild(deleteBtn);

        const label = document.createElement('span');
        label.className = 'history-set__label';
        label.textContent = 'Подход ' + set.set_number;

        const value = document.createElement('span');
        value.className = 'history-set__value';
        value.textContent = set.weight_kg + ' кг × ' + set.reps;

        row.appendChild(label);
        row.appendChild(value);

        row.addEventListener('click', function() {
            if (!isEditMode) {
                return;
            }
            window.location.href = 'edit-set.html?id=' + set.id;
        });

        setsContainer.appendChild(row);
    });

    block.appendChild(setsContainer);

    return block;
}

editToggleBtn.addEventListener('click', function() {
    isEditMode = !isEditMode;
    editToggleBtn.classList.toggle('machine-screen__edit-btn--active', isEditMode);
    editToggleBtn.querySelector('span').textContent = isEditMode ? 'готово' : 'изменить';
    historyList.classList.toggle('history-list--edit-mode', isEditMode);
    addSetLink.classList.toggle('machine-screen__add-btn--disabled', isEditMode);
})

// Рекорд — просто подход с максимальным весом за всё время на этом
// тренажёре; при равном весе повторы не учитываются.
function findRecord(sets) {
    if (sets.length === 0) {
        return null;
    }

    let record = sets[0];
    sets.forEach(function(set) {
        if (set.weight_kg > record.weight_kg) {
            record = set;
        }
    });

    return record;
}

function openDeleteConfirm(set) {
    pendingDeleteSetId = set.id;
    confirmModalDesc.innerHTML = 'Подход ' + set.set_number + ' · ' + set.weight_kg + ' кг × ' + set.reps +
        '<br>' + formatDate(set.workout_date.split('T')[0]) + ' — действие необратимо';

    confirmModal.hidden = false;
    // hidden -> visible нужен один кадр между снятием [hidden] и добавлением
    // класса, иначе transition не запустится (браузер схлопнёт оба изменения).
    requestAnimationFrame(function() {
        confirmModal.classList.add('confirm-modal--visible');
    });
}

function closeDeleteConfirm() {
    pendingDeleteSetId = null;
    confirmModal.classList.remove('confirm-modal--visible');
    setTimeout(function() {
        confirmModal.hidden = true;
    }, 220);
}

confirmCancelBtn.addEventListener('click', closeDeleteConfirm);
confirmModalBackdrop.addEventListener('click', closeDeleteConfirm);

confirmDeleteBtn.addEventListener('click', async function() {
    const setId = pendingDeleteSetId;
    closeDeleteConfirm();
    await handleDelete(setId);
});

async function handleDelete(setId) {
    try {
        await deleteSet(setId);
        await loadMachine();
    } catch (error) {
        alert('Не удалось удалить подход: ' + error.message);
    }
}