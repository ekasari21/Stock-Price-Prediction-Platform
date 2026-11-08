/**
 * AJAX API Client for Stockwise AI
 */
const API = {
    baseUrl: '/api/v1',

    getHeaders(withAuth = true) {
        const headers = {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        };
        if (withAuth) {
            const token = localStorage.getItem('access_token');
            if (token) {
                headers['Authorization'] = `Bearer ${token}`;
            }
        }
        return headers;
    },

    async request(endpoint, options = {}, withAuth = true) {
        const url = `${this.baseUrl}${endpoint}`;
        const config = {
            ...options,
            headers: {
                ...this.getHeaders(withAuth),
                ...(options.headers || {})
            }
        };

        try {
            const response = await fetch(url, config);
            
            // Check for unauthorized / expired token
            if (response.status === 401 && withAuth && !endpoint.includes('/auth/login') && !endpoint.includes('/auth/refresh')) {
                const refreshed = await this.refreshToken();
                if (refreshed) {
                    // Retry with new token
                    config.headers['Authorization'] = `Bearer ${localStorage.getItem('access_token')}`;
                    const retryResponse = await fetch(url, config);
                    return await this.handleResponse(retryResponse);
                } else {
                    // Log out if refresh failed
                    Auth.logout();
                    throw new Error('Session expired. Please log in again.');
                }
            }

            return await this.handleResponse(response);
        } catch (error) {
            console.error(`API Error [${endpoint}]:`, error);
            throw error;
        }
    },

    async handleResponse(response) {
        const data = await response.json().catch(() => null);
        if (!response.ok) {
            const errorMsg = (data && (data.error || data.message)) || `HTTP ${response.status}: ${response.statusText}`;
            throw new Error(errorMsg);
        }
        return data;
    },

    async refreshToken() {
        const refreshToken = localStorage.getItem('refresh_token');
        if (!refreshToken) return false;

        try {
            const res = await fetch(`${this.baseUrl}/auth/refresh`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ refresh_token: refreshToken })
            });
            const data = await res.json();
            if (res.ok && data.success && data.data.access_token) {
                localStorage.setItem('access_token', data.data.access_token);
                if (data.data.refresh_token) {
                    localStorage.setItem('refresh_token', data.data.refresh_token);
                }
                return true;
            }
        } catch (e) {
            console.error('Refresh token request failed', e);
        }
        return false;
    },

    // Auth endpoints
    login(email, password) {
        return this.request('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ email, password })
        }, false);
    },

    register(username, email, password) {
        return this.request('/auth/register', {
            method: 'POST',
            body: JSON.stringify({ username, email, password })
        }, false);
    },

    getProfile() {
        return this.request('/auth/me', { method: 'GET' }, true);
    },

    // Stock endpoints
    getStocks() {
        return this.request('/stocks', { method: 'GET' }, false);
    },

    getStockDetail(symbol) {
        return this.request(`/stocks/${symbol}`, { method: 'GET' }, false);
    },

    // Prediction endpoints
    runPrediction(payload) {
        return this.request('/predictions/forecast', {
            method: 'POST',
            body: JSON.stringify(payload)
        }, true);
    },

    getPredictionHistory(limit = 10) {
        return this.request(`/predictions/history?limit=${limit}`, { method: 'GET' }, true);
    },

    getPredictionById(id) {
        return this.request(`/predictions/${id}`, { method: 'GET' }, true);
    }
};
