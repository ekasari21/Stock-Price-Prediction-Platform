import numpy as np
import pandas as pd
from datetime import timedelta
from sklearn.ensemble import RandomForestRegressor
from sklearn.linear_model import LinearRegression, Ridge
import xgboost as xgb
from ml.utils.metrics import calculate_metrics

FEATURE_COLS = ["open", "high", "low", "volume", "sma_10", "sma_20", "sma_50", "rsi_14", "macd", "lag_1", "lag_2", "lag_3"]

def train_and_predict_ml(df: pd.DataFrame, model_type: str = "xgboost", days_ahead: int = 7) -> dict:
    """
    Train a machine learning model (XGBoost or Scikit-Learn) and produce forecasts.
    """
    # Ensure features exist
    available_features = [col for col in FEATURE_COLS if col in df.columns]
    
    # Target is next-day close
    df = df.copy()
    df["target"] = df["close"].shift(-1)
    
    # Valid dataset for training/evaluation
    valid_df = df.dropna(subset=available_features + ["target"]).reset_index(drop=True)
    
    n_samples = len(valid_df)
    train_size = int(n_samples * 0.8)
    
    X_train = valid_df.loc[:train_size, available_features]
    y_train = valid_df.loc[:train_size, "target"]
    X_test = valid_df.loc[train_size:, available_features]
    y_test = valid_df.loc[train_size:, "target"]
    
    # Instantiate Model
    if model_type == "xgboost":
        model = xgb.XGBRegressor(
            n_estimators=120,
            learning_rate=0.04,
            max_depth=5,
            subsample=0.8,
            colsample_bytree=0.8,
            random_state=42,
            n_jobs=2
        )
        model_name = "XGBoost Regressor"
    elif model_type == "random_forest":
        model = RandomForestRegressor(
            n_estimators=100,
            max_depth=8,
            random_state=42,
            n_jobs=2
        )
        model_name = "Scikit-Learn Random Forest"
    elif model_type == "linear_regression":
        model = Ridge(alpha=1.0)
        model_name = "Scikit-Learn Ridge Regression"
    else:
        model = xgb.XGBRegressor(n_estimators=100, random_state=42)
        model_name = f"ML Model ({model_type})"
        
    model.fit(X_train, y_train)
    
    # Evaluation on test set
    y_pred_test = model.predict(X_test)
    metrics = calculate_metrics(y_test, y_pred_test)
    
    # Feature importance
    feature_importances = []
    if hasattr(model, "feature_importances_"):
        fi = model.feature_importances_
        sorted_idx = np.argsort(fi)[::-1]
        for idx in sorted_idx[:6]:
            feature_importances.append({
                "feature": available_features[idx],
                "importance": round(float(fi[idx]), 4)
            })
    elif hasattr(model, "coef_"):
        coefs = np.abs(model.coef_)
        sorted_idx = np.argsort(coefs)[::-1]
        for idx in sorted_idx[:6]:
            feature_importances.append({
                "feature": available_features[idx],
                "importance": round(float(coefs[idx]), 4)
            })
            
    # Autoregressive multi-day forecasting
    last_row = df.iloc[-1].copy()
    current_features = last_row[available_features].to_dict()
    current_close = float(last_row["close"])
    last_date = pd.to_datetime(last_row["date"])
    
    predictions = []
    curr_date = last_date
    recent_lags = [current_close, float(last_row.get("lag_1", current_close)), float(last_row.get("lag_2", current_close))]
    
    for day in range(1, days_ahead + 1):
        # Next business day
        curr_date += timedelta(days=1)
        while curr_date.weekday() >= 5: # Skip weekends
            curr_date += timedelta(days=1)
            
        # Build input vector
        feature_vector = pd.DataFrame([current_features])
        pred_val = float(model.predict(feature_vector)[0])
        
        # Prevent runaway divergence
        pred_val = max(pred_val, current_close * 0.7)
        pred_val = min(pred_val, current_close * 1.3)
        
        # Calculate bounds using test RMSE
        uncertainty = metrics["rmse"] * (1.0 + 0.15 * day)
        lower_bound = max(0.01, pred_val - uncertainty)
        upper_bound = pred_val + uncertainty
        
        predictions.append({
            "date": curr_date.strftime("%Y-%m-%d"),
            "predicted": round(pred_val, 2),
            "lower_bound": round(lower_bound, 2),
            "upper_bound": round(upper_bound, 2)
        })
        
        # Update state for next step
        recent_lags = [pred_val] + recent_lags[:2]
        if "lag_1" in current_features:
            current_features["lag_1"] = recent_lags[0]
        if "lag_2" in current_features:
            current_features["lag_2"] = recent_lags[1]
        if "lag_3" in current_features:
            current_features["lag_3"] = recent_lags[2]
        if "sma_10" in current_features:
            current_features["sma_10"] = (current_features["sma_10"] * 9 + pred_val) / 10
            
    return {
        "model_name": model_name,
        "model_type": model_type,
        "metrics": metrics,
        "predictions": predictions,
        "feature_importances": feature_importances
    }
