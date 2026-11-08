/**
 * Authentication management for Stockwise AI
 */
const Auth = {
    currentUser: null,

    init() {
        const token = localStorage.getItem('access_token');
        const userStr = localStorage.getItem('user_info');
        if (token && userStr) {
            try {
                this.currentUser = JSON.parse(userStr);
            } catch (e) {
                this.currentUser = null;
            }
        }
        this.updateUI();
    },

    isLoggedIn() {
        return !!localStorage.getItem('access_token');
    },

    async login(email, password) {
        try {
            const res = await API.login(email, password);
            if (res.success && res.data) {
                localStorage.setItem('access_token', res.data.access_token);
                localStorage.setItem('refresh_token', res.data.refresh_token);
                localStorage.setItem('user_info', JSON.stringify(res.data.user));
                this.currentUser = res.data.user;
                this.updateUI();
                return { success: true };
            }
            return { success: false, message: res.error || 'Login failed' };
        } catch (err) {
            return { success: false, message: err.message };
        }
    },

    async register(username, email, password) {
        try {
            const res = await API.register(username, email, password);
            if (res.success && res.data) {
                localStorage.setItem('access_token', res.data.access_token);
                localStorage.setItem('refresh_token', res.data.refresh_token);
                localStorage.setItem('user_info', JSON.stringify(res.data.user));
                this.currentUser = res.data.user;
                this.updateUI();
                return { success: true };
            }
            return { success: false, message: res.error || 'Registration failed' };
        } catch (err) {
            return { success: false, message: err.message };
        }
    },

    logout() {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('user_info');
        this.currentUser = null;
        this.updateUI();
    },

    updateUI() {
        const guestMenu = document.getElementById('guest-menu');
        const userMenu = document.getElementById('user-menu');
        const userDisplayName = document.getElementById('user-display-name');
        const authWarning = document.getElementById('auth-warning');

        if (this.isLoggedIn() && this.currentUser) {
            if (guestMenu) guestMenu.classList.add('d-none');
            if (userMenu) userMenu.classList.remove('d-none');
            if (userDisplayName) userDisplayName.textContent = this.currentUser.username;
            if (authWarning) authWarning.classList.add('d-none');
        } else {
            if (guestMenu) guestMenu.classList.remove('d-none');
            if (userMenu) userMenu.classList.add('d-none');
            if (authWarning) authWarning.classList.remove('d-none');
        }
    }
};
