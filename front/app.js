const API_BASE = window.location.origin + '/api/v1';
const API = {
    BOOK_LIST: `${API_BASE}/book/list`,
    BOOK_READ: `${API_BASE}/book/read`,
    BOOK_UPLOAD: `${API_BASE}/book/upload`,
    BOOK_DELETE: `${API_BASE}/book/delete`,
    BOOK_PROGRESS_GET: `${API_BASE}/book/progress/get`,
    BOOK_PROGRESS_SAVE: `${API_BASE}/book/progress/set`,
    REGISTER: `${API_BASE}/registration`,
    LOGIN: `${API_BASE}/login`,
    LOGOUT: `${API_BASE}/logout`,
    PROFILE: `${API_BASE}/profile`,
    USER_LIST: `${API_BASE}/admin/user/list`,
    USER_DELETE: `${API_BASE}/admin/user/delete`,
    ROLE_LIST: `${API_BASE}/admin/role/list`,
    ADD_ROLE: `${API_BASE}/admin/role/set`
};

const elements = {
    loginBtn: document.getElementById('loginBtn'),
    registerBtn: document.getElementById('registerBtn'),
    profileBtn: document.getElementById('profileBtn'),
    logoutBtn: document.getElementById('logoutBtn'),
    closeReaderBtn: document.getElementById('closeReaderBtn'),
    prevPageBtn: document.getElementById('prevPageBtn'),
    nextPageBtn: document.getElementById('nextPageBtn'),
    bookList: document.getElementById('bookList'),
    bookListSection: document.getElementById('bookListSection'),
    readerSection: document.getElementById('readerSection'),
    readerContent: document.getElementById('readerContent'),
    bookTitle: document.getElementById('bookTitle'),
    pageInfo: document.getElementById('pageInfo'),
    uploadSection: document.getElementById('uploadSection'),
    adminSection: document.getElementById('adminSection'),
    loginModal: document.getElementById('loginModal'),
    registerModal: document.getElementById('registerModal'),
    profileModal: document.getElementById('profileModal'),
    closeLoginModal: document.getElementById('closeLoginModal'),
    closeRegisterModal: document.getElementById('closeRegisterModal'),
    closeProfileModal: document.getElementById('closeProfileModal'),
    loginForm: document.getElementById('loginForm'),
    loginUsername: document.getElementById('loginUsername'),
    loginPassword: document.getElementById('loginPassword'),
    registerForm: document.getElementById('registerForm'),
    registerLogin: document.getElementById('registerLogin'),
    registerEmail: document.getElementById('registerEmail'),
    registerPassword: document.getElementById('registerPassword'),
    profileInfo: document.getElementById('profileInfo'),
    uploadForm: document.getElementById('uploadForm'),
    bookFile: document.getElementById('bookFile'),
    bookUrl: document.getElementById('bookUrl'),
    userList: document.querySelector('#userList tbody'),
    roleList: document.querySelector('#roleList tbody'),
    fontFamily: document.getElementById('fontFamily'),
    fontSize: document.getElementById('fontSize'),
    themeSelector: document.getElementById('themeSelector'),
};

let state = {
    currentUser: null,
    books: [],
    currentBook: null,
    currentPage: 1,
    totalPages: 1,
    users: [],
    roles: [],
    readerSettings: {
        fontFamily: "'Roboto', sans-serif",
        fontSize: "16px",
        theme: "light"
    }
};

function init() {
    loadUserFromStorage();
    loadReaderTheme();
    setupEventListeners();
    fetchBooks();
    applyReaderSettings();
    applyReaderTheme();

    if (state.currentUser) {
        updateUIForUser();
        if (state.currentUser.role?.role_name === 'admin') {
            fetchUsers();
            fetchRoles();
        }
    }
}

function loadUserFromStorage() {
    const userData = localStorage.getItem('currentUser');
    if (userData) {
        state.currentUser = JSON.parse(userData);
    }
}

function saveUserToStorage() {
    if (state.currentUser) {
        localStorage.setItem('currentUser', JSON.stringify(state.currentUser));
    } else {
        localStorage.removeItem('currentUser');
    }
}

function setupEventListeners() {
    elements.loginBtn.addEventListener('click', () => elements.loginModal.style.display = 'flex');
    elements.registerBtn.addEventListener('click', () => elements.registerModal.style.display = 'flex');
    elements.profileBtn.addEventListener('click', () => {
        elements.profileInfo.innerHTML = `
                    <p><strong>Логин:</strong> ${state.currentUser.login}</p>
                    <p><strong>Email:</strong> ${state.currentUser.email}</p>
                    <p><strong>Роль:</strong> ${state.currentUser.role?.role_name}</p>
                `;
        elements.profileModal.style.display = 'flex';
    });
    elements.logoutBtn.addEventListener('click', logout);

    elements.closeLoginModal.addEventListener('click', () => elements.loginModal.style.display = 'none');
    elements.closeRegisterModal.addEventListener('click', () => elements.registerModal.style.display = 'none');
    elements.closeProfileModal.addEventListener('click', () => elements.profileModal.style.display = 'none');

    elements.loginForm.addEventListener('submit', handleLogin);
    elements.registerForm.addEventListener('submit', handleRegister);
    elements.uploadForm.addEventListener('submit', handleUpload);

    elements.closeReaderBtn.addEventListener('click', () => {
        toggleReaderMode(false)
    });

    elements.prevPageBtn.addEventListener('click', async () => {
        if (state.currentPage > 1) {
            state.currentPage--;
            await saveBookProgress(state.currentBook.id, state.currentPage);
            loadBookPage();
        }
    });

    elements.nextPageBtn.addEventListener('click', async () => {
        if (state.currentPage < state.totalPages) {
            state.currentPage++;
            await saveBookProgress(state.currentBook.id, state.currentPage);
            loadBookPage();
        }
    });

    elements.fontFamily.addEventListener('change', updateReaderSettings);
    elements.fontSize.addEventListener('change', updateReaderSettings);
    if (elements.themeSelector) {
        elements.themeSelector.addEventListener('change', updateReaderSettings);
    }

    window.addEventListener('click', (e) => {
        if (e.target === elements.loginModal) elements.loginModal.style.display = 'none';
        if (e.target === elements.registerModal) elements.registerModal.style.display = 'none';
        if (e.target === elements.profileModal) elements.profileModal.style.display = 'none';
    });
}

function updateUIForUser() {
    if (state.currentUser) {
        elements.loginBtn.classList.add('hidden');
        elements.registerBtn.classList.add('hidden');
        elements.profileBtn.classList.remove('hidden');
        elements.logoutBtn.classList.remove('hidden');

        if (state.currentUser.role?.role_name === 'admin' || state.currentUser.role?.role_name === 'super') {
            elements.uploadSection.classList.remove('hidden');
        } else {
            elements.uploadSection.classList.add('hidden');
        }

        if (state.currentUser.role?.role_name === 'admin') {
            elements.adminSection.classList.remove('hidden');
        } else {
            elements.adminSection.classList.add('hidden');
        }
    } else {
        elements.loginBtn.classList.remove('hidden');
        elements.registerBtn.classList.remove('hidden');
        elements.profileBtn.classList.add('hidden');
        elements.logoutBtn.classList.add('hidden');
        elements.uploadSection.classList.add('hidden');
        elements.adminSection.classList.add('hidden');
    }

    renderBooks();
}

async function fetchBooks() {
    try {
        const response = await fetch(API.BOOK_LIST);
        if (!response.ok) throw new Error('Ошибка загрузки книг');

        const books = await response.json();
        state.books = books;
        renderBooks();
    } catch (error) {
        console.error('Error fetching books:', error);
        alert('Не удалось загрузить книги');
    }
}

function renderBooks() {
    console.log('Rendering books:', state.books);
    elements.bookList.innerHTML = '';

    if (!state.books.length) {
        console.error('No books to render!');
        return;
    }

    state.books.forEach(book => {
        const bookCard = document.createElement('div');
        bookCard.className = 'book-card';
        let createdAt = 'Дата не указана'
        if (book.created_at){
            const date = new Date(book.created_at * 1000)
            createdAt = date.toLocaleDateString('ru-RU',
                {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric'
                });
        }

        bookCard.innerHTML = `
                    ${canDeleteBook(book) ?
                    `<button class="delete-book-btn" data-id="${book.id}" title="Удалить книгу"></button>` : ''}
                    <h3 class="book-title">${book.title}</h3>
                    <p class="book-author">${book.author}</p>
                    <p class="book-annotation">${book.annotation}</p>
                    <div class="book-meta">
                        <span>${book.pages} стр.</span>
                        <span>${new Date(book.created_at * 1000).toLocaleDateString()}</span>
                    </div>
                    <button class="read-btn" data-id="${book.id}">Читать</button>
                `;

        elements.bookList.appendChild(bookCard);
    });

    document.querySelectorAll('.read-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const bookId = e.target.getAttribute('data-id');
            openBook(bookId);
        });
    });

    document.querySelectorAll('.delete-book-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const bookId = e.target.getAttribute('data-id');
            deleteItem('book', bookId);
        });
    });
}

async function loadBookProgress(bookId) {
    try {
        const token = localStorage.getItem('token');
        if (!token) return 1;

        const response = await fetch(`${API.BOOK_PROGRESS_GET}?id=${bookId}`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (response.ok) {
            const data = await response.json();
            return data.current_page || 1;
        }
    } catch (error) {
        console.error('Error loading progress:', error);
    }
    return 1;
}

async function saveBookProgress(bookId, page) {
    try {
        const token = localStorage.getItem('token');
        if (!token) return;

        await fetch(API.BOOK_PROGRESS_SAVE, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                book_id: bookId,
                page: page
            })
        });
    } catch (error) {
        console.error('Error saving progress:', error);
    }
}

async function openBook(bookId) {
    if (!state.currentUser) {
        alert('Для чтения книг необходимо авторизоваться');
        elements.loginModal.style.display = 'flex';
        return;
    }

    const token = localStorage.getItem('token');
    if (!token) {
        logout();
        alert('Требуется авторизация');
        return;
    }

    const book = state.books.find(b => b.id == bookId);
    if (!book) return;

    state.currentBook = book;
    state.currentPage = 1;

    state.currentPage = await loadBookProgress(bookId);

    if (state.currentUser) {
        if (state.currentUser.role_name === 'user') {
            state.totalPages = Math.min(15, book.pages);
        } else {
            state.totalPages = book.pages;
        }
    } else {
        return;
    }

    elements.bookTitle.textContent = `${book.title} - ${book.author}`;
    toggleReaderMode(true)

    loadBookPage();
}

function toggleReaderMode(isReading) {
    if (isReading) {
        document.querySelectorAll('#bookListSection, #uploadSection, #adminSection').forEach(el => {
            el.classList.add('hidden');
        });
        elements.readerSection.classList.remove('hidden');
    } else {
        elements.bookListSection.classList.remove('hidden');
        elements.readerSection.classList.add('hidden');

        if (state.currentUser) {
            if (state.currentUser.role?.role_name === 'admin' || state.currentUser.role?.role_name === 'super') {
                elements.uploadSection.classList.remove('hidden');
            }
            if (state.currentUser.role?.role_name === 'admin') {
                elements.adminSection.classList.remove('hidden');
            }
        }
    }
}

async function loadBookPage() {
    if (!state.currentBook) return;

    const token = localStorage.getItem('token');
    if (!token) {
        alert('Требуется авторизация');
        logout();
        return;
    }

    try {
        const response = await fetch(`${API.BOOK_READ}?id=${state.currentBook.id}&page=${state.currentPage}`,
            {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
        if (!response.ok) {
            if (response.status === 401) {
                logout();
                alert('Сессия истекла, войдите снова');
                return;
            }
            throw new Error('Ошибка загрузки страницы');
        }

        const pageContent = await response.text();

        const formattedContent = formatBookText(pageContent);
        elements.readerContent.innerHTML = formattedContent;

        elements.readerContent.innerHTML = `<p>${formattedContent}</p>`;
        elements.pageInfo.textContent = `Страница ${state.currentPage} из ${state.totalPages}`;

        elements.prevPageBtn.disabled = state.currentPage <= 1;
        elements.nextPageBtn.disabled = state.currentPage >= state.totalPages;
    } catch (error) {
        console.error('Error loading book page:', error);
        elements.readerContent.innerHTML = '<p>Не удалось загрузить страницу</p>';
    }
}

function formatBookText(text) {
    text = text.replace(/\\n/g, '\n');
    text = text.replace(/"/g, '');
    text = text.replace(/\*\s\*\s\*/g, '<div class="separator">* * *</div>');

    const paragraphs = text.split(/(?:\n|\r|\r\n){2,}/);

    return paragraphs.map(p => {
        const trimmed = p.trim();
        if (!trimmed) return '';

        return `<p>${trimmed.replace(/(?:\n|\r|\r\n)/g, '<br>')}</p>`;
    }).join('');
}

async function handleLogin(e) {
    e.preventDefault();

    const login = elements.loginUsername.value;
    const password = elements.loginPassword.value;

    try {
        const response = await fetch(API.LOGIN, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                login,
                password
            })
        });

        if (!response.ok) throw new Error('Ошибка входа');

        const data = await response.json();

        localStorage.setItem('token', data.token);
        localStorage.setItem('refreshToken', data.refresh_token);

        await fetchProfile();

        elements.loginModal.style.display = 'none';
        elements.loginForm.reset();
    } catch (error) {
        console.error('Login error:', error);
        alert('Неверный логин или пароль');
    }
}

async function handleRegister(e) {
    e.preventDefault();

    const login = elements.registerLogin.value;
    const email = elements.registerEmail.value;
    const password = elements.registerPassword.value;

    try {
        const registerResponse = await fetch(API.REGISTER, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                login,
                email,
                password
            })
        });

        if (!registerResponse.ok) {
            const errorData = await registerResponse.json();
            throw new Error(errorData.message || 'Ошибка регистрации');
        }

        const loginResponse = await fetch(API.LOGIN, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                login,
                password
            })
        });

        if (!loginResponse.ok) {
            throw new Error('Ошибка автоматического входа после регистрации');
        }

        const { token, refresh_token } = await loginResponse.json();
        localStorage.setItem('token', token);
        localStorage.setItem('refreshToken', refresh_token);

        await fetchProfile();

        elements.registerModal.style.display = 'none';
        elements.registerForm.reset();

        alert('Регистрация и вход выполнены успешно!');

    } catch (error) {
        console.error('Register error:', error);
        alert(error.message || 'Ошибка регистрации');
        logout();
    }
}

async function fetchProfile() {
    try {
        const token = localStorage.getItem('token');
        if (!token) return null;

        const response = await fetch(API.PROFILE, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok){
            if (response.status === 401) {
                logout();
            }
            throw new Error('Ошибка загрузки профиля');
        }

        const user = await response.json();
        state.currentUser = user;
        saveUserToStorage();
        updateUIForUser();

        if (user.role?.role_name === 'admin') {
            fetchUsers();
            fetchRoles();
        }
    } catch (error) {
        console.error('Error fetching profile:', error);
        logout();
    }
}

function logout() {
    fetch(API.LOGOUT, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    }).catch(err => console.error('Logout error:', err));

    state.currentUser = null;
    localStorage.removeItem('token');
    localStorage.removeItem('refreshToken');
    saveUserToStorage();
    updateUIForUser();
    renderBooks();

    elements.readerSection.classList.add('hidden');
    elements.bookListSection.classList.remove('hidden');
}

async function handleUpload(e) {
    e.preventDefault();

    const fileInput = elements.bookFile;
    const url = elements.bookUrl.value;
    const formData = new FormData();

    if (!fileInput.files.length && !url) {
        alert('Загрузите файл или укажите URL');
        return;
    }

    if (fileInput.files.length) {
        const file = fileInput.files[0];
        const allowedExtensions = ['.fb2', '.epub'];
        const extension = file.name.slice(file.name.lastIndexOf('.')).toLowerCase();

        if (!allowedExtensions.includes(extension)) {
            alert('Поддерживаются только файлы .fb2 и .epub');
            return;
        }
        formData.append('file', file);
    }

    else if (urlInput.value) {
        try {
            new URL(urlInput.value);
        } catch {
            alert('Укажите корректный URL');
            return;
        }
        formData.append('url', urlInput.value);
    }

    if (fileInput.files.length) {
        formData.append('file', fileInput.files[0]);
    } else {
        formData.append('url', url);
    }

    try {
        const token = localStorage.getItem('token');
        if (!token) throw new Error('Необходима авторизация');

        const submitBtn = elements.uploadForm.querySelector('.submit-btn');
        const originalBtnText = submitBtn.textContent;
        submitBtn.disabled = true;
        submitBtn.textContent = 'Загрузка...';

        const response = await fetch(API.BOOK_UPLOAD, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            },
            body: formData
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.message || 'Ошибка загрузки книги');
        }

        elements.uploadForm.reset();
        fetchBooks();

        alert('Книга успешно загружена!')
    } catch (error) {
        console.error('Upload error:', error);
        alert(error.message || 'Ошибка загрузки книги');
    }
}

async function fetchUsers() {
    try {
        const token = localStorage.getItem('token');
        if (!token) return;

        const response = await fetch(API.USER_LIST, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) throw new Error('Ошибка загрузки пользователей');

        const users = await response.json();
        state.users = users;
        renderUsers();
    } catch (error) {
        console.error('Error fetching users:', error);
    }
}

function renderUsers() {
    elements.userList.innerHTML = '';

    state.users.forEach(user => {
        const row = document.createElement('tr');

        row.innerHTML = `
            <td>${user.login}</td>
            <td>${user.role.role_name}</td>
            <td>
                <select class="select-role" data-user-id="${user.id}">
                    ${state.roles.map(role =>
            `<option value="${role.id}" ${role.id === user.role.id ? 'selected' : ''}>${role.role_name}</option>`
        ).join('')}
                </select>
                <button class="update-role-btn" data-user-id="${user.id}">Обновить</button>
                <button class="delete-user-btn" data-user-id="${user.id}" title="Удалить пользователя"></button>
            </td>
        `;

        elements.userList.appendChild(row);
    });

    document.querySelectorAll('.update-role-btn').forEach(btn => {
        btn.addEventListener('click', async (e) => {
            const userId = e.target.getAttribute('data-user-id');
            const select = document.querySelector(`.select-role[data-user-id="${userId}"]`);
            const roleId = select.value;

            try {
                const token = localStorage.getItem('token');
                if (!token) throw new Error('Необходима авторизация');

                const response = await fetch(API.ADD_ROLE, {
                    method: 'PUT',
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        user_id: parseInt(userId),
                        role_id: parseInt(roleId)
                    })
                });

                if (!response.ok) throw new Error('Ошибка обновления роли');

                alert('Роль успешно обновлена');
                fetchUsers();
            } catch (error) {
                console.error('Error updating role:', error);
                alert(error.message || 'Ошибка обновления роли');
            }
        });
    });

    document.querySelectorAll('.delete-user-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const userId = e.target.getAttribute('data-user-id');
            deleteItem('user', userId);
        });
    });
}

async function fetchRoles() {
    try {
        const token = localStorage.getItem('token');
        if (!token) return;

        const response = await fetch(API.ROLE_LIST, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) throw new Error('Ошибка загрузки ролей');

        const roles = await response.json();
        state.roles = roles;
        renderRoles();
    } catch (error) {
        console.error('Error fetching roles:', error);
    }
}

function renderRoles() {
    elements.roleList.innerHTML = '';

    state.roles.forEach(role => {
        const row = document.createElement('tr');

        row.innerHTML = `
                    <td>${role.role_name}</td>
                `;

        elements.roleList.appendChild(row);
    });
}

function canDeleteBook(book) {
    if (!state.currentUser) return false;

    if (state.currentUser.role?.role_name === 'admin'){
        return true;
    }

    return book.user_id === state.currentUser.id;
}

async function deleteItem(itemType, itemId) {
    let apiUrl, confirmMessage, successMessage, errorMessage, fetchAfterDelete;
    let item = null;

    if (itemType === 'book') {
        item = state.books.find(b => b.id == itemId);
        if (!item) return;

        if (!canDeleteBook(item)) {
            alert('У вас нет прав для удаления этой книги');
            return;
        }

        apiUrl = `${API.BOOK_DELETE}?id=${itemId}`;
        confirmMessage = `Вы уверены, что хотите удалить книгу "${item.title}"?`;
        successMessage = `Книга "${item.title}" успешно удалена`;
        errorMessage = 'Ошибка удаления книги';
        fetchAfterDelete = fetchBooks;
    }
    else if (itemType === 'user') {
        item = state.users.find(u => u.id == itemId);
        if (!item) return;

        apiUrl = `${API.USER_DELETE}?id=${itemId}`;
        confirmMessage = `Вы уверены, что хотите удалить пользователя "${item.login}"?`;
        successMessage = 'Пользователь успешно удалён';
        errorMessage = 'Ошибка удаления пользователя';
        fetchAfterDelete = fetchUsers;
    }
    else {
        console.error('Unknown item type for deletion:', itemType);
        return;
    }

    if (!confirm(confirmMessage)) return;

    try {
        const token = localStorage.getItem('token');
        if (!token) throw new Error('Необходима авторизация');

        const response = await fetch(apiUrl, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.message || errorMessage);
        }

        alert(successMessage);

        if (fetchAfterDelete) {
            await fetchAfterDelete();
        }

        if (itemType === 'book' && state.currentBook?.id == itemId) {
            toggleReaderMode(false);
            state.currentBook = null;
        }
    } catch (error) {
        console.error(`Delete ${itemType} error:`, error);
        alert(error.message || errorMessage);
    }
}

function updateReaderSettings() {
    state.readerSettings.fontFamily = elements.fontFamily.value;
    state.readerSettings.fontSize = elements.fontSize.value;
    if (elements.themeSelector) {
        state.readerSettings.theme = elements.themeSelector.value;
    }

    applyReaderSettings();
    applyReaderTheme();

    localStorage.setItem('readerSettings', JSON.stringify(state.readerSettings));
}

function applyReaderSettings() {
    elements.readerContent.style.fontFamily = state.readerSettings.fontFamily;
    elements.readerContent.style.fontSize = state.readerSettings.fontSize;

    elements.fontFamily.value = state.readerSettings.fontFamily;
    elements.fontSize.value = state.readerSettings.fontSize;

    if (elements.themeSelector) {
        elements.themeSelector.value = state.readerSettings.theme;
    }
}

function applyReaderTheme() {
    const readerContainer = document.querySelector('.reader-container');
    if (!readerContainer) return;

    readerContainer.classList.remove('reader-theme-light', 'reader-theme-dark', 'reader-theme-sepia', 'reader-theme-night');

    readerContainer.classList.add(`reader-theme-${state.readerSettings.theme}`);

    localStorage.setItem('readerTheme', state.readerSettings.theme);
}

function loadReaderTheme() {
    const saved = localStorage.getItem('readerSettings');
    if (saved) {
        const settings = JSON.parse(saved);
        state.readerSettings = {
            ...state.readerSettings,
            ...settings
        };
    }

    applyReaderTheme();
}

init();
