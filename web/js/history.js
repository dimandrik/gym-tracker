getToken()



// Сравнение по строке "YYYY-MM-DD", чтобы не связываться с часовыми
// поясами при сравнении Date-объектов.
function isToday(dateString) {
    const today = new Date().toISOString().split('T')[0];
    const date = dateString.split('T')[0];
    return today === date;
}

async function loadHistory() {
    const history = await getWorkoutHistory();
    const container = document.querySelector('#workout-history');

    history.forEach(function(day) {
        const card = createDayCard(day);
        container.appendChild(card);
    });
}

function createDayCard(day) {
    const card = document.createElement('div');
    card.className = 'workout-day-card';

    const dateEl = document.createElement('p');
    dateEl.className = 'workout-day-card__date';

    const formattedDate = formatDate(day.workout_date.split('T')[0]);
    if (isToday(day.workout_date)) {
        dateEl.textContent = 'Сегодня, ' + formattedDate;
    } else {
        dateEl.textContent = formattedDate;
    }

    const summaryEl = document.createElement('p');
    summaryEl.className = 'workout-day-card__summary';
    summaryEl.textContent = day.machines_count + ' тренажёра · ' + day.sets_count + ' подходов';

    card.appendChild(dateEl);
    card.appendChild(summaryEl);

    return card;
}

loadHistory()