getToken();

// ===== Состояние страницы =====

const machineId = new URLSearchParams(window.location.search).get('machine_id');
const SVG_NS = 'http://www.w3.org/2000/svg';

let allSets = []; // все подходы по тренажёру, загружаются один раз и переиспользуются во всех вкладках
let setIdToCircle = new Map(); // id подхода -> его точка на графике, чтобы подсветить по клику из списка снизу
let highlightedCircle = null; // точка, подсвеченная сейчас, чтобы снять подсветку перед новой
let activeTab = 'scatter'; // текущая вкладка графика, нужна чтобы перерисовать её при смене периода
let selectedRange = 'month'; // выбранный период сверху графика, по умолчанию как в дизайне

// Сколько дней назад начинается период — используется для фильтрации подходов
const RANGE_DAYS = {
    week: 7,
    month: 30,
    '3m': 90,
    year: 365
};

// Оставляет только подходы, попадающие в выбранный период. 'all' — без фильтра
function filterSetsByRange(sets, range) {
    if (range === 'all') {
        return sets;
    }

    const cutoff = Date.now() - RANGE_DAYS[range] * 24 * 60 * 60 * 1000;
    return sets.filter(function(set) {
        return new Date(set.workout_date).getTime() >= cutoff;
    });
}

// ===== Элементы модалки с деталями точки =====

const pointDetailModal = document.querySelector('#point-detail-modal');
const pointDetailBackdrop = document.querySelector('#point-detail-backdrop');
const pointDetailTitle = document.querySelector('#point-detail-title');
const pointDetailList = document.querySelector('#point-detail-list');
const pointDetailCloseBtn = document.querySelector('#point-detail-close-btn');

// ===== Загрузка данных =====

// Загружает тренажёр и все его подходы, заполняет шапку и рисует первую вкладку
async function loadData() {
    allSets = await getSets(machineId);
    const machine = await getMachine(machineId);
    document.querySelector('#chart-subtitle').textContent = machine.name;
    document.querySelector('#back-link').href = 'machine.html?id=' + machineId;

    renderTab('scatter');
}

loadData();

// ===== Вкладки графиков =====

// Отрисовывает содержимое выбранной вкладки в #chart-content с учётом выбранного периода
function renderTab(tabName) {
    activeTab = tabName;

    const content = document.querySelector('#chart-content');
    content.innerHTML = '';

    const filteredSets = filterSetsByRange(allSets, selectedRange);

    if (tabName === 'scatter') {
        content.appendChild(renderScatterChart(filteredSets));
        content.appendChild(renderRecentSetsCard(filteredSets));
    } else if (tabName === 'progress') {
        content.appendChild(renderProgressChart(filteredSets));
        content.appendChild(renderProgressRecordsCard(filteredSets));
    } else if (tabName === 'volume') {
        content.appendChild(renderVolumeChart(filteredSets));
        content.appendChild(renderVolumeRecordsCard(filteredSets));
    } else if (tabName === 'frequency') {
        content.appendChild(renderFrequencyChart(filteredSets));
        content.appendChild(renderFrequencyRecordsCard(filteredSets));
    }
}

// Переключение активной вкладки по клику
document.querySelectorAll('.chart-tab').forEach(function(tab) {
    tab.addEventListener('click', function() {
        document.querySelectorAll('.chart-tab').forEach(function(t) {
            t.classList.remove('chart-tab--active');
        });
        tab.classList.add('chart-tab--active');
        renderTab(tab.dataset.tab);
    });
});

// Переключение периода (неделя/месяц/...) над графиком — перерисовывает текущую вкладку
document.querySelectorAll('.chart-range-tab').forEach(function(tab) {
    tab.addEventListener('click', function() {
        document.querySelectorAll('.chart-range-tab').forEach(function(t) {
            t.classList.remove('chart-range-tab--active');
        });
        tab.classList.add('chart-range-tab--active');
        selectedRange = tab.dataset.range;
        renderTab(activeTab);
    });
});

// ===== Числовые и цветовые утилиты =====

// Линейно переводит value из диапазона [domainMin, domainMax] в диапазон [rangeMin, rangeMax]
function scale(value, domainMin, domainMax, rangeMin, rangeMax) {
    const ratio = (value - domainMin) / (domainMax - domainMin);
    return rangeMin + ratio * (rangeMax - rangeMin);
}

// Смешивает два RGB-цвета по коэффициенту ratio (0 — colorStart, 1 — colorEnd)
function interpolateColor(ratio, colorStart, colorEnd) {
    const r = Math.round(colorStart[0] + ratio * (colorEnd[0] - colorStart[0]));
    const g = Math.round(colorStart[1] + ratio * (colorEnd[1] - colorStart[1]));
    const b = Math.round(colorStart[2] + ratio * (colorEnd[2] - colorStart[2]));
    return 'rgb(' + r + ',' + g + ',' + b + ')';
}

// Возвращает цвет точки по дате: чем раньше дата в диапазоне allDates — тем ближе к синему,
// чем позже — тем ближе к красному. Так на карте подходов видно, где старые тренировки, а где недавние
function getPointColor(setDate, allDates) {
    const minDate = Math.min(...allDates);
    const maxDate = Math.max(...allDates);

    let ratio = 0.5;
    if (maxDate > minDate) {
        ratio = scale(setDate, minDate, maxDate, 0, 1);
    }

    const blue = [59, 111, 224];
    const red = [229, 72, 77];
    return interpolateColor(ratio, blue, red);
}

// ===== SVG-утилиты =====

// Создаёт SVG-элемент нужного тега с заданными атрибутами
function createSVGElement(tag, attrs) {
    const el = document.createElementNS(SVG_NS, tag);
    for (const key in attrs) {
        el.setAttribute(key, attrs[key]);
    }
    return el;
}

// ===== Карта подходов (scatter chart) =====

// Несколько подходов часто совпадают по весу и повторениям — если рисовать
// каждый отдельной точкой, они лягут друг на друга и часть данных пропадёт
// с графика. Группируем такие подходы и рисуем одну точку размером побольше.
function groupSetsByWeightAndReps(sets) {
    const groups = new Map();

    sets.forEach(function(set) {
        const key = set.weight_kg + '_' + set.reps;
        if (!groups.has(key)) {
            groups.set(key, {
                weight_kg: set.weight_kg,
                reps: set.reps,
                dates: [],
                sets: [],
                count: 0
            });
        }
        const group = groups.get(key);
        group.dates.push(new Date(set.workout_date).getTime());
        group.sets.push(set);
        group.count += 1;
    });

    return Array.from(groups.values());
}

// Строит карточку с графиком «вес × повторения»: точка — один подход (или
// группа совпавших подходов), цвет точки — насколько давно он был выполнен
function renderScatterChart(sets) {
    const container = document.createElement('div');
    container.className = 'chart-card';

    const title = document.createElement('div');
    title.className = 'chart-card__title';
    title.textContent = 'Карта подходов';
    container.appendChild(title);

    const desc = document.createElement('div');
    desc.className = 'chart-card__desc';
    desc.textContent = 'Точка = один подход. Видно, что вы выбирали: больше веса или больше повторений';
    container.appendChild(desc);

    if (sets.length === 0) {
        const empty = document.createElement('p');
        empty.className = 'chart-card__empty';
        empty.textContent = 'Пока нет данных для графика';
        container.appendChild(empty);
        return container;
    }

    const weights = sets.map(function (s) {
        return s.weight_kg;
    });
    const reps = sets.map(function (s) {
        return s.reps;
    });
    const dates = sets.map(function (s) {
        return new Date(s.workout_date).getTime();
    });

    const weightMin = Math.min(...weights);
    const weightMax = Math.max(...weights);
    const repsMin = Math.min(...reps);
    const repsMax = Math.max(...reps);
    const weightMaxSafe = weightMax > weightMin ? weightMax : weightMin + 1;
    const repsMaxSafe = repsMax > repsMin ? repsMax : repsMin + 1;

    // Отступы области построения внутри viewBox 300×320. Левый и правый
    // симметричны (46), чтобы график не «съезжал» в сторону, и с запасом,
    // чтобы крупные точки-группы не обрезались по краям. viewBox выше, чем
    // у остальных графиков вкладки (226) — у карты подходов две значащие
    // оси одновременно, и близкие по весу/повторениям подходы иначе легко
    // накладываются друг на друга; больше высоты — больше пикселей на
    // единицу повторений, точки лучше разносятся по вертикали
    const padLeft = 46, padRight = 254, padTop = 20, padBottom = 280;

    const svg = createSVGElement('svg', {
        viewBox: '0 0 300 320',
        style: 'width:100%;height:300px'
    });

    // Горизонтальные линии сетки и подписи оси Y (повторения)
    const gridLinesCount = 4;
    for (let i = 0; i <= gridLinesCount; i++) {
        const repsValue = repsMin + (i / gridLinesCount) * (repsMaxSafe - repsMin);
        const y = scale(repsValue, repsMin, repsMaxSafe, padBottom, padTop);

        const gridLine = createSVGElement('line', {
            x1: padLeft, y1: y, x2: padRight, y2: y,
            stroke: '#241f18'
        });
        svg.appendChild(gridLine);

        const label = createSVGElement('text', {
            x: padLeft - 6, y: y + 3,
            fill: '#6f675c', 'font-size': 9,
            'text-anchor': 'end'
        });
        label.textContent = Math.round(repsValue);
        svg.appendChild(label);
    }

    // Ось X (вес) и подписи значений веса вдоль неё
    const axisLine = createSVGElement('line', {
        x1: padLeft, y1: padBottom, x2: padRight, y2: padBottom,
        stroke: '#241f18'
    });
    svg.appendChild(axisLine);

    const weightLabelsCount = 4;
    for (let i = 0; i <= weightLabelsCount; i++) {
        const weightValue = weightMin + (i / weightLabelsCount) * (weightMaxSafe - weightMin);
        const x = scale(weightValue, weightMin, weightMaxSafe, padLeft, padRight);

        const label = createSVGElement('text', {
            x: x, y: padBottom + 14,
            fill: '#6f675c', 'font-size': 9,
            'text-anchor': 'middle'
        });
        label.textContent = Math.round(weightValue) + ' кг';
        svg.appendChild(label);
    }

    // Подписи осей целиком («Вес, кг» снизу, «Повторения» повёрнуто слева)
    const xAxisTitle = createSVGElement('text', {
        x: (padLeft + padRight) / 2, y: padBottom + 28,
        fill: '#a49a8c', 'font-size': 10, 'font-weight': 600,
        'text-anchor': 'middle'
    });
    xAxisTitle.textContent = 'Вес, кг';
    svg.appendChild(xAxisTitle);

    const yAxisTitle = createSVGElement('text', {
        x: 12, y: (padTop + padBottom) / 2,
        fill: '#a49a8c', 'font-size': 10, 'font-weight': 600,
        'text-anchor': 'middle',
        transform: 'rotate(-90 12 ' + (padTop + padBottom) / 2 + ')'
    });
    yAxisTitle.textContent = 'Повторения';
    svg.appendChild(yAxisTitle);

    // Сами точки: одна на группу совпавших подходов. Цвет — по средней дате
    // группы, размер — растёт с количеством совпавших подходов, по клику
    // открывается плашка с подробностями по каждому подходу в группе
    const groups = groupSetsByWeightAndReps(sets);

    setIdToCircle = new Map();
    highlightedCircle = null;

    groups.forEach(function(group) {
        const x = scale(group.weight_kg, weightMin, weightMaxSafe, padLeft, padRight);
        const y = scale(group.reps, repsMin, repsMaxSafe, padBottom, padTop);
        const avgDate = group.dates.reduce(function(sum, d) { return sum + d; }, 0) / group.dates.length;
        const color = getPointColor(avgDate, dates);
        const radius = 4 + Math.sqrt(group.count - 1) * 2;

        const circle = createSVGElement('circle', {
            cx: x, cy: y, r: radius,
            fill: color,
            'fill-opacity': group.count > 1 ? 0.85 : 1,
            class: 'chart-point'
        });

        const tooltip = createSVGElement('title', {});
        tooltip.textContent = group.weight_kg + ' кг × ' + group.reps +
            (group.count > 1 ? ' — ' + group.count + ' подхода' : '');
        circle.appendChild(tooltip);

        circle.addEventListener('click', function() {
            openPointDetail(group);
        });

        group.sets.forEach(function(set) {
            setIdToCircle.set(set.id, circle);
        });

        svg.appendChild(circle);

        // Для группы из нескольких подходов показываем их число прямо на точке
        if (group.count > 1) {
            const countLabel = createSVGElement('text', {
                x: x, y: y + 3,
                fill: '#fff', 'font-size': 7, 'font-weight': 700,
                'text-anchor': 'middle',
                'pointer-events': 'none'
            });
            countLabel.textContent = group.count;
            svg.appendChild(countLabel);
        }
    });

    container.appendChild(svg);

    // Легенда снизу: самая старая и самая новая дата по краям цветового градиента
    const oldestDate = sets.reduce(function(min, s) {
        return s.workout_date < min ? s.workout_date : min;
    }, sets[0].workout_date);
    const newestDate = sets.reduce(function(max, s) {
        return s.workout_date > max ? s.workout_date : max;
    }, sets[0].workout_date);

    const legend = document.createElement('div');
    legend.className = 'chart-legend';
    const oldestLabel = document.createElement('span');
    oldestLabel.textContent = formatDate(oldestDate.split('T')[0]);
    const gradient = document.createElement('span');
    gradient.className = 'chart-legend__gradient';
    const newestLabel = document.createElement('span');
    newestLabel.textContent = formatDate(newestDate.split('T')[0]);
    legend.appendChild(oldestLabel);
    legend.appendChild(gradient);
    legend.appendChild(newestLabel);
    container.appendChild(legend);

    return container;
}

// ===== Прогресс по весу (line chart) =====

// Для каждой тренировки берёт самый тяжёлый подход (максимальный вес) —
// так график показывает прогресс по силе, а не смешивает разминочные
// и рабочие подходы. Возвращает дни по возрастанию даты (слева направо на графике)
function computeDailyMaxWeights(sets) {
    const grouped = groupSetsByDate(sets);
    return Object.keys(grouped).sort().map(function(date) {
        const daySets = grouped[date];
        const maxWeight = daySets.reduce(function(max, s) {
            return s.weight_kg > max ? s.weight_kg : max;
        }, daySets[0].weight_kg);
        return {date: date, maxWeight: maxWeight};
    });
}

// Минимальное расстояние в пикселях viewBox между соседними подписями дат,
// при котором подпись вида «05.08» (моноширинный шрифт 9px) ещё не налезает
// на соседнюю
const DATE_LABEL_MIN_GAP = 30;

// Выбирает индексы точек/столбцов, под которыми показываем подпись даты, по
// их фактическим пиксельным координатам xs. Идём от последней точки к первой
// и берём следующую, только если она достаточно далеко от уже выбранной —
// так соседние подписи гарантированно не накладываются, в отличие от выбора
// по номеру индекса, где округление могло свести две подписи вплотную
function pickLabelIndices(xs) {
    const indices = [];
    let lastX = null;

    for (let i = xs.length - 1; i >= 0; i--) {
        if (lastX === null || lastX - xs[i] >= DATE_LABEL_MIN_GAP) {
            indices.push(i);
            lastX = xs[i];
        }
    }

    return new Set(indices);
}

// Короткая подпись даты для оси графика: «05.08» вместо «5 августа» — полное
// название не влезло бы под несколько точек на узком экране, год тоже
// убрали — подписи соседних точек налезали друг на друга
function formatShortDate(dateString) {
    const parts = dateString.split('-');
    const day = parts[2].padStart(2, '0');
    const month = parts[1].padStart(2, '0');
    return day + '.' + month;
}

// Строит карточку с линейным графиком «максимальный вес по тренировкам»:
// одна точка — один день тренировки, последняя точка (текущий рекорд) крупнее и залита акцентом
function renderProgressChart(sets) {
    const container = document.createElement('div');
    container.className = 'chart-card';

    const header = document.createElement('div');
    header.className = 'chart-card__header';

    const titles = document.createElement('div');
    const title = document.createElement('div');
    title.className = 'chart-card__title';
    title.textContent = 'Максимальный вес';
    const desc = document.createElement('div');
    desc.className = 'chart-card__desc';
    desc.style.margin = '0';
    desc.textContent = 'самый тяжёлый подход за тренировку';
    titles.appendChild(title);
    titles.appendChild(desc);
    header.appendChild(titles);

    const daily = computeDailyMaxWeights(sets);

    if (daily.length === 0) {
        container.appendChild(header);
        const empty = document.createElement('p');
        empty.className = 'chart-card__empty';
        empty.textContent = 'Пока нет данных для графика';
        container.appendChild(empty);
        return container;
    }

    // Бейдж со сдвигом веса за выбранный период — от первой до последней тренировки
    if (daily.length > 1) {
        const badge = document.createElement('div');
        badge.className = 'chart-card__badge';
        const change = daily[daily.length - 1].maxWeight - daily[0].maxWeight;
        badge.textContent = (change >= 0 ? '+' : '') + change + ' кг';
        header.appendChild(badge);
    }

    container.appendChild(header);

    const weights = daily.map(function(d) { return d.maxWeight; });
    const weightMin = Math.min(...weights);
    const weightMax = Math.max(...weights);
    const weightMaxSafe = weightMax > weightMin ? weightMax : weightMin + 1;

    // Те же отступы viewBox, что и у карты подходов — графики визуально согласованы между собой
    const padLeft = 46, padRight = 254, padTop = 20, padBottom = 186;

    const svg = createSVGElement('svg', {
        viewBox: '0 0 300 226',
        style: 'width:100%;height:216px'
    });

    // Горизонтальные линии сетки и подписи оси Y (вес)
    const gridLinesCount = 3;
    for (let i = 0; i <= gridLinesCount; i++) {
        const weightValue = weightMin + (i / gridLinesCount) * (weightMaxSafe - weightMin);
        const y = scale(weightValue, weightMin, weightMaxSafe, padBottom, padTop);

        const gridLine = createSVGElement('line', {
            x1: padLeft, y1: y, x2: padRight, y2: y,
            stroke: '#241f18'
        });
        svg.appendChild(gridLine);

        const label = createSVGElement('text', {
            x: padLeft - 6, y: y + 3,
            fill: '#6f675c', 'font-size': 9,
            'text-anchor': 'end'
        });
        label.textContent = Math.round(weightValue);
        svg.appendChild(label);
    }

    const axisLine = createSVGElement('line', {
        x1: padLeft, y1: padBottom, x2: padRight, y2: padBottom,
        stroke: '#241f18'
    });
    svg.appendChild(axisLine);

    // Подписи осей целиком («Вес, кг» повёрнуто слева, «Дата тренировки» снизу)
    const yAxisTitle = createSVGElement('text', {
        x: 12, y: (padTop + padBottom) / 2,
        fill: '#a49a8c', 'font-size': 10, 'font-weight': 600,
        'text-anchor': 'middle',
        transform: 'rotate(-90 12 ' + (padTop + padBottom) / 2 + ')'
    });
    yAxisTitle.textContent = 'Вес, кг';
    svg.appendChild(yAxisTitle);

    const xAxisTitle = createSVGElement('text', {
        x: (padLeft + padRight) / 2, y: 219,
        fill: '#8f877b', 'font-size': 9,
        'text-anchor': 'middle'
    });
    xAxisTitle.textContent = 'ДАТА ТРЕНИРОВКИ';
    svg.appendChild(xAxisTitle);

    // Координаты точек: по оси X — равномерно, по индексу дня (не по самой дате,
    // чтобы редкие тренировки не растягивали график), по оси Y — вес
    const lastIndex = daily.length > 1 ? daily.length - 1 : 1;
    const points = daily.map(function(d, i) {
        return {
            x: scale(i, 0, lastIndex, padLeft, padRight),
            y: scale(d.maxWeight, weightMin, weightMaxSafe, padBottom, padTop),
            date: d.date,
            maxWeight: d.maxWeight
        };
    });

    const polyline = createSVGElement('polyline', {
        points: points.map(function(p) { return p.x + ',' + p.y; }).join(' '),
        fill: 'none',
        stroke: '#2ecbff',
        'stroke-width': 2.5,
        'stroke-linejoin': 'round'
    });
    svg.appendChild(polyline);

    // Подписи дат — только у части точек (иначе они наслаиваются друг на друга)
    const labelIndices = pickLabelIndices(points.map(function(p) { return p.x; }));

    points.forEach(function(p, i) {
        const isLast = i === points.length - 1;

        const circle = createSVGElement('circle', {
            cx: p.x, cy: p.y,
            r: isLast ? 5.5 : 4,
            fill: isLast ? '#2ecbff' : '#14110D',
            stroke: '#2ecbff',
            'stroke-width': isLast ? 0 : 2
        });
        const tooltip = createSVGElement('title', {});
        tooltip.textContent = formatDate(p.date) + ' — ' + p.maxWeight + ' кг';
        circle.appendChild(tooltip);
        svg.appendChild(circle);

        if (labelIndices.has(i)) {
            const label = createSVGElement('text', {
                x: p.x, y: 200,
                fill: isLast ? '#f5f1ea' : '#6f675c', 'font-size': 9,
                'font-family': 'Space Mono,monospace',
                'text-anchor': 'middle'
            });
            label.textContent = formatShortDate(p.date);
            svg.appendChild(label);
        }
    });

    container.appendChild(svg);

    return container;
}

// Строит карточку «Рекорд по тренировкам»: список дней с максимальным весом
// и приростом веса относительно предыдущей тренировки, от новых к старым
function renderProgressRecordsCard(sets) {
    const fragment = document.createDocumentFragment();

    const daily = computeDailyMaxWeights(sets);
    if (daily.length === 0) {
        return fragment;
    }

    const label = document.createElement('div');
    label.className = 'chart-section-label';
    label.textContent = 'Рекорд по тренировкам';
    fragment.appendChild(label);

    const list = document.createElement('div');
    list.className = 'chart-records';

    // Идём от новых тренировок к старым, но прирост считаем относительно
    // предыдущей по времени тренировки — поэтому дельту берём из исходного
    // (хронологического) массива до разворота
    const withDelta = daily.map(function(d, i) {
        const prevWeight = i > 0 ? daily[i - 1].maxWeight : d.maxWeight;
        return {date: d.date, maxWeight: d.maxWeight, delta: d.maxWeight - prevWeight};
    });

    withDelta.slice().reverse().forEach(function(d) {
        const row = document.createElement('div');
        row.className = 'chart-record';

        const dateEl = document.createElement('span');
        dateEl.className = 'chart-record__date';
        dateEl.textContent = formatDate(d.date);

        const weightEl = document.createElement('span');
        weightEl.className = 'chart-record__weight';
        weightEl.textContent = d.maxWeight + ' кг';

        const deltaEl = document.createElement('span');
        deltaEl.className = 'chart-record__delta' + (d.delta > 0 ? ' chart-record__delta--positive' : '');
        deltaEl.textContent = (d.delta > 0 ? '+' : '') + d.delta;

        row.appendChild(dateEl);
        row.appendChild(weightEl);
        row.appendChild(deltaEl);
        list.appendChild(row);
    });

    fragment.appendChild(list);
    return fragment;
}

// ===== Объём тренировки (bar chart) =====

// Для каждой тренировки складывает вес × повторения по всем подходам дня —
// так видно суммарную нагрузку за тренировку, а не отдельные подходы
function computeDailyVolumes(sets) {
    const grouped = groupSetsByDate(sets);
    return Object.keys(grouped).sort().map(function(date) {
        const daySets = grouped[date];
        const volume = daySets.reduce(function(sum, s) {
            return sum + s.weight_kg * s.reps;
        }, 0);
        return {date: date, volume: volume, setsCount: daySets.length};
    });
}

// Строит карточку со столбчатым графиком «объём по тренировкам»: столбец —
// один день тренировки, чем свежее тренировка, тем непрозрачнее столбец,
// последняя (текущая) — залита акцентом полностью и подписана значением
function renderVolumeChart(sets) {
    const container = document.createElement('div');
    container.className = 'chart-card';

    const title = document.createElement('div');
    title.className = 'chart-card__title';
    title.textContent = 'Объём за тренировку';
    container.appendChild(title);

    const desc = document.createElement('div');
    desc.className = 'chart-card__desc';
    desc.textContent = 'вес × повторения, сложенные по всем подходам дня';
    container.appendChild(desc);

    const daily = computeDailyVolumes(sets);

    if (daily.length === 0) {
        const empty = document.createElement('p');
        empty.className = 'chart-card__empty';
        empty.textContent = 'Пока нет данных для графика';
        container.appendChild(empty);
        return container;
    }

    const volumes = daily.map(function(d) { return d.volume; });
    const volumeMax = Math.max(...volumes);
    const volumeMaxSafe = volumeMax > 0 ? volumeMax : 1;

    // Те же отступы viewBox, что и у остальных графиков вкладки — визуально согласованы между собой.
    // Левый отступ чуть больше, чем у карты подходов — под четырёхзначные значения объёма
    const padLeft = 52, padRight = 254, padTop = 20, padBottom = 186;

    const svg = createSVGElement('svg', {
        viewBox: '0 0 300 226',
        style: 'width:100%;height:216px'
    });

    // Горизонтальные линии сетки и подписи оси Y (объём)
    const gridLinesCount = 3;
    for (let i = 0; i <= gridLinesCount; i++) {
        const volumeValue = (i / gridLinesCount) * volumeMaxSafe;
        const y = scale(volumeValue, 0, volumeMaxSafe, padBottom, padTop);

        const gridLine = createSVGElement('line', {
            x1: padLeft, y1: y, x2: padRight, y2: y,
            stroke: i === 0 ? '#3a3128' : '#241f18'
        });
        svg.appendChild(gridLine);

        const label = createSVGElement('text', {
            x: padLeft - 6, y: y + 3,
            fill: '#6f675c', 'font-size': 9,
            'text-anchor': 'end'
        });
        label.textContent = Math.round(volumeValue);
        svg.appendChild(label);
    }

    // Подписи осей целиком («Объём, кг» повёрнуто слева, «Дата тренировки» снизу)
    const yAxisTitle = createSVGElement('text', {
        x: 8, y: (padTop + padBottom) / 2,
        fill: '#a49a8c', 'font-size': 10, 'font-weight': 600,
        'text-anchor': 'middle',
        transform: 'rotate(-90 8 ' + (padTop + padBottom) / 2 + ')'
    });
    yAxisTitle.textContent = 'Объём, кг';
    svg.appendChild(yAxisTitle);

    const xAxisTitle = createSVGElement('text', {
        x: (padLeft + padRight) / 2, y: 219,
        fill: '#8f877b', 'font-size': 9,
        'text-anchor': 'middle'
    });
    xAxisTitle.textContent = 'ДАТА ТРЕНИРОВКИ';
    svg.appendChild(xAxisTitle);


    // Столбцы всегда укладываются в фиксированную ширину области графика —
    // шаг считается от числа тренировок, а не наоборот, иначе при большом
    // их количестве столбцы уезжают за правый край viewBox и обрезаются.
    // Ширина столбца сужается вместе с шагом, но не меньше минимальной,
    // чтобы совсем узкие столбцы оставались видимыми
    const plotWidth = padRight - padLeft;
    const step = daily.length > 0 ? plotWidth / daily.length : plotWidth;
    const barWidth = Math.max(4, Math.min(24, step - 6));

    // Подписи дат показываем не под каждым столбцом — иначе при большом их
    // числе подписи наслаиваются друг на друга
    const barCenterXs = daily.map(function(d, i) {
        return padLeft + i * step + step / 2;
    });
    const labelIndices = pickLabelIndices(barCenterXs);

    daily.forEach(function(d, i) {
        const isLast = i === daily.length - 1;
        const barHeight = scale(d.volume, 0, volumeMaxSafe, 0, padBottom - padTop);
        const x = padLeft + i * step + (step - barWidth) / 2;
        const y = padBottom - barHeight;

        // Старые тренировки бледнее, свежие — насыщеннее, последняя — полностью залита акцентом
        const opacity = daily.length > 1 ? scale(i, 0, daily.length - 1, 0.4, 1) : 1;

        const bar = createSVGElement('rect', {
            x: x, y: y, width: barWidth, height: Math.max(barHeight, 1),
            rx: 4,
            fill: '#2ecbff',
            opacity: isLast ? 1 : opacity
        });
        const tooltip = createSVGElement('title', {});
        tooltip.textContent = formatDate(d.date) + ' — ' + d.volume + ' кг';
        bar.appendChild(tooltip);
        svg.appendChild(bar);

        if (isLast) {
            const valueLabel = createSVGElement('text', {
                x: x + barWidth / 2, y: y - 6,
                fill: '#f5f1ea', 'font-size': 9,
                'font-family': 'Space Mono,monospace',
                'text-anchor': 'middle'
            });
            valueLabel.textContent = d.volume;
            svg.appendChild(valueLabel);
        }

        if (labelIndices.has(i)) {
            const dateLabel = createSVGElement('text', {
                x: x + barWidth / 2, y: 200,
                fill: isLast ? '#f5f1ea' : '#6f675c', 'font-size': 9,
                'font-family': 'Space Mono,monospace',
                'text-anchor': 'middle'
            });
            dateLabel.textContent = formatShortDate(d.date);
            svg.appendChild(dateLabel);
        }
    });

    container.appendChild(svg);

    return container;
}

// Строит карточку «Тренировки»: список дней с числом подходов и суммарным
// объёмом, от новых к старым
function renderVolumeRecordsCard(sets) {
    const fragment = document.createDocumentFragment();

    const daily = computeDailyVolumes(sets);
    if (daily.length === 0) {
        return fragment;
    }

    const label = document.createElement('div');
    label.className = 'chart-section-label';
    label.textContent = 'Тренировки';
    fragment.appendChild(label);

    const list = document.createElement('div');
    list.className = 'chart-records';

    daily.slice().reverse().forEach(function(d) {
        const row = document.createElement('div');
        row.className = 'chart-record';

        const dateEl = document.createElement('span');
        dateEl.className = 'chart-record__date';
        dateEl.textContent = formatDate(d.date);

        const countEl = document.createElement('span');
        countEl.className = 'chart-record__count';
        countEl.textContent = d.setsCount + ' подхода';

        const volumeEl = document.createElement('span');
        volumeEl.className = 'chart-record__weight';
        volumeEl.textContent = d.volume + ' кг';

        row.appendChild(dateEl);
        row.appendChild(countEl);
        row.appendChild(volumeEl);
        list.appendChild(row);
    });

    fragment.appendChild(list);
    return fragment;
}

// ===== Частота тренировок (weekly bar chart) =====

const FREQUENCY_GOAL = 3; // порог «хорошей» недели — 3 и больше подходов за неделю

// Начало календарной недели (понедельник, 00:00) для даты — подходы
// группируются по календарным неделям, а не по скользящим 7-дневным окнам
function getWeekStart(date) {
    const d = new Date(date);
    const day = d.getDay(); // 0 — воскресенье
    const diff = (day === 0 ? -6 : 1) - day;
    d.setHours(0, 0, 0, 0);
    d.setDate(d.getDate() + diff);
    return d;
}

// Строит список недель от самой старой тренировки до текущей недели
// включительно — недели без подходов тоже попадают в список с нулём,
// иначе провалы пропадали бы с графика и не учитывались бы в «лучшей неделе»
// и общей сумме. weekOffset — 0 для текущей недели, отрицательный для прошлых
function computeWeeklyFrequency(sets) {
    if (sets.length === 0) {
        return [];
    }

    const currentWeekStart = getWeekStart(new Date());
    const msPerWeek = 7 * 24 * 60 * 60 * 1000;

    const setsCountByWeeksAgo = new Map();
    let oldestWeeksAgo = 0;

    sets.forEach(function(set) {
        const weekStart = getWeekStart(new Date(set.workout_date));
        const weeksAgo = Math.round((currentWeekStart - weekStart) / msPerWeek);

        setsCountByWeeksAgo.set(weeksAgo, (setsCountByWeeksAgo.get(weeksAgo) || 0) + 1);

        if (weeksAgo > oldestWeeksAgo) {
            oldestWeeksAgo = weeksAgo;
        }
    });

    const weeks = [];
    for (let weeksAgo = oldestWeeksAgo; weeksAgo >= 0; weeksAgo--) {
        weeks.push({weekOffset: -weeksAgo, setsCount: setsCountByWeeksAgo.get(weeksAgo) || 0});
    }

    return weeks;
}

// Число тренировок в неделю в среднем за период, с запятой вместо точки —
// как принято в русской локали
function formatAverage(value) {
    return value.toFixed(1).replace('.', ',');
}

// Строит карточку со столбчатым графиком «подходов в неделю»: столбец —
// одна календарная неделя, недели с целью (3+ подхода) — акцентные,
// остальные — приглушённые, снизу легенда с расшифровкой цвета
function renderFrequencyChart(sets) {
    const container = document.createElement('div');
    container.className = 'chart-card';

    const header = document.createElement('div');
    header.className = 'chart-card__header';

    const titles = document.createElement('div');
    const title = document.createElement('div');
    title.className = 'chart-card__title';
    title.textContent = 'Подходов в неделю';
    const desc = document.createElement('div');
    desc.className = 'chart-card__desc';
    desc.style.margin = '0';
    desc.textContent = 'по этому тренажёру';
    titles.appendChild(title);
    titles.appendChild(desc);
    header.appendChild(titles);

    const weeks = computeWeeklyFrequency(sets);

    if (weeks.length === 0) {
        container.appendChild(header);
        const empty = document.createElement('p');
        empty.className = 'chart-card__empty';
        empty.textContent = 'Пока нет данных для графика';
        container.appendChild(empty);
        return container;
    }

    const totalSets = weeks.reduce(function(sum, w) { return sum + w.setsCount; }, 0);
    const badge = document.createElement('div');
    badge.className = 'chart-card__badge';
    badge.textContent = 'в среднем ' + formatAverage(totalSets / weeks.length);
    header.appendChild(badge);

    container.appendChild(header);

    const counts = weeks.map(function(w) { return w.setsCount; });
    const countMax = Math.max(...counts, FREQUENCY_GOAL + 1);

    // Те же отступы viewBox, что и у остальных графиков вкладки
    const padLeft = 44, padRight = 254, padTop = 20, padBottom = 186;

    const svg = createSVGElement('svg', {
        viewBox: '0 0 300 226',
        style: 'width:100%;height:216px'
    });

    // Горизонтальные линии сетки и подписи оси Y (число подходов)
    const gridLinesCount = 3;
    for (let i = 0; i <= gridLinesCount; i++) {
        const countValue = (i / gridLinesCount) * countMax;
        const y = scale(countValue, 0, countMax, padBottom, padTop);

        const gridLine = createSVGElement('line', {
            x1: padLeft, y1: y, x2: padRight, y2: y,
            stroke: i === 0 ? '#3a3128' : '#241f18'
        });
        svg.appendChild(gridLine);

        const label = createSVGElement('text', {
            x: padLeft - 6, y: y + 3,
            fill: '#6f675c', 'font-size': 9,
            'text-anchor': 'end'
        });
        label.textContent = Math.round(countValue);
        svg.appendChild(label);
    }

    // Подписи осей целиком («Подходов» повёрнуто слева, «Неделя» снизу)
    const yAxisTitle = createSVGElement('text', {
        x: 10, y: (padTop + padBottom) / 2,
        fill: '#a49a8c', 'font-size': 10, 'font-weight': 600,
        'text-anchor': 'middle',
        transform: 'rotate(-90 10 ' + (padTop + padBottom) / 2 + ')'
    });
    yAxisTitle.textContent = 'Подходов';
    svg.appendChild(yAxisTitle);

    const xAxisTitle = createSVGElement('text', {
        x: (padLeft + padRight) / 2, y: 219,
        fill: '#8f877b', 'font-size': 9,
        'text-anchor': 'middle'
    });
    xAxisTitle.textContent = 'НЕДЕЛЯ (0 — ТЕКУЩАЯ)';
    svg.appendChild(xAxisTitle);

    // Столбцы всегда укладываются в фиксированную ширину области графика —
    // тот же приём, что и на графике объёма, чтобы столбцы не уезжали
    // за правый край при большом числе недель в выбранном периоде
    const plotWidth = padRight - padLeft;
    const step = plotWidth / weeks.length;
    const barWidth = Math.max(4, Math.min(24, step - 6));

    const barCenterXs = weeks.map(function(w, i) {
        return padLeft + i * step + step / 2;
    });
    const labelIndices = pickLabelIndices(barCenterXs);

    weeks.forEach(function(w, i) {
        const barHeight = scale(w.setsCount, 0, countMax, 0, padBottom - padTop);
        const x = padLeft + i * step + (step - barWidth) / 2;
        const y = padBottom - barHeight;
        const reachedGoal = w.setsCount >= FREQUENCY_GOAL;

        const bar = createSVGElement('rect', {
            x: x, y: y, width: barWidth, height: Math.max(barHeight, 1),
            rx: 4,
            fill: reachedGoal ? '#2ecbff' : '#3a3128'
        });
        const tooltip = createSVGElement('title', {});
        tooltip.textContent = (w.weekOffset === 0 ? 'Текущая неделя' : w.weekOffset + ' нед.') +
            ' — ' + w.setsCount + ' подх.';
        bar.appendChild(tooltip);
        svg.appendChild(bar);

        if (labelIndices.has(i)) {
            const label = createSVGElement('text', {
                x: x + barWidth / 2, y: 200,
                fill: w.weekOffset === 0 ? '#f5f1ea' : '#6f675c', 'font-size': 9,
                'font-family': 'Space Mono,monospace',
                'text-anchor': 'middle'
            });
            label.textContent = w.weekOffset;
            svg.appendChild(label);
        }
    });

    container.appendChild(svg);

    // Легенда снизу: расшифровка цвета столбцов по достижению цели
    const legend = document.createElement('div');
    legend.className = 'chart-bar-legend';

    const goalItem = document.createElement('span');
    goalItem.className = 'chart-bar-legend__item';
    const goalSwatch = document.createElement('span');
    goalSwatch.className = 'chart-bar-legend__swatch';
    goalSwatch.style.background = '#2ecbff';
    goalItem.appendChild(goalSwatch);
    goalItem.appendChild(document.createTextNode(FREQUENCY_GOAL + ' и больше за неделю'));

    const belowItem = document.createElement('span');
    belowItem.className = 'chart-bar-legend__item';
    const belowSwatch = document.createElement('span');
    belowSwatch.className = 'chart-bar-legend__swatch';
    belowSwatch.style.background = '#3a3128';
    belowItem.appendChild(belowSwatch);
    belowItem.appendChild(document.createTextNode('меньше ' + FREQUENCY_GOAL));

    legend.appendChild(goalItem);
    legend.appendChild(belowItem);
    container.appendChild(legend);

    return container;
}

// Ищет последнюю (считая от текущей недели назад) неделю, не дотянувшую до
// цели, и возвращает сколько недель назад это было. Если таких недель не
// было ни разу за период — возвращает null
function findWeeksSinceBelowGoal(weeks) {
    for (let i = weeks.length - 1; i >= 0; i--) {
        if (weeks[i].setsCount < FREQUENCY_GOAL) {
            return -weeks[i].weekOffset;
        }
    }
    return null;
}

// Строит карточку «Регулярность»: серия недель подряд с достижением цели,
// лучшая неделя, сумма подходов за период и давность последнего провала
function renderFrequencyRecordsCard(sets) {
    const fragment = document.createDocumentFragment();

    const weeks = computeWeeklyFrequency(sets);
    if (weeks.length === 0) {
        return fragment;
    }

    const label = document.createElement('div');
    label.className = 'chart-section-label';
    label.textContent = 'Регулярность';
    fragment.appendChild(label);

    const list = document.createElement('div');
    list.className = 'chart-records';

    let streak = 0;
    for (let i = weeks.length - 1; i >= 0; i--) {
        if (weeks[i].setsCount < FREQUENCY_GOAL) {
            break;
        }
        streak++;
    }

    const bestWeek = Math.max(...weeks.map(function(w) { return w.setsCount; }));
    const totalSets = weeks.reduce(function(sum, w) { return sum + w.setsCount; }, 0);
    const weeksSinceBelowGoal = findWeeksSinceBelowGoal(weeks);

    const rows = [
        {
            label: 'Недель подряд с ' + FREQUENCY_GOAL + '+ подходами',
            value: String(streak),
            modifier: 'chart-record__weight--accent'
        },
        {
            label: 'Лучшая неделя',
            value: bestWeek + ' подходов',
            modifier: ''
        },
        {
            label: 'Всего за ' + weeks.length + ' недель',
            value: totalSets + ' подходов',
            modifier: ''
        },
        {
            label: 'Последняя неделя ниже цели',
            value: weeksSinceBelowGoal === null ? 'ни разу' :
                weeksSinceBelowGoal === 0 ? 'на этой неделе' : weeksSinceBelowGoal + ' нед. назад',
            modifier: 'chart-record__weight--muted'
        }
    ];

    rows.forEach(function(rowData) {
        const row = document.createElement('div');
        row.className = 'chart-record';

        const labelEl = document.createElement('span');
        labelEl.className = 'chart-record__label';
        labelEl.textContent = rowData.label;

        const valueEl = document.createElement('span');
        valueEl.className = 'chart-record__weight' + (rowData.modifier ? ' ' + rowData.modifier : '');
        valueEl.textContent = rowData.value;

        row.appendChild(labelEl);
        row.appendChild(valueEl);
        list.appendChild(row);
    });

    fragment.appendChild(list);
    return fragment;
}

// Подсвечивает на графике точку, которой принадлежит подход с данным id,
// и плавно прокручивает её в область видимости. Снимает подсветку с прошлой точки
function highlightPoint(setId) {
    const circle = setIdToCircle.get(setId);
    if (!circle) {
        return;
    }

    if (highlightedCircle) {
        highlightedCircle.classList.remove('chart-point--highlighted');
    }

    circle.classList.add('chart-point--highlighted');
    highlightedCircle = circle;
    circle.scrollIntoView({behavior: 'smooth', block: 'center'});
}

// ===== Недавние подходы под графиком =====

const RECENT_DAYS_COUNT = 3; // сколько последних дней тренировок показывать под графиком

// Группирует подходы по дате тренировки — бэкенд отдаёт их плоским списком
function groupSetsByDate(sets) {
    const groups = {};

    sets.forEach(function(set) {
        const date = set.workout_date.split('T')[0];
        if (!groups[date]) {
            groups[date] = [];
        }
        groups[date].push(set);
    });

    return groups;
}

// Строит карточку «Недавние подходы» с последними днями тренировок под графиком
function renderRecentSetsCard(sets) {
    const container = document.createElement('div');
    container.className = 'chart-card';

    const title = document.createElement('div');
    title.className = 'chart-card__title';
    title.textContent = 'Недавние подходы';
    container.appendChild(title);

    if (sets.length === 0) {
        const empty = document.createElement('p');
        empty.className = 'chart-card__empty';
        empty.textContent = 'Пока нет данных';
        container.appendChild(empty);
        return container;
    }

    const grouped = groupSetsByDate(sets);
    const recentDates = Object.keys(grouped).sort().reverse().slice(0, RECENT_DAYS_COUNT);

    recentDates.forEach(function(date) {
        container.appendChild(createRecentDayBlock(date, grouped[date]));
    });

    return container;
}

// Один день из карточки «Недавние подходы»: дата, число подходов и список сетов
function createRecentDayBlock(date, sets) {
    const block = document.createElement('div');
    block.className = 'history-day';

    const header = document.createElement('div');
    header.className = 'history-day__header';

    const dateEl = document.createElement('span');
    dateEl.className = 'history-day__date';
    dateEl.textContent = formatDate(date);

    const countEl = document.createElement('span');
    countEl.className = 'history-day__count';
    countEl.textContent = sets.length + ' подхода';

    header.appendChild(dateEl);
    header.appendChild(countEl);
    block.appendChild(header);

    const setsContainer = document.createElement('div');
    setsContainer.className = 'history-day__sets';

    sets.forEach(function(set) {
        const row = document.createElement('div');
        row.className = 'history-set history-set--editable';

        const label = document.createElement('span');
        label.className = 'history-set__label';
        label.textContent = 'Подход ' + set.set_number;

        const value = document.createElement('span');
        value.className = 'history-set__value';
        value.textContent = set.weight_kg + ' кг × ' + set.reps;

        row.appendChild(label);
        row.appendChild(value);
        row.addEventListener('click', function() {
            highlightPoint(set.id);
        });
        setsContainer.appendChild(row);
    });

    block.appendChild(setsContainer);
    return block;
}

// ===== Модалка с деталями точки =====

// Открывает нижнюю плашку со списком всех подходов, совпавших в одну точку
// на графике (дата + номер подхода), отсортированных от новых к старым
function openPointDetail(group) {
    pointDetailTitle.textContent = group.weight_kg + ' кг × ' + group.reps;

    const sortedSets = group.sets.slice().sort(function(a, b) {
        return b.workout_date.localeCompare(a.workout_date);
    });

    pointDetailList.innerHTML = '';
    sortedSets.forEach(function(set) {
        const item = document.createElement('li');

        const dateWrap = document.createElement('span');
        const dot = document.createElement('span');
        dot.className = 'chart-point-modal__dot';
        dot.style.background = getPointColor(new Date(set.workout_date).getTime(), group.dates);
        const dateText = document.createElement('span');
        dateText.textContent = formatDate(set.workout_date.split('T')[0]);
        dateWrap.appendChild(dot);
        dateWrap.appendChild(dateText);

        const setSpan = document.createElement('span');
        setSpan.textContent = 'Подход ' + set.set_number;

        item.appendChild(dateWrap);
        item.appendChild(setSpan);
        pointDetailList.appendChild(item);
    });

    pointDetailModal.hidden = false;
    // hidden -> visible нужен один кадр между снятием [hidden] и добавлением
    // класса, иначе transition не запустится (браузер схлопнёт оба изменения)
    requestAnimationFrame(function() {
        pointDetailModal.classList.add('confirm-modal--visible');
    });
}

// Закрывает плашку с деталями точки
function closePointDetail() {
    pointDetailModal.classList.remove('confirm-modal--visible');
    setTimeout(function() {
        pointDetailModal.hidden = true;
    }, 220);
}

pointDetailCloseBtn.addEventListener('click', closePointDetail);
pointDetailBackdrop.addEventListener('click', closePointDetail);