let API_BASE_URL;

if (window.location.hostname === 'gym-tracker.dev') {
    API_BASE_URL = window.location.origin;
} else {
    API_BASE_URL = 'http://' + window.location.hostname + ':8080';
}