getToken();

const modal = document.querySelector('#delete-account-modal');
const backdrop = document.querySelector('#delete-account-backdrop');
const passwordInput = document.querySelector('#delete-password');
const errorEl = document.querySelector('#delete-error');

async function loadAccountInfo() {
    try {
        const profile = await getProfile();

        document.querySelector('#current-name').textContent = profile.first_name + ' ' + profile.last_name;
        document.querySelector('#current-email').textContent = profile.email;

    } catch (error) {
        console.error(error);
    }
}

loadAccountInfo();

function openModal() {
    modal.hidden = false;
    requestAnimationFrame(function() {
        modal.classList.add('confirm-modal--visible');
    });
}

function closeModal() {
    modal.classList.remove('confirm-modal--visible');
    passwordInput.value = '';
    errorEl.style.display = 'none';
    setTimeout(function() {
        modal.hidden = true;
    }, 220);
}

document.querySelector('#delete-account-btn').addEventListener('click', openModal);
document.querySelector('#cancel-delete-account-btn').addEventListener('click', closeModal);
backdrop.addEventListener('click', closeModal);

document.querySelector('#confirm-delete-account-btn').addEventListener('click', async function() {
    const password = passwordInput.value;
    if (!password) {
        errorEl.textContent = 'Введите пароль';
        errorEl.style.display = 'block';
        return;
    }

    try {
        await deleteAccount(password);
        localStorage.removeItem('token');
        window.location.href = 'login';
    } catch (error) {
        if (error.status === 401) {
            errorEl.textContent = 'Неверный пароль';
        } else {
            errorEl.textContent = 'Не удалось удалить аккаунт. Попробуйте ещё раз';
        }
        errorEl.style.display = 'block';
    }
});