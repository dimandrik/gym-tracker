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

const logoutModal = document.querySelector('#logout-modal');
const logoutBackdrop = document.querySelector('#logout-backdrop');

function openLogoutModal() {
    logoutModal.hidden = false;
    requestAnimationFrame(function() {
        logoutModal.classList.add('confirm-modal--visible');
    });
}

function closeLogoutModal() {
    logoutModal.classList.remove('confirm-modal--visible');
    setTimeout(function() {
        logoutModal.hidden = true;
    }, 220);
}

document.querySelector('#logout-btn').addEventListener('click', openLogoutModal);
document.querySelector('#cancel-logout-btn').addEventListener('click', closeLogoutModal);
logoutBackdrop.addEventListener('click', closeLogoutModal);

document.querySelector('#confirm-logout-btn').addEventListener('click', function() {
    localStorage.removeItem('token');
    window.location.href = 'login';
});