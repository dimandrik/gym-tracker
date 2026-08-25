const form = document.querySelector('.auth-form');

form.addEventListener('submit', async function(event) {
    event.preventDefault();

    const email = document.querySelector('#email').value;
    const password = document.querySelector('#password').value;
    const errorEl = document.querySelector('#error-message');

    try {
        const data = await loginUser(email, password);
        console.log('успех:', data);
        // токен живёт в localStorage — его же читает getToken() на всех
        // остальных страницах
        localStorage.setItem('token', data.token);
        window.location.href = 'index';
    } catch (error) {
        console.log('ошибка:', error.message);
        if (error.status === 401) {
            errorEl.textContent = 'Неверный логин или пароль';
        } else {
            errorEl.textContent = error.message;
        }
        errorEl.style.display = 'block';
    }
});