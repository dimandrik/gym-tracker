// Форматирует дату из "YYYY-MM-DD" в "18 июля" — общий хелпер для истории
// тренировок и страницы тренажёра. Разбор строки вручную, а не через
// new Date(), чтобы не словить сдвиг часового пояса.
function formatDate(dateString) {
    const months = ['января', 'февраля', 'марта', 'апреля', 'мая', 'июня', 'июля', 'августа', 'сентября', 'октября', 'ноября', 'декабря'];
    const parts = dateString.split('-');
    const day = parseInt(parts[2]);
    const month = months[parseInt(parts[1]) - 1];
    return day + ' ' + month;
}