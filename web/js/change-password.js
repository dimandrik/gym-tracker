getToken();

document.querySelector('#change-password-form').addEventListener('submit', async function(event) {
    event.preventDefault();

    const currentPassword = document.querySelector('#current-password').value;
    const newPassword = document.querySelector('#new-password').value;
    const confirmPassword = document.querySelector('#confirm-password').value;
    const errorEl = document.querySelector('#error-message');

    if (!currentPassword || !newPassword || !confirmPassword) {
        errorEl.textContent = 'Введите текущий и новый пароль';
        errorEl.style.display = 'block';
        return;
    }

    if (newPassword !== confirmPassword) {
        errorEl.textContent = 'Новый пароль и подтверждение не совпадают';
        errorEl.style.display = 'block';
        return;
    }

    const passwordError = validatePassword(newPassword);
    if (passwordError !== null) {
        errorEl.textContent = passwordError;
        errorEl.style.display = 'block';
        return;
    }

    try {
        await updatePassword(currentPassword, newPassword);
        window.location.href = 'account-security';
    } catch (error) {
        if (error.status === 401) {
            errorEl.textContent = 'Неверный текущий пароль';
        } else if (error.status === 400) {
            errorEl.textContent = 'Пароль не соответствует требованиям';
        } else {
            errorEl.textContent = 'Не удалось изменить пароль. Попробуйте ещё раз';
        }
        errorEl.style.display = 'block';
    }
});

function validatePassword(password) {
    if (password.length < 8) {
        return 'Пароль должен быть не короче 8 символов';
    }
    if (!/[0-9]/.test(password)) {
        return 'Пароль должен содержать хотя бы одну цифру';
    }
    if (!/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
        return 'Пароль должен содержать хотя бы один спецсимвол';
    }
    return null;
}
