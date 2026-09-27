#!/usr/bin/env python3
"""
Sample local FastAPI server to receive fasting sync data from the Garmin Instinct 3 AMOLED.
"""

import sys
import json
import sqlite3
from datetime import datetime
from typing import List, Optional
from pydantic import BaseModel

try:
    from fastapi import FastAPI, Header, HTTPException, status
    import uvicorn
except ImportError:
    print("FastAPI / Uvicorn not installed. Install with: pip install fastapi uvicorn")
    sys.exit(1)

app = FastAPI(title="Garmin Fasting Local Sync Server")

DB_FILE = "fasting_records.db"

def init_db():
    conn = sqlite3.connect(DB_FILE)
    c = conn.cursor()
    c.execute("""
        CREATE TABLE IF NOT EXISTS fasts (
            id INTEGER PRIMARY KEY,
            start_epoch INTEGER,
            end_epoch INTEGER,
            duration_hours REAL,
            target_hours INTEGER,
            state TEXT,
            synced_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    """)
    conn.commit()
    conn.close()

init_db()

class FastRecord(BaseModel):
    id: int
    start: int
    end: int
    duration_hours: float
    target_hours: int
    state: str

class SyncPayload(BaseModel):
    device: Optional[str] = "Garmin Instinct 3 AMOLED 50mm"
    synced_at: Optional[int] = None
    fasts: List[FastRecord]

EXPECTED_TOKEN = "your_secret_token_here"

@app.post("/api/fasting/sync")
async def receive_sync(payload: SyncPayload, authorization: Optional[str] = Header(None)):
    if EXPECTED_TOKEN and EXPECTED_TOKEN != "your_secret_token_here":
        if authorization != f"Bearer {EXPECTED_TOKEN}":
            raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token")

    conn = sqlite3.connect(DB_FILE)
    c = conn.cursor()
    inserted = 0

    for fast in payload.fasts:
        c.execute("""
            INSERT OR REPLACE INTO fasts (id, start_epoch, end_epoch, duration_hours, target_hours, state)
            VALUES (?, ?, ?, ?, ?, ?)
        """, (fast.id, fast.start, fast.end, fast.duration_hours, fast.target_hours, fast.state))
        inserted += 1

    conn.commit()
    conn.close()

    print(f"[{datetime.now().isoformat()}] Received {inserted} fasts from {payload.device}")
    return {"status": "success", "received_count": inserted}

@app.get("/api/fasting/history")
async def get_history():
    conn = sqlite3.connect(DB_FILE)
    conn.row_factory = sqlite3.Row
    c = conn.cursor()
    c.execute("SELECT * FROM fasts ORDER BY start_epoch DESC LIMIT 30")
    rows = [dict(r) for r in c.fetchall()]
    conn.close()
    return {"records": rows}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
