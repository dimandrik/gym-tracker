// Отдельный от apiFetch хелпер, потому что для логина/регистрации токена
// ещё нет и Authorization-заголовок не нужен.
async function postJSON(path, data) {
    const response = await fetch(API_BASE_URL + path, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    });

    if (!response.ok) {
        const errorText = await response.text();
        const error = new Error(errorText);
        error.status = response.status;
        throw error;
    }

    return response.json();
}

async function loginUser(email, password) {
    return postJSON('/api/login', { email: email, password: password });
}

async function registerUser(email, password, firstName, lastName) {
    return postJSON('/api/register', {
        email: email,
        password: password,
        first_name: firstName,
        last_name: lastName
    });
}