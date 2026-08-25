// Простая проверка "залогинен ли пользователь" — вызывается в начале каждой
// защищённой страницы, при отсутствии токена сразу редиректит на логин.
function getToken() {
    const token = localStorage.getItem('token');
    if (!token) {
        window.location.href = 'login';
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
        const error = new Error(errorText);
        error.status = response.status;
        throw error;
    }

    // DELETE/PUT отвечают 204 без тела — парсить нечего
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

async function addSet(machineId, weightKg, reps, date) {
    return apiFetch('/api/sets', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ machine_id: machineId, weight_kg: weightKg, reps: reps, date: date || '' })
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

async function updateMachine(id, formData) {
    return apiFetch('/api/machines/' + id, {
        method: 'PUT',
        body: formData
    });
}

async function deleteMachine(id) {
    return apiFetch('/api/machines/' + id, { method: 'DELETE' });
}

async function getSetsByDate(date) {
    return apiFetch('/api/workouts/day?date=' + date);
}

async function getProfile() {
    return apiFetch('/api/user/profile');
}

async function deleteAccount(password) {
    return apiFetch('/api/user/account', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password: password })
    });
}

async function updateName(firstName, lastName) {
    return apiFetch('/api/user/name', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ first_name: firstName, last_name: lastName })
    });
}

async function updateEmail(newEmail, currentPassword) {
    return apiFetch('/api/user/email', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ new_email: newEmail, current_password: currentPassword })
    });
}

async function updatePassword(currentPassword, newPassword) {
    return apiFetch('/api/user/password', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ current_password: currentPassword, new_password: newPassword })
    });
}