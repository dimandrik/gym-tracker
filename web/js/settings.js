getToken();

async function loadProfile() {
    try {
        const profile = await getProfile();

        document.querySelector('#settings-name').textContent = profile.first_name + ' ' + profile.last_name;
        document.querySelector('#settings-email').textContent = profile.email;
        document.querySelector('#settings-avatar').textContent = profile.first_name.charAt(0).toUpperCase();

    } catch (error) {
        console.error(error);
    }
}

loadProfile();

document.querySelector('#logout-btn').addEventListener('click', function() {
    if (!confirm('Выйти из аккаунта?')) {
        return;
    }
    localStorage.removeItem('token');
    window.location.href = 'login.html';
});