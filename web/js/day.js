getToken();

const date = new URLSearchParams(window.location.search).get('date');

function groupByMachine(entries) {
    const groups = {};

    entries.forEach(function(entry) {
        const machineId = entry.machine_id;

        if (!groups[machineId]) {
            groups[machineId] = {
                machineName: entry.machine_name,
                machinePhotoUrl: entry.machine_photo_url,
                sets: []
            };
        }
        groups[machineId].sets.push(entry);
    });

    return groups;
}

function createMachineBlock(machineId, group) {
    const block = document.createElement('div');
    block.className = 'day-machine-block';

    const header = document.createElement('div');
    header.className = 'day-machine-block__header';

    const photo = document.createElement('img');
    photo.className = 'day-machine-block__photo';
    photo.src = API_BASE_URL + group.machinePhotoUrl;
    photo.alt = group.machineName;

    const nameEl = document.createElement('p');
    nameEl.className = 'day-machine-block__name';
    nameEl.textContent = group.machineName;

    header.appendChild(photo);
    header.appendChild(nameEl);
    block.appendChild(header);

    group.sets.forEach(function(set) {
        const row = document.createElement('div');
        row.className = 'history-set';

        const label = document.createElement('span');
        label.className = 'history-set__label';
        label.textContent = 'Подход ' + set.set_number;

        const value = document.createElement('span');
        value.className = 'history-set__value';
        value.textContent = set.weight_kg + ' кг × ' + set.reps;

        row.appendChild(label);
        row.appendChild(value);
        block.appendChild(row);
    });

    return block;
}

function renderDay(grouped) {
    const dayList = document.querySelector('#day-list');
    dayList.innerHTML = '';

    Object.keys(grouped).forEach(function(machineId) {
        const block = createMachineBlock(machineId, grouped[machineId]);
        dayList.appendChild(block);
    });
}

async function loadDay() {
    try {
        const entries = await getSetsByDate(date);

        document.querySelector('#day-title').textContent = formatDate(date);

        const totalSets = entries.length;

        const grouped = groupByMachine(entries);
        const machinesCount = Object.keys(grouped).length;
        document.querySelector('#day-meta').textContent = machinesCount + ' тренажёра · ' + totalSets + ' подхода';

        renderDay(grouped);

    } catch (error) {
        console.error(error);
    }
}

loadDay();