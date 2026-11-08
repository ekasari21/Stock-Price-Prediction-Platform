/**
 * Main Application Controller for Stockwise AI
 */
document.addEventListener('DOMContentLoaded', () => {
    // 1. Initialize Auth and Charts
    Auth.init();
    StockCharts.init();

    // 2. Load Initial Data
    loadStocks();
    if (Auth.isLoggedIn()) {
        loadPredictionHistory();
    }

    // 3. Setup Event Listeners
    setupEventListeners();

    // 4. Run an initial forecast for demo
    runInitialForecast();
});

function setupEventListeners() {
    // Prediction Form Submit
    const predictionForm = document.getElementById('predictionForm');
    if (predictionForm) {
        predictionForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            await executePrediction();
        });
    }

    // Stock selection change
    const stockSelect = document.getElementById('stockSelect');
    if (stockSelect) {
        stockSelect.addEventListener('change', (e) => {
            const customInput = document.getElementById('customSymbolInput');
            if (e.target.value === 'CUSTOM') {
                customInput.classList.remove('d-none');
                customInput.focus();
            } else {
                customInput.classList.add('d-none');
            }
        });
    }

    // Login Form Submit
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const email = document.getElementById('loginEmail').value;
            const password = document.getElementById('loginPassword').value;
            const alertEl = document.getElementById('loginAlert');

            const res = await Auth.login(email, password);
            if (res.success) {
                bootstrap.Modal.getInstance(document.getElementById('loginModal')).hide();
                loadPredictionHistory();
                showToast('Welcome back! Logged in successfully.', 'success');
            } else {
                alertEl.textContent = res.message;
                alertEl.classList.remove('d-none');
            }
        });
    }

    // Register Form Submit
    const registerForm = document.getElementById('registerForm');
    if (registerForm) {
        registerForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const username = document.getElementById('registerUsername').value;
            const email = document.getElementById('registerEmail').value;
            const password = document.getElementById('registerPassword').value;
            const alertEl = document.getElementById('registerAlert');

            const res = await Auth.register(username, email, password);
            if (res.success) {
                bootstrap.Modal.getInstance(document.getElementById('registerModal')).hide();
                loadPredictionHistory();
                showToast('Account created successfully! You are now logged in.', 'success');
            } else {
                alertEl.textContent = res.message;
                alertEl.classList.remove('d-none');
            }
        });
    }

    // Logout Button
    const logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', (e) => {
            e.preventDefault();
            Auth.logout();
            showToast('Logged out successfully.', 'info');
            document.getElementById('historyTableBody').innerHTML = '<tr><td colspan="6" class="text-center text-muted">Please log in to view prediction history.</td></tr>';
        });
    }
}

async function loadStocks() {
    try {
        const res = await API.getStocks();
        if (res.success && res.data) {
            const stockSelect = document.getElementById('stockSelect');
            const stocksListContainer = document.getElementById('stocksOverview');
            
            // Clear current options except custom
            stockSelect.innerHTML = '';
            
            res.data.forEach(stock => {
                const opt = document.createElement('option');
                opt.value = stock.symbol;
                opt.textContent = `${stock.symbol} - ${stock.name}`;
                stockSelect.appendChild(opt);
            });

            // Add custom option
            const customOpt = document.createElement('option');
            customOpt.value = 'CUSTOM';
            customOpt.textContent = '+ Custom Ticker...';
            stockSelect.appendChild(customOpt);

            // Render Overview Cards
            if (stocksListContainer) {
                stocksListContainer.innerHTML = res.data.slice(0, 4).map(s => {
                    const isPositive = s.price_change >= 0;
                    const changeColor = isPositive ? 'text-success' : 'text-danger';
                    const changeSign = isPositive ? '+' : '';
                    return `
                        <div class="col-md-3 col-6 mb-3">
                            <div class="metric-card cursor-pointer" onclick="selectStock('${s.symbol}')">
                                <div class="d-flex justify-content-between align-items-center mb-1">
                                    <span class="fw-bold">${s.symbol}</span>
                                    <span class="badge bg-secondary" style="font-size:0.7rem">${s.currency}</span>
                                </div>
                                <div class="metric-value mb-1">${s.last_price.toLocaleString()}</div>
                                <div class="${changeColor}" style="font-size: 0.8rem;">
                                    ${changeSign}${s.price_change.toFixed(2)} (${changeSign}${s.pct_change.toFixed(2)}%)
                                </div>
                            </div>
                        </div>
                    `;
                }).join('');
            }
        }
    } catch (err) {
        console.error('Failed to load stocks:', err);
    }
}

function selectStock(symbol) {
    const select = document.getElementById('stockSelect');
    if (select) {
        select.value = symbol;
        const customInput = document.getElementById('customSymbolInput');
        if (customInput) customInput.classList.add('d-none');
    }
}

async function executePrediction() {
    const stockSelect = document.getElementById('stockSelect');
    let symbol = stockSelect.value;
    if (symbol === 'CUSTOM') {
        symbol = document.getElementById('customSymbolInput').value.trim();
        if (!symbol) {
            showToast('Please enter a valid stock symbol', 'warning');
            return;
        }
    }

    const modelType = document.getElementById('modelSelect').value;
    const days = parseInt(document.getElementById('daysSelect').value, 10);
    const period = document.getElementById('periodSelect').value;

    const overlay = document.getElementById('loadingOverlay');
    overlay.classList.add('active');

    try {
        const payload = {
            symbol: symbol,
            model_type: modelType,
            days: days,
            period: period
        };

        const res = await API.runPrediction(payload);
        if (res.success && res.data) {
            renderPredictionResults(res.data);
            showToast(`Forecast complete using ${res.data.model_name}`, 'success');
            if (Auth.isLoggedIn()) {
                loadPredictionHistory();
            }
        } else {
            showToast(res.error || 'Prediction failed', 'danger');
        }
    } catch (err) {
        showToast(err.message || 'Forecast error occurred', 'danger');
    } finally {
        overlay.classList.remove('active');
    }
}

function renderPredictionResults(data) {
    // 1. Update Title and Badges
    document.getElementById('chartStockTitle').textContent = `${data.symbol} Stock Price Forecast`;
    const modelBadge = document.getElementById('chartModelBadge');
    modelBadge.textContent = data.model_name;
    modelBadge.className = `badge badge-model badge-${data.model_type}`;

    // 2. Update Header Price
    document.getElementById('currentStockPrice').textContent = `${data.currency} ${data.last_price.toLocaleString(undefined, {minimumFractionDigits: 2})}`;
    const changeEl = document.getElementById('currentStockChange');
    const isPos = data.price_change >= 0;
    changeEl.className = isPos ? 'text-success fw-bold' : 'text-danger fw-bold';
    changeEl.textContent = `${isPos ? '+' : ''}${data.price_change.toFixed(2)} (${isPos ? '+' : ''}${data.pct_change.toFixed(2)}%)`;

    // 3. Update Evaluation Metrics
    document.getElementById('metricMAE').textContent = data.metrics.mae.toFixed(4);
    document.getElementById('metricRMSE').textContent = data.metrics.rmse.toFixed(4);
    document.getElementById('metricMAPE').textContent = `${data.metrics.mape.toFixed(2)}%`;
    document.getElementById('metricR2').textContent = data.metrics.r2_score.toFixed(4);

    // 4. Render Main Multi-series Chart
    StockCharts.renderPredictionChart(data.history, data.predictions, data.symbol, data.currency);

    // 5. Render Feature Importance
    StockCharts.renderFeatureImportanceChart(data.feature_importances);

    // 6. Update Forecast Summary Table
    const tableBody = document.getElementById('forecastSummaryBody');
    if (tableBody) {
        tableBody.innerHTML = data.predictions.map(p => `
            <tr>
                <td><strong>${p.date}</strong></td>
                <td class="text-primary fw-bold">${data.currency} ${p.predicted.toFixed(2)}</td>
                <td class="text-muted">${data.currency} ${p.lower_bound.toFixed(2)}</td>
                <td class="text-muted">${data.currency} ${p.upper_bound.toFixed(2)}</td>
            </tr>
        `).join('');
    }
}

async function loadPredictionHistory() {
    try {
        const res = await API.getPredictionHistory(10);
        const tableBody = document.getElementById('historyTableBody');
        if (!tableBody) return;

        if (res.success && res.data && res.data.length > 0) {
            tableBody.innerHTML = res.data.map(item => `
                <tr>
                    <td><strong>${item.symbol}</strong></td>
                    <td><span class="badge badge-model badge-${item.model_type}">${item.model_name}</span></td>
                    <td>${item.currency} ${item.last_price.toFixed(2)}</td>
                    <td><span class="badge bg-dark border">${item.mae.toFixed(2)}</span></td>
                    <td><span class="badge bg-dark border">${item.r2_score.toFixed(2)}</span></td>
                    <td><small class="text-muted">${new Date(item.created_at).toLocaleDateString()}</small></td>
                </tr>
            `).join('');
        } else {
            tableBody.innerHTML = '<tr><td colspan="6" class="text-center text-muted">No past predictions yet. Run your first forecast above!</td></tr>';
        }
    } catch (err) {
        console.error('Failed to load prediction history:', err);
    }
}

async function runInitialForecast() {
    // Run an initial quick demo forecast on load
    const overlay = document.getElementById('loadingOverlay');
    overlay.classList.add('active');
    try {
        const res = await fetch('/api/v1/stocks/AAPL');
        const data = await res.json();
        // Run forecast
        const predRes = await fetch('/api/v1/predictions/forecast', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ symbol: 'AAPL', model_type: 'xgboost', days: 7, period: '1y' })
        });
        const predData = await predRes.json();
        if (predData.success && predData.data) {
            renderPredictionResults(predData.data);
        }
    } catch (e) {
        console.log('Initial forecast info:', e);
    } finally {
        overlay.classList.remove('active');
    }
}

function showToast(message, type = 'info') {
    const toastContainer = document.getElementById('toastContainer');
    if (!toastContainer) return;

    const toastEl = document.createElement('div');
    toastEl.className = `toast align-items-center text-white bg-${type === 'info' ? 'primary' : type} border-0 mb-2`;
    toastEl.setAttribute('role', 'alert');
    toastEl.setAttribute('aria-live', 'assertive');
    toastEl.setAttribute('aria-atomic', 'true');
    toastEl.innerHTML = `
        <div class="d-flex">
            <div class="toast-body">${message}</div>
            <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
        </div>
    `;

    toastContainer.appendChild(toastEl);
    const toast = new bootstrap.Toast(toastEl, { delay: 4000 });
    toast.show();
    toastEl.addEventListener('hidden.bs.toast', () => toastEl.remove());
}
