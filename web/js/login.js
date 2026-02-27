// web/js/login.js

// ========== ГЛОБАЛЬНЫЕ ПЕРЕМЕННЫЕ ==========
let currentEmail = '';

// ========== ИНИЦИАЛИЗАЦИЯ ПРИ ЗАГРУЗКЕ ==========
document.addEventListener('DOMContentLoaded', () => {
    const loginForm = document.getElementById('login-form');
    const registerForm = document.getElementById('register-form');

    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }

    if (registerForm) {
        registerForm.addEventListener('submit', handleRegister);
    }

    // Инициализируем обработчики модального окна
    setupCodeInputs();

    document.getElementById('verify-code')?.addEventListener('click', verifyCode);
    document.getElementById('cancel-verification')?.addEventListener('click', hideVerificationModal);
    document.getElementById('resend-code')?.addEventListener('click', (e) => {
        e.preventDefault();
        resendCode();
    });

    // Закрытие по клику вне модалки
    window.addEventListener('click', (e) => {
        const modal = document.getElementById('verification-modal');
        if (e.target === modal) {
            hideVerificationModal();
        }
    });

    const token = localStorage.getItem('token');
    const currentPath = window.location.pathname;

    // Если есть токен и мы на логине - редирект на задачи
    if (token && currentPath === '/login') {
        const hasCookie = document.cookie.includes('token=');

        if (!hasCookie) {
            document.cookie = `token=${token}; path=/; max-age=86400; samesite=strict`;
        }

        window.location.href = '/tasks';
        return;
    }
});

// ========== ПЕРЕКЛЮЧЕНИЕ ФОРМ ==========
function switchForm(formName, event) {
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    event.target.classList.add('active');

    if (formName === 'login') {
        document.getElementById('login-form').classList.add('active');
        document.getElementById('register-form').classList.remove('active');
    } else {
        document.getElementById('login-form').classList.remove('active');
        document.getElementById('register-form').classList.add('active');
    }
}

// ========== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ==========
function isValidEmail(email) {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

function showMessage(formId, message, isError = true) {
    const form = document.getElementById(formId);
    let msgDiv = form.querySelector('.error-message, .success-message');

    if (!msgDiv) {
        msgDiv = document.createElement('div');
        msgDiv.className = isError ? 'error-message' : 'success-message';
        form.insertBefore(msgDiv, form.firstChild);
    } else {
        msgDiv.className = isError ? 'error-message' : 'success-message';
    }

    msgDiv.textContent = message;
    msgDiv.classList.add('show');

    setTimeout(() => {
        msgDiv.classList.remove('show');
    }, 5000);
}

// ========== МОДАЛЬНОЕ ОКНО ВЕРИФИКАЦИИ ==========
function showVerificationModal(email) {
    currentEmail = email;
    const modal = document.getElementById('verification-modal');
    const emailDisplay = document.getElementById('verification-email');
    emailDisplay.textContent = email;
    modal.style.display = 'flex';

    document.querySelector('.code-digit')?.focus();

    document.querySelectorAll('.code-digit').forEach(input => input.value = '');
    document.getElementById('verify-code').disabled = true;
    document.getElementById('verification-message').textContent = '';
}

function hideVerificationModal() {
    document.getElementById('verification-modal').style.display = 'none';
}

function setupCodeInputs() {
    const inputs = document.querySelectorAll('.code-digit');

    inputs.forEach((input, index) => {
        input.addEventListener('input', (e) => {
            if (e.target.value.length > 1) {
                e.target.value = e.target.value.slice(0, 1);
            }

            if (e.target.value && index < inputs.length - 1) {
                inputs[index + 1].focus();
            }

            const allFilled = Array.from(inputs).every(inp => inp.value.length === 1);
            document.getElementById('verify-code').disabled = !allFilled;
        });

        input.addEventListener('keydown', (e) => {
            if (e.key === 'Backspace' && !e.target.value && index > 0) {
                inputs[index - 1].focus();
            }
        });

        input.addEventListener('keypress', (e) => {
            if (!/[0-9]/.test(e.key)) {
                e.preventDefault();
            }
        });
    });
}

// ========== API ЗАПРОСЫ ==========
async function verifyCode() {
    const inputs = document.querySelectorAll('.code-digit');
    const code = Array.from(inputs).map(inp => inp.value).join('');

    console.log('🔍 Отправка кода:', code); // 👈 1. Проверь код
    console.log('🔍 Email:', currentEmail); // 👈 2. Проверь email

    const verifyBtn = document.getElementById('verify-code');
    verifyBtn.disabled = true;
    verifyBtn.textContent = 'Проверка...';

    try {
        console.log('📤 Отправка запроса...');

        const response = await fetch('/api/v1/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                email: currentEmail,
                code: code
            })
        });

        console.log('📥 Статус ответа:', response.status); // 👈 3. Статус

        const data = await response.json();
        console.log('📦 Данные ответа:', data); // 👈 4. Что вернул сервер

        if (response.ok) {
            console.log('✅ Успех! Сохраняем токен...');

            localStorage.setItem('token', data.token);
            localStorage.setItem('user', JSON.stringify(data.user));
            document.cookie = `token=${data.token}; path=/; max-age=86400; samesite=strict`;

            document.getElementById('verification-message').className = 'verification-message success';
            document.getElementById('verification-message').textContent = '✓ Email подтверждён!';

            setTimeout(() => {
                console.log('➡️ Редирект на /tasks');
                window.location.href = '/tasks';
            }, 1000);

        } else {
            console.log('❌ Ошибка от сервера:', data.error);

            document.getElementById('verification-message').className = 'verification-message error';
            document.getElementById('verification-message').textContent = data.error || 'Ошибка подтверждения';
            verifyBtn.disabled = false;
            verifyBtn.textContent = 'Подтвердить';
        }
    } catch (error) {
        console.log('🔥 Ошибка соединения:', error); // 👈 5. Сетевая ошибка

        document.getElementById('verification-message').className = 'verification-message error';
        document.getElementById('verification-message').textContent = 'Ошибка соединения';
        verifyBtn.disabled = false;
        verifyBtn.textContent = 'Подтвердить';
    }
}

async function resendCode() {
    const resendLink = document.getElementById('resend-code');
    resendLink.textContent = 'Отправка...';
    resendLink.style.pointerEvents = 'none';

    try {
        const response = await fetch('/api/v1/resend-code', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email: currentEmail })
        });

        const data = await response.json();

        if (response.ok) {
            document.getElementById('verification-message').className = 'verification-message success';
            document.getElementById('verification-message').textContent = 'Код отправлен повторно';

            document.querySelectorAll('.code-digit').forEach(input => input.value = '');
            document.querySelector('.code-digit')?.focus();
        } else {
            document.getElementById('verification-message').className = 'verification-message error';
            document.getElementById('verification-message').textContent = data.error || 'Ошибка отправки';
        }
    } catch (error) {
        document.getElementById('verification-message').className = 'verification-message error';
        document.getElementById('verification-message').textContent = 'Ошибка соединения';
    } finally {
        resendLink.textContent = 'Отправить код повторно';
        resendLink.style.pointerEvents = 'auto';
    }
}

// ========== ОБРАБОТЧИК ЛОГИНА ==========
async function handleLogin(e) {
    e.preventDefault();

    const form = e.target;
    const submitBtn = form.querySelector('button[type="submit"]');
    const formData = new FormData(form);

    const email = formData.get('email');
    if (!isValidEmail(email)) {
        showMessage('login-form', 'Введите корректный email');
        return;
    }

    submitBtn.disabled = true;
    submitBtn.textContent = 'Вход...';

    form.querySelectorAll('.error-message, .success-message').forEach(msg => msg.remove());

    try {
        const response = await fetch('/api/v1/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                email: email,
                password: formData.get('password')
            })
        });

        const data = await response.json();

        if (response.ok) {
            localStorage.setItem('token', data.token);
            localStorage.setItem('user', JSON.stringify(data.user));
            document.cookie = `token=${data.token}; path=/; max-age=86400; samesite=strict`;

            window.location.href = '/tasks';
        } else {
            // 👇 ОСОБАЯ ОБРАБОТКА ДЛЯ НЕВЕРИФИЦИРОВАННЫХ
            if (response.status === 403 && data.email) {
                showMessage('login-form', 'Email не подтверждён. Пожалуйста, проверьте почту.', true);
                setTimeout(() => {
                    showVerificationModal(data.email);
                }, 1000);
            } else {
                showMessage('login-form', data.error || 'Ошибка сервера');
            }
        }
    } catch (error) {
        showMessage('login-form', 'Ошибка соединения с сервером');
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = 'Войти';
    }
}

// ========== ОБРАБОТЧИК РЕГИСТРАЦИИ ==========
async function handleRegister(e) {
    e.preventDefault();

    const form = e.target;
    const submitBtn = form.querySelector('button[type="submit"]');
    const formData = new FormData(form);

    const firstName = formData.get('first-name');
    const lastName = formData.get('last-name');
    const email = formData.get('email');
    const password = formData.get('password');
    const confirmPassword = formData.get('confirm-password');

    if (!isValidEmail(email)) {
        showMessage('register-form', 'Введите корректный email');
        return;
    }

    if (!password || password.length < 6) {
        showMessage('register-form', 'Пароль должен быть минимум 6 символов');
        return;
    }

    if (password !== confirmPassword) {
        showMessage('register-form', 'Пароли не совпадают!');
        return;
    }

    submitBtn.disabled = true;
    submitBtn.textContent = 'Регистрация...';

    form.querySelectorAll('.error-message, .success-message').forEach(msg => msg.remove());

    const data = {
        firstName: firstName,
        lastName: lastName,
        email: email,
        password: password
    };

    try {
        const response = await fetch('/api/v1/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });

        const result = await response.json();

        if (response.ok) {
            showMessage('register-form', 'Код подтверждения отправлен на почту', false);
            setTimeout(() => {
                showVerificationModal(email);
            }, 1000);
        } else {
            showMessage('register-form', result.error || 'Ошибка сервера');
        }
    } catch (error) {
        showMessage('register-form', 'Ошибка соединения с сервером');
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = 'Зарегистрироваться';
    }
}
