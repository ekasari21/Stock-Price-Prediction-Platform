import numpy as np
import pandas as pd
from datetime import datetime, timedelta
import socket
import urllib.request

def is_online(host="query1.finance.yahoo.com", port=443, timeout=1.5) -> bool:
    """Check if internet/yahoo finance is quickly reachable."""
    try:
        socket.setdefaulttimeout(timeout)
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.connect((host, port))
        s.close()
        return True
    except Exception:
        return False

def fetch_stock_data(symbol: str = "AAPL", period: str = "1y") -> pd.DataFrame:
    """
    Fetch historical stock data using yfinance, with deterministic realistic fallback.
    """
    df = None
    if is_online():
        try:
            import yfinance as yf
            ticker = yf.Ticker(symbol)
            history = ticker.history(period=period, timeout=3)
            if history is not None and not history.empty and len(history) >= 30:
                df = history.reset_index()
                # Standardize column names
                df.columns = [c.lower() for c in df.columns]
                if 'date' in df.columns:
                    df['date'] = pd.to_datetime(df['date']).dt.tz_localize(None)
                elif 'datetime' in df.columns:
                    df['date'] = pd.to_datetime(df['datetime']).dt.tz_localize(None)
        except Exception:
            df = None

    if df is None or df.empty or len(df) < 30:
        df = generate_synthetic_data(symbol)

    # Compute technical indicators
    df = add_technical_indicators(df)
    return df

def generate_synthetic_data(symbol: str, days: int = 250) -> pd.DataFrame:
    """
    Generate realistic geometric Brownian motion stock series as fallback.
    """
    np.random.seed(abs(hash(symbol)) % (2**31 - 1))
    
    base_prices = {
        "AAPL": 180.0,
        "MSFT": 415.0,
        "GOOGL": 175.0,
        "AMZN": 185.0,
        "TSLA": 220.0,
        "NVDA": 130.0,
        "BBCA.JK": 10200.0,
        "BBRI.JK": 4800.0,
    }
    start_price = base_prices.get(symbol.upper(), 100.0 + (abs(hash(symbol)) % 150))
    
    mu = 0.0006  # drift
    sigma = 0.018 # volatility
    
    end_date = datetime.now()
    dates = []
    curr = end_date - timedelta(days=int(days * 1.5))
    while len(dates) < days:
        if curr.weekday() < 5: # Monday to Friday
            dates.append(curr)
        curr += timedelta(days=1)
    
    returns = np.random.normal(mu, sigma, size=days)
    price_series = start_price * np.cumprod(1 + returns)
    
    # Create realistic Open, High, Low, Close, Volume
    data = []
    for i, date in enumerate(dates):
        close_p = price_series[i]
        daily_vol = np.random.uniform(0.005, 0.02) * close_p
        open_p = close_p + np.random.uniform(-0.5, 0.5) * daily_vol
        high_p = max(open_p, close_p) + np.random.uniform(0.1, 0.9) * daily_vol
        low_p = min(open_p, close_p) - np.random.uniform(0.1, 0.9) * daily_vol
        volume = int(np.random.uniform(15000000, 60000000))
        data.append({
            "date": pd.to_datetime(date.strftime("%Y-%m-%d")),
            "open": round(open_p, 2),
            "high": round(high_p, 2),
            "low": round(low_p, 2),
            "close": round(close_p, 2),
            "volume": volume
        })
    
    return pd.DataFrame(data)

def add_technical_indicators(df: pd.DataFrame) -> pd.DataFrame:
    """
    Calculate moving averages, RSI, MACD, and lag features.
    """
    df = df.copy()
    df.sort_values(by="date", inplace=True)
    df.reset_index(drop=True, inplace=True)
    
    # Simple Moving Averages
    df["sma_10"] = df["close"].rolling(window=10, min_periods=1).mean()
    df["sma_20"] = df["close"].rolling(window=20, min_periods=1).mean()
    df["sma_50"] = df["close"].rolling(window=50, min_periods=1).mean()
    
    # Exponential Moving Averages
    df["ema_12"] = df["close"].ewm(span=12, adjust=False).mean()
    df["ema_26"] = df["close"].ewm(span=26, adjust=False).mean()
    
    # MACD
    df["macd"] = df["ema_12"] - df["ema_26"]
    df["macd_signal"] = df["macd"].ewm(span=9, adjust=False).mean()
    
    # RSI (Relative Strength Index)
    delta = df["close"].diff()
    gain = (delta.where(delta > 0, 0)).rolling(window=14, min_periods=1).mean()
    loss = (-delta.where(delta < 0, 0)).rolling(window=14, min_periods=1).mean()
    rs = gain / (loss.replace(0, 0.0001))
    df["rsi_14"] = 100 - (100 / (1 + rs))
    
    # Lags
    for lag in [1, 2, 3, 5, 10]:
        df[f"lag_{lag}"] = df["close"].shift(lag)
        
    df.bfill(inplace=True)
    df.ffill(inplace=True)
    return df
