// web/js/tasks.js

// Хранилище задач (будет заполняться с сервера)
let tasks = [];

// Текущие фильтры
let currentFilter = 'all';
let searchQuery = '';
let sortOrder = 'desc';

// ========== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ==========

// Форматирование даты из ISO строки
function formatDate(isoString) {
    if (!isoString) return '';
    const date = new Date(isoString);
    return date.toLocaleDateString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    }).replace(',', '');
}

// Создаем контейнер для уведомлений при загрузке страницы
function initNotifications() {
    // Создаем контейнер, если его нет
    if (!document.getElementById('notification-container')) {
        const container = document.createElement('div');
        container.id = 'notification-container';
        container.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 999;
        `;
        document.body.appendChild(container);
    }
}

// Показать глобальное уведомление
function showNotification(message, type = 'info') {
    initNotifications();

    const container = document.getElementById('notification-container');

    // Создаем уведомление
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    notification.style.cssText = `
        background: ${type === 'error' ? '#f44336' : type === 'success' ? '#4CAF50' : '#2196F3'};
        color: white;
        padding: 15px 20px;
        margin-bottom: 10px;
        border-radius: 5px;
        box-shadow: 0 2px 5px rgba(0,0,0,0.2);
        animation: slideIn 0.3s ease;
        cursor: pointer;
    `;

    // Добавляем анимацию
    const style = document.createElement('style');
    style.textContent = `
        @keyframes slideIn {
            from { transform: translateX(100%); opacity: 0; }
            to { transform: translateX(0); opacity: 1; }
        }
        @keyframes fadeOut {
            from { opacity: 1; }
            to { opacity: 0; }
        }
    `;
    document.head.appendChild(style);

    // Добавляем в контейнер
    container.appendChild(notification);

    // Удаляем через 5 секунд
    setTimeout(() => {
        notification.style.animation = 'fadeOut 0.3s ease';
        setTimeout(() => {
            if (notification.parentNode) {
                notification.remove();
            }
        }, 300);
    }, 5000);

    // Можно закрыть кликом
    notification.addEventListener('click', () => {
        notification.remove();
    });
}

// ========== РАБОТА С API ==========

// TODO: 1. Загрузить задачи с сервера
async function fetchTasks() {
    try {
        const token = localStorage.getItem('token');
        const response = await fetch('/api/v1/tasks', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token }
        });

        if (!response.ok) {
            if (response.status === 401) {
                window.location.href = '/login';
                return;
            }
            throw new Error('Failed to fetch tasks');
        }

        const data = await response.json();
        console.log('📦 Данные с сервера:', data);  // 👈 ПОСМОТРИ СЮДА!
        console.log('📦 Первая задача:', data.tasks[0]);  // 👈 И СЮДА!

        tasks = data.tasks;
        renderTasks();
    } catch (error) {
        console.log('❌ Ошибка:', error);
        showNotification('Ошибка загрузки задач');
    }
}

// TODO: 2. Создать задачу
async function createTask(taskData) {
    try {
        const token = localStorage.getItem('token');
        const response = await fetch('/api/v1/tasks', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token
            },
            body: JSON.stringify(taskData)
        });

        if (!response.ok) {
            throw new Error('Failed to create task');
        }

        const newTask = await response.json();
        tasks.unshift(newTask);
        renderTasks();
        showNotification('Задача создана', false);
    } catch (error) {
        showNotification('Ошибка создания задачи');
    }
}

// TODO: 3. Обновить задачу (текст)
async function updateTask(taskId, updates) {
    // updates: { title?, description? }
    try {
        const token = localStorage.getItem('token');
        const response = await fetch(`/api/v1/tasks/${taskId}`, {
            method: 'PATCH',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': token
            },
            body: JSON.stringify(updates)
        });

        if (!response.ok) {
            throw new Error('Failed to update task');
        }

        const updatedTask = await response.json();
        const index = tasks.findIndex(t => t.id === taskId);
        if (index !== -1) {
            tasks[index] = updatedTask;
            renderTasks();
        }
        showNotification('Задача обновлена', false);
    } catch (error) {
        showNotification('Ошибка обновления задачи');
    }
}

// TODO: 4. Переключить статус задачи
async function toggleTask(taskId) {
    // Пример:
    try {
        const token = localStorage.getItem('token');
        const response = await fetch(`/api/v1/tasks/${taskId}/toggle`, {
            method: 'PUT',
            headers: {
                'Authorization': token
            }
        });

        if (!response.ok) {
            throw new Error('Failed to toggle task');
        }

        const result = await response.json();
        const index = tasks.findIndex(t => t.id === taskId);
        if (index !== -1) {
            tasks[index].completed = result.completed;
            renderTasks();
        }
        showNotification(`Задача ${result.completed ? 'выполнена' : 'возобновлена'}`, false);
    } catch (error) {
        showNotification('Ошибка изменения статуса');
    }
}

// TODO: 5. Удалить задачу
async function deleteTask(taskId) {
    try {
        const token = localStorage.getItem('token');
        const response = await fetch(`/api/v1/tasks/${taskId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': token
            }
        });

        if (!response.ok) {
            throw new Error('Failed to delete task');
        }

        tasks = tasks.filter(t => t.id !== taskId);
        renderTasks();
        showNotification('Задача удалена', false);
    } catch (error) {
        showNotification('Ошибка удаления задачи');
    }
}

// ========== РЕНДЕРИНГ ЗАДАЧ ==========

function renderTasks() {
    const tasksGrid = document.getElementById('tasksGrid');
    const template = document.getElementById('taskTemplate');

    // Фильтрация
    let filteredTasks = tasks.filter(task => {
        if (currentFilter === 'active' && task.completed) return false;
        if (currentFilter === 'completed' && !task.completed) return false;
        if (searchQuery && !task.title.toLowerCase().includes(searchQuery.toLowerCase())) return false;
        return true;
    });

    // Сортировка
    filteredTasks.sort((a, b) => {
        const dateA = new Date(a.createdAt);
        const dateB = new Date(b.createdAt);
        return sortOrder === 'desc' ? dateB - dateA : dateA - dateB;
    });

    tasksGrid.innerHTML = '';

    if (filteredTasks.length === 0) {
        tasksGrid.innerHTML = '<div class="no-tasks">Задачи не найдены</div>';
        return;
    }

    filteredTasks.forEach(task => {
        const taskElement = document.importNode(template.content, true);
        const card = taskElement.querySelector('.task-card');

        const taskData = {
            id: task.id,
            title: task.title,
            description: task.description,
            completed: task.completed,
            createdAt: task.createdAt,           // с большой A!
            completedAt: task.completedAt        // если есть
        };

        card.dataset.taskId = taskData.id;
        if (taskData.completed) {
            card.classList.add('completed');
        }

        // Заголовок и описание
        card.querySelector('.title-text').textContent = taskData.title;
        card.querySelector('.description-text').textContent = taskData.description || 'Нет описания';

        // Поля для редактирования
        card.querySelector('.edit-title-input').value = taskData.title;
        card.querySelector('.edit-description-input').value = taskData.description || '';

        // Чекбокс
        const checkbox = card.querySelector('.task-checkbox-input');
        checkbox.id = `task-${taskData.id}`;
        checkbox.checked = taskData.completed;
        checkbox.addEventListener('change', () => toggleTask(task.id));

        card.querySelector('.checkbox-custom').setAttribute('for', `task-${taskData.id}`);

        // Даты
        card.querySelector('.created-value').textContent = formatDate(taskData.createdAt);

        const completedSpan = card.querySelector('.completed-date');
        if (taskData.completed && taskData.completedAt) {
            completedSpan.style.display = 'inline';
            completedSpan.querySelector('.completed-value').textContent = formatDate(taskData.completedAt);
        }

        // Статус
        card.querySelector('.task-status').textContent = task.completed ? 'Выполнено' : 'В работе';
        // Кнопки действий
        const editBtn = card.querySelector('.edit-btn');
        const saveBtn = card.querySelector('.save-btn');
        const cancelBtn = card.querySelector('.cancel-btn');
        const deleteBtn = card.querySelector('.delete-btn');

        editBtn.addEventListener('click', () => startEditing(card, task.id));
        saveBtn.addEventListener('click', () => saveEdit(card, task.id));
        cancelBtn.addEventListener('click', () => cancelEdit(card, task.id));
        deleteBtn.addEventListener('click', () => deleteTask(task.id));

        tasksGrid.appendChild(taskElement);
    });
}

// ========== РЕДАКТИРОВАНИЕ ==========

function startEditing(card, taskId) {
    document.querySelectorAll('.task-card.editing').forEach(c => {
        cancelEdit(c);
    });
    card.classList.add('editing');
}

function saveEdit(card, taskId) {
    const newTitle = card.querySelector('.edit-title-input').value.trim();
    const newDescription = card.querySelector('.edit-description-input').value.trim();

    if (newTitle) {
        updateTask(taskId, { title: newTitle, description: newDescription });
        card.classList.remove('editing');
    }
}

function cancelEdit(card) {
    card.classList.remove('editing');
    const taskId = parseInt(card.dataset.taskId);
    const task = tasks.find(t => t.id === taskId);
    if (task) {
        card.querySelector('.edit-title-input').value = task.title;
        card.querySelector('.edit-description-input').value = task.description || '';
    }
}

// ========== ИНИЦИАЛИЗАЦИЯ ==========

document.addEventListener('DOMContentLoaded', () => {
    initNotifications()

    const token = localStorage.getItem('token');
    if (!token) {
        showNotification("Not authorized", "error")
        window.location.href = '/login';
    }
    fetchTasks();

    // Форма создания задачи
    const showFormBtn = document.getElementById('showCreateFormBtn');
    const formContainer = document.getElementById('taskFormContainer');
    const cancelBtn = document.getElementById('cancelFormBtn');
    const taskForm = document.getElementById('taskForm');

    showFormBtn.addEventListener('click', () => {
        formContainer.style.display = 'block';
        showFormBtn.style.display = 'none';
    });

    cancelBtn.addEventListener('click', () => {
        formContainer.style.display = 'none';
        showFormBtn.style.display = 'flex';
        taskForm.reset();
    });

    // Создание новой задачи
    taskForm.addEventListener('submit', (e) => {
        e.preventDefault();

        const newTask = {
            title: document.getElementById('taskName').value,
            description: document.getElementById('taskDescription').value
        };

        createTask(newTask);

        taskForm.reset();
        formContainer.style.display = 'none';
        showFormBtn.style.display = 'flex';
    });

    // Поиск
    document.getElementById('searchInput').addEventListener('input', (e) => {
        searchQuery = e.target.value;
        renderTasks();
    });

    // Фильтры
    document.querySelectorAll('.filter-tab').forEach(btn => {
        btn.addEventListener('click', (e) => {
            document.querySelectorAll('.filter-tab').forEach(b => b.classList.remove('active'));
            e.target.classList.add('active');
            currentFilter = e.target.dataset.filter;
            renderTasks();
        });
    });

    // Сортировка
    document.getElementById('dateSort').addEventListener('change', (e) => {
        sortOrder = e.target.value;
        renderTasks();
    });

    // Выход (уже правильно настроен)
    document.getElementById('logoutBtn').addEventListener('click', () => {
        const modal = document.createElement('div');
        modal.className = 'confirm-modal';
        modal.innerHTML = `
            <div class="confirm-modal-content">
                <h3>Выход из системы</h3>
                <p>Вы действительно хотите выйти?</p>
                <div class="confirm-modal-actions">
                    <button class="btn-secondary" id="cancelLogout">Отмена</button>
                    <button class="btn-primary" id="confirmLogout">Выйти</button>
                </div>
            </div>
        `;

        document.body.appendChild(modal);

        document.getElementById('cancelLogout').addEventListener('click', () => {
            modal.remove();
        });

        document.getElementById('confirmLogout').addEventListener('click', () => {
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            document.cookie = 'token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
            window.location.href = '/login';
        });

        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.remove();
            }
        });
    });
});
