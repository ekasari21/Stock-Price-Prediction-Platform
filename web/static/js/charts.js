/**
 * Chart.js Visualizations for Stockwise AI
 */
const StockCharts = {
    mainChart: null,
    featureChart: null,

    init() {
        // Configure Chart.js global dark theme defaults
        Chart.defaults.color = '#94a3b8';
        Chart.defaults.borderColor = '#334155';
    },

    renderPredictionChart(historyData, predictionsData, symbol, currency = 'USD') {
        const ctx = document.getElementById('priceChart').getContext('2d');

        if (this.mainChart) {
            this.mainChart.destroy();
        }

        const labels = [];
        const historicalClose = [];
        const sma10 = [];
        const sma20 = [];
        const predictedSeries = [];
        const lowerBoundSeries = [];
        const upperBoundSeries = [];

        // 1. Process historical data
        historyData.forEach(item => {
            labels.push(item.date);
            historicalClose.push(item.close);
            sma10.push(item.sma_10);
            sma20.push(item.sma_20);
            predictedSeries.push(null);
            lowerBoundSeries.push(null);
            upperBoundSeries.push(null);
        });

        // Bridge between history and prediction
        if (historyData.length > 0 && predictionsData.length > 0) {
            const lastHist = historyData[historyData.length - 1];
            // Bridge point
            predictedSeries[predictedSeries.length - 1] = lastHist.close;
            lowerBoundSeries[lowerBoundSeries.length - 1] = lastHist.close;
            upperBoundSeries[upperBoundSeries.length - 1] = lastHist.close;
        }

        // 2. Process predicted data
        predictionsData.forEach(item => {
            labels.push(item.date);
            historicalClose.push(null);
            sma10.push(null);
            sma20.push(null);
            predictedSeries.push(item.predicted);
            lowerBoundSeries.push(item.lower_bound);
            upperBoundSeries.push(item.upper_bound);
        });

        this.mainChart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: labels,
                datasets: [
                    {
                        label: 'Historical Close (' + symbol + ')',
                        data: historicalClose,
                        borderColor: '#3b82f6',
                        backgroundColor: 'rgba(59, 130, 246, 0.1)',
                        borderWidth: 2,
                        pointRadius: 1.5,
                        tension: 0.1,
                        fill: false
                    },
                    {
                        label: 'SMA (10)',
                        data: sma10,
                        borderColor: 'rgba(16, 185, 129, 0.7)',
                        borderWidth: 1.5,
                        borderDash: [5, 5],
                        pointRadius: 0,
                        fill: false
                    },
                    {
                        label: 'SMA (20)',
                        data: sma20,
                        borderColor: 'rgba(245, 158, 11, 0.7)',
                        borderWidth: 1.5,
                        borderDash: [5, 5],
                        pointRadius: 0,
                        fill: false
                    },
                    {
                        label: 'AI Forecast',
                        data: predictedSeries,
                        borderColor: '#a855f7',
                        backgroundColor: '#a855f7',
                        borderWidth: 3,
                        pointRadius: 5,
                        pointHoverRadius: 7,
                        tension: 0.2,
                        fill: false
                    },
                    {
                        label: 'Upper Bound (90% Conf.)',
                        data: upperBoundSeries,
                        borderColor: 'rgba(168, 85, 247, 0.25)',
                        borderWidth: 1,
                        borderDash: [3, 3],
                        pointRadius: 0,
                        fill: '+1',
                        backgroundColor: 'rgba(168, 85, 247, 0.12)'
                    },
                    {
                        label: 'Lower Bound (90% Conf.)',
                        data: lowerBoundSeries,
                        borderColor: 'rgba(168, 85, 247, 0.25)',
                        borderWidth: 1,
                        borderDash: [3, 3],
                        pointRadius: 0,
                        fill: false
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                interaction: {
                    mode: 'index',
                    intersect: false
                },
                plugins: {
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                let label = context.dataset.label || '';
                                if (label) {
                                    label += ': ';
                                }
                                if (context.parsed.y !== null) {
                                    label += currency + ' ' + context.parsed.y.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
                                }
                                return label;
                            }
                        }
                    },
                    legend: {
                        position: 'top',
                        labels: {
                            boxWidth: 12,
                            padding: 15
                        }
                    }
                },
                scales: {
                    x: {
                        grid: { color: 'rgba(51, 65, 85, 0.4)' },
                        ticks: { maxTicksLimit: 12 }
                    },
                    y: {
                        grid: { color: 'rgba(51, 65, 85, 0.4)' },
                        ticks: {
                            callback: function(value) {
                                return currency + ' ' + value.toLocaleString();
                            }
                        }
                    }
                }
            }
        });
    },

    renderFeatureImportanceChart(features) {
        const container = document.getElementById('featureChartContainer');
        const ctx = document.getElementById('featureChart');
        if (!ctx) return;

        if (!features || features.length === 0) {
            if (container) container.classList.add('d-none');
            return;
        }

        if (container) container.classList.remove('d-none');

        if (this.featureChart) {
            this.featureChart.destroy();
        }

        const labels = features.map(f => f.feature);
        const data = features.map(f => f.importance);

        this.featureChart = new Chart(ctx.getContext('2d'), {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [{
                    label: 'Relative Weight / Importance',
                    data: data,
                    backgroundColor: 'rgba(59, 130, 246, 0.8)',
                    borderColor: '#3b82f6',
                    borderWidth: 1,
                    borderRadius: 4
                }]
            },
            options: {
                indexAxis: 'y',
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false }
                },
                scales: {
                    x: {
                        beginAtZero: true,
                        grid: { color: 'rgba(51, 65, 85, 0.4)' }
                    },
                    y: {
                        grid: { display: false }
                    }
                }
            }
        });
    }
};
