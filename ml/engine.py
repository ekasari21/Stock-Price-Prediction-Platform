import sys
import json
import argparse
from datetime import datetime

from ml.utils.data_loader import fetch_stock_data

def main():
    parser = argparse.ArgumentParser(description="Stockwise AI ML Inference Engine")
    parser.add_argument("--symbol", type=str, default="AAPL", help="Stock ticker symbol (e.g. AAPL, MSFT, BBCA.JK)")
    parser.add_argument("--model", type=str, default="xgboost", choices=["xgboost", "random_forest", "linear_regression", "lstm", "arima"], help="Model type")
    parser.add_argument("--days", type=int, default=7, help="Number of future days to forecast")
    parser.add_argument("--period", type=str, default="1y", help="Historical data period")
    parser.add_argument("--output", type=str, default="", help="Optional output file path")
    args = parser.parse_args()

    symbol = args.symbol.upper().strip()
    model_type = args.model.lower().strip()
    days_ahead = max(1, min(args.days, 30))

    try:
        # 1. Fetch & preprocess data
        df = fetch_stock_data(symbol=symbol, period=args.period)
        
        last_row = df.iloc[-1]
        prev_row = df.iloc[-2] if len(df) > 1 else last_row
        last_price = round(float(last_row["close"]), 2)
        prev_price = round(float(prev_row["close"]), 2)
        price_change = round(last_price - prev_price, 2)
        pct_change = round((price_change / prev_price) * 100, 2) if prev_price > 0 else 0.0

        currency = "IDR" if ".JK" in symbol else "USD"

        # 2. Train and generate predictions based on model type (lazy loaded for speed)
        if model_type in ["xgboost", "random_forest", "linear_regression"]:
            from ml.models.ml_models import train_and_predict_ml
            res = train_and_predict_ml(df, model_type=model_type, days_ahead=days_ahead)
        elif model_type == "lstm":
            from ml.models.dl_model import train_and_predict_dl
            res = train_and_predict_dl(df, days_ahead=days_ahead)
        elif model_type == "arima":
            from ml.models.stats_model import train_and_predict_stats
            res = train_and_predict_stats(df, days_ahead=days_ahead)
        else:
            from ml.models.ml_models import train_and_predict_ml
            res = train_and_predict_ml(df, model_type="xgboost", days_ahead=days_ahead)

        # 3. Format recent historical data (up to last 60 days) for Chart.js
        history_df = df.tail(60).copy()
        history_records = []
        for _, row in history_df.iterrows():
            date_str = str(row["date"])[:10]
            history_records.append({
                "date": date_str,
                "open": round(float(row["open"]), 2),
                "high": round(float(row["high"]), 2),
                "low": round(float(row["low"]), 2),
                "close": round(float(row["close"]), 2),
                "volume": int(row["volume"]),
                "sma_10": round(float(row.get("sma_10", row["close"])), 2),
                "sma_20": round(float(row.get("sma_20", row["close"])), 2)
            })

        output_payload = {
            "success": True,
            "symbol": symbol,
            "currency": currency,
            "last_price": last_price,
            "price_change": price_change,
            "pct_change": pct_change,
            "model_name": res["model_name"],
            "model_type": res["model_type"],
            "metrics": res["metrics"],
            "predictions": res["predictions"],
            "feature_importances": res.get("feature_importances", []),
            "history": history_records,
            "generated_at": datetime.now().isoformat()
        }

        output_json = json.dumps(output_payload, indent=2)

        if args.output:
            with open(args.output, "w") as f:
                f.write(output_json)
        else:
            print(output_json)

    except Exception as e:
        error_payload = {
            "success": False,
            "error": str(e),
            "symbol": symbol,
            "model_type": model_type
        }
        print(json.dumps(error_payload))
        sys.exit(1)

if __name__ == "__main__":
    main()
