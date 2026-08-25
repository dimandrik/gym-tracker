getToken();

async function getMachines() {
    return apiFetch('/api/machines');
}

function createMachineCard(machine) {
    const card = document.createElement('a');
    card.className = 'machine-card';
    card.href = 'machine?id=' + machine.id;

    const img = document.createElement('img');
    img.className = 'machine-card__photo';
    img.src = API_BASE_URL + machine.photo_url;
    img.alt = machine.name;

    const name = document.createElement('p');
    name.className = 'machine-card__name';
    name.textContent = machine.name;

    card.appendChild(img);
    card.appendChild(name);

    return card;
}

async function loadMachines() {
    const grid = document.querySelector('#machines-grid');
    try {
        const machines = await getMachines();

        if (machines.length === 0) {
            grid.textContent = "У вас пока нет тренажёров — добавьте первый!"
            return;
        }

        machines.forEach(function(machine) {
            const card = createMachineCard(machine);
            grid.appendChild(card);
        });
    } catch (error) {
        grid.textContent = 'Не удалось загрузить тренажёры';
    }
}

loadMachines();