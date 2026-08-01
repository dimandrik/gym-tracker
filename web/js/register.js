const form = document.querySelector('.auth-form');

form.addEventListener('submit', async function(event) {
    event.preventDefault();
    const email = document.querySelector('#email').value;
    const password = document.querySelector('#password').value;
    let res = validatePassword(password);
    if (res !== null) {
        console.log(res);
        const errorEl = document.querySelector('#error-message');
        errorEl.textContent = res;
        errorEl.style.display = 'block';
    } else {
        try {
            const data = await registerUser(email, password);
            console.log('успех:', data);
            localStorage.setItem('token', data.token);
            window.location.href = 'index.html';
        } catch (error) {
        console.error('Ошибка регистрации:', error);
        const errorEl = document.querySelector('#error-message');
        if (error.status === 409) {
            errorEl.textContent = 'Этот email уже зарегистрирован';
        } else {
            errorEl.textContent = error.message;
        }
        errorEl.style.display = 'block';
    }
    }
})



// Правила пароля проверяются только здесь, на бэкенде (Register-хендлер)
// валидации пароля нет вообще — так что это единственная защита от слабых
// паролей, если понадобится обойти фронт, проверка не сработает.
function validatePassword(password) {
    if (password.length < 8) {
        return "Пароль должен быть не короче 8 символов";
    }
    if (!/[0-9]/.test(password)) {
        return "Пароль должен содержать хотя бы одну цифру";
    }
    if (!/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
        return 'Пароль должен содержать хотя бы один спецсимвол';
    }
    return null;
}
