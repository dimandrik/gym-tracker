getToken();

document.querySelector('#change-email-form').addEventListener('submit', async function(event) {
    event.preventDefault();

    const newEmail = document.querySelector('#new-email').value;
    const currentPassword = document.querySelector('#current-password').value;
    const errorEl = document.querySelector('#error-message');

    if (!newEmail || !currentPassword) {
        errorEl.textContent = 'Введите новый email и текущий пароль';
        errorEl.style.display = 'block';
        return;
    }

    try {
        await updateEmail(newEmail, currentPassword);
        window.location.href = 'account-security';
    } catch (error) {
        if (error.status === 400) {
            errorEl.textContent = 'Некорректный формат email';
        } else if (error.status === 401) {
            errorEl.textContent = 'Неверный пароль';
        } else if (error.status === 409) {
            errorEl.textContent = 'Этот email уже используется';
        } else {
            errorEl.textContent = 'Не удалось изменить email. Попробуйте ещё раз';
        }
        errorEl.style.display = 'block';
    }
});