getToken();

async function loadProfile() {
    try {
        const profile = await getProfile();

        document.querySelector('#first-name').value = profile.first_name;
        document.querySelector('#last-name').value = profile.last_name;

    } catch (error) {
        console.error(error);
    }
}

loadProfile();

document.querySelector('#change-name-form').addEventListener('submit', async function(event) {
    event.preventDefault();

    const firstName = document.querySelector('#first-name').value;
    const lastName = document.querySelector('#last-name').value;
    const errorEl = document.querySelector('#error-message');

    if (!firstName || !lastName) {
        errorEl.textContent = 'Введите имя и фамилию';
        errorEl.style.display = 'block';
        return;
    }

    try {
        await updateName(firstName, lastName);
        window.location.href = 'account-security.html';
    } catch (error) {
        if (error.status === 400) {
            errorEl.textContent = 'Введите имя и фамилию';
        } else {
            errorEl.textContent = 'Не удалось изменить имя. Попробуйте ещё раз';
        }
        errorEl.style.display = 'block';
    }
});