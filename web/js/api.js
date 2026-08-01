// Простая проверка "залогинен ли пользователь" — вызывается в начале каждой
// защищённой страницы, при отсутствии токена сразу редиректит на логин.
function getToken() {
    const token = localStorage.getItem('token');
    if (!token) {
        window.location.href = 'login.html';
    }
    return token;
}

// Обёртка над fetch для всех запросов к API, кроме логина/регистрации:
// сама подставляет Bearer-токен и кидает Error с текстом ответа при неудаче.
async function apiFetch(path, options) {
    const token = getToken();
    options = options || {};
    options.headers = options.headers || {};
    options.headers['Authorization'] = 'Bearer ' + token;

    const response = await fetch(API_BASE_URL + path, options);

    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText);
    }

    // DELETE/PUT respond 204 No Content — no body to parse
    if (response.status === 204) {
        return null;
    }

    return response.json();
}

async function getMachine(id) {
    return apiFetch('/api/machines/' + id);
}

async function getSets(machineId) {
    return apiFetch('/api/sets?machine_id=' + machineId);
}

async function addSet(machineId, weightKg, reps) {
    return apiFetch('/api/sets', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ machine_id: machineId, weight_kg: weightKg, reps: reps })
    });
}

async function getWorkoutHistory() {
    return apiFetch('/api/workouts/history');
}

async function deleteSet(id) {
    return apiFetch('/api/sets/' + id, { method: 'DELETE' });
}

async function getSet(id) {
    return apiFetch('/api/sets/' + id);
}

async function updateSet(id, weightKg, reps) {
    return apiFetch('/api/sets/' + id, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ weight_kg: weightKg, reps: reps })
    });
}