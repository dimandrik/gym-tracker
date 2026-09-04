getToken();

// ===== Профиль и аватар =====

// Фото и буква-заглушка лежат в разметке оба сразу — просто переключаем
// hidden в зависимости от того, есть ли у пользователя photo_url.
function renderAvatar(photoUrl, initial) {
    const photoImg = document.querySelector('#settings-avatar-photo');
    const initialSpan = document.querySelector('#settings-avatar-initial');
    const sheetPhotoImg = document.querySelector('#photo-sheet-avatar-photo');
    const sheetInitialSpan = document.querySelector('#photo-sheet-avatar-initial');

    if (photoUrl) {
        photoImg.src = API_BASE_URL + photoUrl;
        photoImg.hidden = false;
        sheetPhotoImg.src = API_BASE_URL + photoUrl;
        sheetPhotoImg.hidden = false;
        initialSpan.hidden = true;
        sheetInitialSpan.hidden = true;
    } else {
        photoImg.hidden = true;
        sheetPhotoImg.hidden = true;
        initialSpan.hidden = false;
        sheetInitialSpan.hidden = false;
    }

    initialSpan.textContent = initial;
    sheetInitialSpan.textContent = initial;
}

async function loadProfile() {
    try {
        const profile = await getProfile();

        const fullName = profile.first_name + ' ' + profile.last_name;
        const initial = profile.first_name.charAt(0).toUpperCase();

        document.querySelector('#settings-name').textContent = fullName;
        document.querySelector('#settings-email').textContent = profile.email;
        document.querySelector('#photo-sheet-name').textContent = fullName;

        renderAvatar(profile.photo_url, initial);

    } catch (error) {
        console.error(error);
    }
}

loadProfile();

// ===== Подтверждение выхода =====

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
    window.location.href = 'login.html';
});

// ===== Шторка выбора источника фото профиля =====

const photoSheet = document.querySelector('#photo-sheet');
const photoSheetBackdrop = document.querySelector('#photo-sheet-backdrop');

function openPhotoSheet() {
    photoSheet.hidden = false;
    requestAnimationFrame(function() {
        photoSheet.classList.add('confirm-modal--visible');
    });
}

function closePhotoSheet() {
    photoSheet.classList.remove('confirm-modal--visible');
    setTimeout(function() {
        photoSheet.hidden = true;
    }, 220);
}

document.querySelector('#settings-avatar').addEventListener('click', openPhotoSheet);
document.querySelector('#photo-sheet-cancel').addEventListener('click', closePhotoSheet);
photoSheetBackdrop.addEventListener('click', closePhotoSheet);

// Два отдельных file-инпута, а не один: у первого capture="environment"
// сразу открывает камеру, второй без capture — обычный выбор из галереи.
const photoCameraInput = document.querySelector('#photo-camera-input');
const photoGalleryInput = document.querySelector('#photo-gallery-input');

document.querySelector('#photo-take-btn').addEventListener('click', function() {
    closePhotoSheet();
    photoCameraInput.click();
});
document.querySelector('#photo-gallery-btn').addEventListener('click', function() {
    closePhotoSheet();
    photoGalleryInput.click();
});
document.querySelector('#photo-delete-btn').addEventListener('click', function() {
    closePhotoSheet();
    deleteCurrentPhoto();
});

// ===== Загрузка выбранного фото =====

async function uploadPhoto(file) {
    const formData = new FormData();
    formData.append('photo', file);

    try {
        const result = await updatePhoto(formData);
        renderAvatar(result.photo_url, document.querySelector('#settings-avatar-initial').textContent);
    } catch (error) {
        console.error(error);
        alert('Не удалось сохранить фото');
    }
}

function onPhotoInputChange(event) {
    const file = event.target.files[0];
    event.target.value = '';
    if (file) {
        uploadPhoto(file);
    }
}

photoCameraInput.addEventListener('change', onPhotoInputChange);
photoGalleryInput.addEventListener('change', onPhotoInputChange);

async function deleteCurrentPhoto() {
    if (!confirm('Удалить фото профиля?')) {
        return;
    }
    try {
        await deletePhoto();
        renderAvatar('', document.querySelector('#settings-avatar-initial').textContent);
    } catch (error) {
        console.error(error);
        alert('Не удалось удалить фото');
    }
}
